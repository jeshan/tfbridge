package crud

import (
	"fmt"
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jeshan/tfbridge/tfbridge/utils"
	"os"
	"plugin"
	"strings"
)

type ProviderCrud struct {
	Init           func() (terraform.ResourceProvider, error)
	Import         func(provider terraform.ResourceProvider, resourceID string, resourceType string) (*terraform.InstanceState, error)
	Create         func(terraform.ResourceProvider, string, map[string]interface{}) (*terraform.InstanceState, error)
	Update         func(terraform.ResourceProvider, string, string, map[string]interface{}, map[string]interface{}) (*terraform.InstanceState, error)
	Delete         func(terraform.ResourceProvider, string, string, map[string]interface{}, map[string]interface{}) (*terraform.InstanceState, error)
	DataSourceRead func(terraform.ResourceProvider, map[string]interface{}, string) (*terraform.InstanceState, error)
}

func Create(provider terraform.ResourceProvider, resourceType string, configuration map[string]interface{}) (*terraform.InstanceState, error) {
	info := &terraform.InstanceInfo{Type: resourceType}
	state := &terraform.InstanceState{}
	rawConfig, err := config.NewRawConfig(configuration)
	if err != nil {
		return nil, err
	}
	resourceConfig := terraform.NewResourceConfig(rawConfig)
	diff, err := provider.Diff(info, state, resourceConfig)
	if err != nil {
		return nil, err
	}
	fmt.Println("diff", diff)

	applied, err := apply(provider, info, state, diff)
	return applied, err
}

func apply(provider terraform.ResourceProvider, info *terraform.InstanceInfo, state *terraform.InstanceState, diff *terraform.InstanceDiff) (*terraform.InstanceState, error) {
	result, err := provider.Apply(info, state, diff)
	fmt.Println("result", result)
	fmt.Println("err", err)
	return result, err
}

func DataSourceRead(provider terraform.ResourceProvider, configuration map[string]interface{}, dataType string) (*terraform.InstanceState, error) {
	info := &terraform.InstanceInfo{Type: dataType}
	rawConfig, _ := config.NewRawConfig(configuration)
	diff, _ := provider.ReadDataDiff(info, terraform.NewResourceConfig(rawConfig))
	state, err := provider.ReadDataApply(info, diff)
	fmt.Println("state", state)
	fmt.Println("err", err)
	return state, err
}

func Update(provider terraform.ResourceProvider, resourceID string, resourceType string, oldConfig map[string]interface{}, currentConfig map[string]interface{}) (*terraform.InstanceState, error) {
	fmt.Println("updating.....")
	info := &terraform.InstanceInfo{Type: resourceType, Id: resourceID}

	for _, item := range []map[string]interface{}{oldConfig, currentConfig} {
		processedConfig := utils.ConvertToHashicorpConfiguration(item)
		state := &terraform.InstanceState{ID: resourceID, Attributes: processedConfig}
		rawConfig, err := config.NewRawConfig(currentConfig)
		if err != nil {
			return nil, err
		}
		diff, err := provider.Diff(info, state, terraform.NewResourceConfig(rawConfig))
		if err != nil {
			return nil, err
		}
		if diff == nil {
			// e.g when deleting an aws_iam_group_membership and user list was not specified or when updating it and cleared user list (TF docs says it was required though)
			continue
		}
		fmt.Println("diff", diff)
		result, err := apply(provider, info, state, diff)
		return result, err
	}
	panic("Could not determine diff for this resource update event")
}

func Import(provider terraform.ResourceProvider, resourceID string, resourceType string) (*terraform.InstanceState, error) {
	fmt.Println("importing", resourceType, "resource:", resourceID)
	info := &terraform.InstanceInfo{Type: resourceType}

	states, _ := provider.ImportState(info, resourceID)
	if len(states) != 1 {
		return nil, fmt.Errorf("import must return exactly one state, got %d states", len(states))
	}
	return provider.Refresh(info, states[0])
}

func Delete(provider terraform.ResourceProvider, resourceID, resourceType string, oldConfig map[string]interface{}, currentConfig map[string]interface{}) (*terraform.InstanceState, error) {
	fmt.Println("deleting.....")
	info := &terraform.InstanceInfo{Type: resourceType}
	// delete must have provider resource ID set to a real value (e.g for aws_iam_user)

	for _, item := range []map[string]interface{}{oldConfig, currentConfig} {
		processedConfiguration := utils.ConvertToHashicorpConfiguration(item)
		state := &terraform.InstanceState{ID: resourceID, Attributes: processedConfiguration}
		rawConfig, _ := config.NewRawConfig(item)
		diff, err := provider.Diff(info, state, terraform.NewResourceConfig(rawConfig))
		if err != nil {
			return nil, err
		}
		if diff == nil {
			// e.g when deleting an aws_iam_group_membership and user list was not specified or when updating it and cleared user list (TF docs says it was required though)
			continue
		}
		diff.SetDestroy(true)
		result, err := apply(provider, info, state, diff)
		return result, err
	}
	panic("Could not determine diff for this resource delete event")
}

func GetProvider(resourceType string) (terraform.ResourceProvider, error) {
	providerName := utils.GetProviderName(resourceType)
	var plug *plugin.Plugin
	var err error
	_, isLambda := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME")
	path := "dist/%s.so"
	if isLambda {
		path = "./%s.so"
	}
	plug, err = plugin.Open(fmt.Sprintf(path, providerName))
	if err != nil {
		return nil, err
	}
	symbol, err := plug.Lookup("CreateProvider")
	if err != nil {
		return nil, err
	}
	crud := ProviderCrud{
		Init:           symbol.(func() (terraform.ResourceProvider, error)),
		Create:         Create,
		Update:         Update,
		Delete:         Delete,
		DataSourceRead: DataSourceRead,
		Import:         Import,
	}
	return crud.Init()

}

func GetConfigurationMap(resourceProvider terraform.ResourceProvider) map[string]interface{} {
	provider := resourceProvider.(*schema.Provider)
	prefix, _ := getProviderPrefix(provider)
	return lookupEnvVars(getSchemaEnvVarNameMap(provider.Schema, prefix, schema.TypeInvalid))
}

func lookupEnvVars(input map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range input {
		// TODO: support maps better, e.g with _KEYS and _VALUES, or listing all env vars and filtering them (could be done for list as well)
		// TODO: test for bool, list of ints, and other non-string values (first, check how TF handles it)
		envName := value
		switch value.(type) {
		case []interface{}:
			envName = value.([]interface{})[0]
			var newList []interface{}
			for index := 0; ; index += 1 {
				envNameKey := fmt.Sprintf("%s_%d", envName.(string), index)
				lookupEnv, b := os.LookupEnv(envNameKey)
				if !b {
					break
				}
				newList = append(newList, lookupEnv)
			}
			if len(newList) != 0 {
				result[key] = newList
			}
		default:
			lookupEnv, b := os.LookupEnv(envName.(string))
			if b {
				result[key] = lookupEnv
			}
		}
	}
	return result
}

func getProviderPrefix(provider *schema.Provider) (string, error) {
	for key := range provider.ResourcesMap {
		return "TFBRIDGE_" + key[0:strings.Index(key, "_")], nil
	}
	for key := range provider.DataSourcesMap {
		return "TFBRIDGE_" + key[0:strings.Index(key, "_")], nil
	}
	return "", fmt.Errorf("no provider name found")
}

func getSchemaEnvVarNameMap(sch map[string]*schema.Schema, prefix string, parentType schema.ValueType) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range sch {
		var newSch map[string]*schema.Schema
		if value.Elem != nil {
			switch value.Elem.(type) {
			case *schema.Resource:
				newSch = value.Elem.(*schema.Resource).Schema
			case *schema.Schema:
				newSch = map[string]*schema.Schema{key: value.Elem.(*schema.Schema)}
			default:
				panic(fmt.Errorf("unsupported elem type: %v", value.Elem))
			}
		}

		switch value.Type {
		case schema.TypeMap:
			for k, v := range getSchemaEnvVarNameMap(newSch, prefix+"_"+key, value.Type) {
				//_, found := result[k]
				//if !found {
				newValue := map[string]interface{}{"": v}
				//map[string]interface{}{v.(string): ""} // TODO: test
				if parentType == schema.TypeMap {
					result[k] = map[string]interface{}{"": newValue}
				} else if parentType == schema.TypeList || parentType == schema.TypeSet {
					result[k] = []interface{}{newValue}
				} else if parentType == schema.TypeInvalid && value.Type == schema.TypeSet {
					result[k] = v // TODO: is this good?
				} else {
					result[k] = newValue
				}
				//}
			}
		case schema.TypeList, schema.TypeSet:
			for k, v := range getSchemaEnvVarNameMap(newSch, prefix+"_"+key, value.Type) {
				//_, found := result[k]
				newValue := []interface{}{v}
				//if !found {
				if parentType == schema.TypeMap {
					result[k] = map[string]interface{}{"": newValue}
				} else if parentType == schema.TypeList || parentType == schema.TypeSet {
					result[k] = []interface{}{newValue}
				} else if parentType == schema.TypeInvalid && value.Type == schema.TypeSet {
					// doesn't work when in a Schema (and not a Resource)
					result[k] = v // newValue: TODO: test
					switch value.Elem.(type) {
					//case *schema.Resource:
					//result[k] = v
					case *schema.Schema:
						result[k] = newValue
					}
				} else {
					result[k] = newValue
				}
				//}
			}
		default:
			newPrefix := prefix + "_" + key
			if strings.HasSuffix(prefix, key) {
				newPrefix = prefix
			}
			result[key] = strings.ToUpper(newPrefix)
		}
	}
	return result
}
