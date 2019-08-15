package release

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/version"
	"net/http"
	"os"
	"sort"
	"strings"
)

func ListSupportedProviders() []string {
	var results []string
	for index := 1; ; index++ {
		page := loopProviders(index)
		results = append(results, page...)
		if len(page) == 0 {
			break
		}
	}
	if len(results) <= 100 {
		panic("Did not get all providers. Maybe api call may have been throttled.")
	}
	sort.Strings(results)
	return results
}

func WriteNewVersion() {
	latestVersion := getLatestVersion("jeshan/tfbridge")
	semantic, _ := version.ParseGeneric(latestVersion)
	oldVersion := semantic.String()
	semantic = semantic.WithMinor(semantic.Minor() + 1)

	newVersion := fmt.Sprintf("v%s", semantic.String())
	fmt.Println("Incrementing release from", oldVersion, "to", newVersion)

	file, _ := os.Create(".version")
	file.WriteString(newVersion)
}

func CreateRelease() {
	newVersion := getNewVersion()
	terraformVersion := getTerraformVersion()
	requestBody, _ := json.Marshal(map[string]interface{}{
		"tag_name": newVersion,
		"body":     createReleaseNotes(newVersion, terraformVersion, os.Getenv("BUCKET"), readProviderInfo()),
	})
	req, _ := http.NewRequest("POST", "https://api.github.com/repos/jeshan/tfbridge/releases", bytes.NewBuffer(requestBody))
	req.Header.Set("Authorization", fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN")))
	client := &http.Client{}
	createResponse, err := client.Do(req)
	fmt.Println(createResponse)
	if err != nil {
		panic(err)
	}
}

func getTerraformVersion() string {
	file, err := ioutil.ReadFile("go.mod")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		if strings.Index(line, "github.com/hashicorp/terraform") >= 0 {
			split := strings.Split(line, " ")
			tfVersion := split[1]
			return tfVersion
		}
	}
	panic("Could not determine Terraform version being used")
}

func getNewVersion() string {
	contents, _ := ioutil.ReadFile(".version")
	return string(contents)
}

func readProviderVersion(providerName string) string {
	file, err := ioutil.ReadFile("download-dependencies.sh")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		if strings.Index(line, fmt.Sprintf("github.com/terraform-providers/terraform-provider-%s", providerName)) >= 0 {
			split := strings.Split(line, "@")
			providerVersion := split[1]
			return providerVersion
		}
	}
	panic(fmt.Sprintf("Could not get version for provider %s", providerName))
}

func readProviderInfo() []ProviderInfo {
	var result []ProviderInfo
	infos, err := ioutil.ReadDir("dist")
	if err != nil {
		panic(err)
	}
	for _, item := range infos {
		if strings.HasSuffix(item.Name(), ".zip") {
			fmt.Println(item.Name())
			providerName := item.Name()[:len(item.Name())-4]
			result = append(result, ProviderInfo{
				Name:    providerName,
				Version: readProviderVersion(providerName),
			})
		}
	}
	return result
}

type ProviderInfo struct {
	Name    string
	Version string
}

func createReleaseNotes(projectVersion string, terraformVersion string, bucket string, providers []ProviderInfo) string {
	files, e := template.ParseFiles("release-note.gohtml")
	if e != nil {
		panic(e)
	}
	var writer bytes.Buffer
	e = files.Templates()[0].Execute(&writer, releaseNoteData{TfBridgeVersion: projectVersion, TerraformVersion: terraformVersion, Providers: providers, Bucket: bucket})
	if e != nil {
		panic(e)
	}
	return writer.String()
}

type releaseNoteData struct {
	TfBridgeVersion  string
	TerraformVersion string
	Providers        []ProviderInfo
	Bucket           string
}

type cfnTemplateData struct {
	ProviderName      string
	ProviderNameTitle string
}

func WriteProviderFiles() {
	supportedProviders := ListSupportedProviders()
	for _, value := range supportedProviders {
		writeProviderFile(value)
	}
	writeDownloadDependenciesScript(supportedProviders)
	writeCfnTemplates(supportedProviders)
}

func writeCfnTemplates(supportedProviders []string) {
	for _, providerName := range supportedProviders {
		os.MkdirAll("tfbridge/providers", 0700)
		path := fmt.Sprintf("%s-cfn-template.yaml", providerName)
		if _, e := os.Stat(path); e == nil {
			return
		}
		files, e := template.ParseFiles("provider-cfn.gohtml")
		if e != nil {
			panic(e)
		}
		fmt.Println("Creating file", path)
		file, e := os.Create(path)
		if e != nil {
			panic(e)
		}
		writer := bufio.NewWriter(file)
		e = files.Templates()[0].Execute(writer, cfnTemplateData{ProviderName: providerName, ProviderNameTitle: strings.Title(providerName)})
		if e != nil {
			panic(e)
		}
		defer func() {
			if e := writer.Flush(); e != nil {
				panic(e)
			}
		}()
	}
}

//noinspection GoUnhandledErrorResult
func writeDownloadDependenciesScript(supportedProviders []string) {
	path := "download-dependencies.sh"
	fmt.Println("Creating file", path)
	file, _ := os.Create(path)
	file.WriteString("#!/usr/bin/env bash\n")
	for _, value := range supportedProviders {
		latest := getLatestVersion(fmt.Sprintf("terraform-providers/terraform-provider-%s", value))
		file.WriteString(fmt.Sprintf("go get -d github.com/terraform-providers/terraform-provider-%s@%s\n", value, latest))
	}
	file.WriteString("exit 0\n")
	file.Close()
	os.Chmod(path, 0700)
}

func getLatestVersion(projectName string) string {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", projectName), nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN")))
	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println("Getting latest version for", projectName)
	if err != nil {
		panic(err)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var parsed map[string]string
	_ = json.Unmarshal(content, &parsed)
	tagName := parsed["tag_name"]
	if len(tagName) == 0 {
		return "latest"
	}
	return tagName
}

//noinspection GoUnhandledErrorResult
func writeProviderFile(providerName string) {
	os.MkdirAll("tfbridge/providers", 0700)
	path := fmt.Sprintf("tfbridge/providers/%s.go", providerName)
	if _, e := os.Stat(path); e == nil {
		return
	}

	files, e := template.ParseFiles("default-provider-file.gohtml")
	if e != nil {
		panic(e)
	}
	fmt.Println("Creating file", path)
	file, e := os.Create(path)
	if e != nil {
		panic(e)
	}
	writer := bufio.NewWriter(file)
	e = files.Templates()[0].Execute(writer, cfnTemplateData{ProviderName: providerName})
	if e != nil {
		panic(e)
	}
	os.Rename(fmt.Sprintf("tfbridge/providers/%s-custom.go", providerName),
		fmt.Sprintf("tfbridge/providers/%s.go", providerName))
	defer func() {
		if e := writer.Flush(); e != nil {
			panic(e)
		}
	}()
}

func loopProviders(index int) []string {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/orgs/terraform-providers/repos?page=%d", index), nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN")))
	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println("Getting page of repo list", index)
	if err != nil {
		panic(err)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var parsed []map[string]interface{}
	var result []string
	_ = json.Unmarshal(content, &parsed)
	for _, value := range parsed {
		repoArchived := value["archived"].(bool)
		if repoArchived {
			continue
		}
		url := value["html_url"].(string)
		if strings.Index(url, "/terraform-provider-") == -1 {
			continue
		}
		item := url[strings.LastIndex(url, "-")+1:]
		result = append(result, item)
	}
	return result
}
