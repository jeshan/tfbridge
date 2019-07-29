package utils

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"regexp"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func GetProviderResourceType(cfnType string) string {
	startKey := "Custom::TfBridge-resource-"
	if strings.Index(strings.ToLower(cfnType), strings.ToLower(startKey)) != 0 {
		return ""
	}
	var str string
	if strings.Index(cfnType, startKey) == 0 {
		str = cfnType[len(startKey):]
	} else {
		str = cfnType[strings.LastIndex(cfnType, "::")+2:]
	}
	return ToSnakeCase(str)
}

func GetProviderDataType(cfnType string) string {
	startKey := "Custom::TfBridge-data-"
	if GetProviderResourceType(cfnType) != "" {
		return ""
	}
	var str string
	if strings.Index(cfnType, startKey) == 0 {
		str = cfnType[len(startKey):]
	} else {
		str = cfnType[strings.LastIndex(cfnType, "::")+2:]
	}
	return ToSnakeCase(str)
}

func GetProviderName(resourceType string) string {
	index := strings.Index(resourceType, "_")
	if index == -1 {
		return resourceType
	}
	return resourceType[:index]
}

func ConvertToHashicorpConfiguration(config map[string]interface{}) map[string]string {
	result := map[string]string{}
	for key, value := range config {
		// fmt.Println("key", key, "value", value, fmt.Sprintf("%T", value))
		switch value.(type) {
		case map[string]interface{}:
			castValue := value.(map[string]interface{})
			for k, v := range castValue {
				switch v.(type) {
				case string:
					castAgain := v.(string)
					result[key+"."+k] = castAgain
				default:
					bytes, _ := json.Marshal(v)
					result[key+"."+k] = string(bytes)
				}
			}
		case []interface{}:
		case []string:
			bytes, _ := json.Marshal(value)
			result[key] = string(bytes)
		default:
			result[key] = value.(string)
		}
	}
	return result
}

func AreAttributesEqual(attributes map[string]string, properties map[string]interface{}) bool {
	attributes = removeUnusedTfProperties(attributes)
	processed := removeNonTfProperties(properties)
	return cmp.Equal(attributes, processed)
}

func removeUnusedTfProperties(properties map[string]string) map[string]string {
	processed := map[string]string{}
	for k, v := range properties {
		if strings.HasSuffix(k, ".#") || strings.HasSuffix(k, ".%") {
			continue
		}
		processed[k] = v
	}
	return processed
}

func removeNonTfProperties(properties map[string]interface{}) map[string]string {
	processed := map[string]string{}
	for k, v := range ConvertToHashicorpConfiguration(properties) {
		if k == "ServiceToken" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(k), "tfbridge_") {
			continue
		}
		processed[k] = v
	}
	return processed
}

func CompareAttributes(importedAttributes map[string]string, properties map[string]interface{}, logicalResourceID string) error {
	if AreAttributesEqual(importedAttributes, properties) {
		return nil
	}
	attrs, _ := json.Marshal(removeUnusedTfProperties(importedAttributes))
	props, _ := json.Marshal(removeNonTfProperties(properties))
	return fmt.Errorf("strict import failed, expected to have properties %v in the %v resource, got %v instead", string(attrs), logicalResourceID, string(props))
}
