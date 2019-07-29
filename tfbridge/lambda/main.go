package main

import (
	"context"
	"fmt"
	"github.com/jeshan/tfbridge/tfbridge/crud"
	"github.com/jeshan/tfbridge/tfbridge/utils"
	"strings"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hashicorp/terraform/terraform"
)

func tfResource(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	fmt.Printf("event %+v\n", event)
	fmt.Printf("context %+v\n", ctx)
	fmt.Printf("physicalResourceID %+v\n", physicalResourceID)
	fmt.Printf("data %+v\n", data)
	requestType := string(event.RequestType)

	separator := ";"
	constantPhysicalResourceID := event.StackID + separator + event.LogicalResourceID
	physicalResourceID = constantPhysicalResourceID
	split := strings.Split(event.PhysicalResourceID, separator)
	resourceID := split[len(split)-1]

	createFailed := "<create-failed>"
	if requestType == "Delete" {
		// attempting to fix scenario: when attempting to delete a resource for which creation just failed
		if event.PhysicalResourceID == createFailed {
			fmt.Println("Resource did not even get created; skipping", err)
			return "", map[string]interface{}{}, nil
		}
	}
	resourceType := utils.GetProviderResourceType(event.ResourceType)
	dataType := utils.GetProviderDataType(event.ResourceType)
	fmt.Println("got resource type", resourceType, "and data type", dataType)
	var provider terraform.ResourceProvider
	isData := dataType != "" // else, is a normal resource
	var typeOfResource string
	if isData {
		provider, err = crud.GetProvider(dataType)
		typeOfResource = "data"
	} else {
		provider, err = crud.GetProvider(resourceType)
		typeOfResource = "resource"
	}
	if err != nil {
		fmt.Println("Error initialising provider", err)
		return physicalResourceID, map[string]interface{}{}, err
	}

	var state *terraform.InstanceState
	properties := event.ResourceProperties
	importID, _ := properties["TFBRIDGE_ID"]
	if isData {
		if requestType == "Delete" {
			return event.PhysicalResourceID, nil, nil
		}
		state, err = crud.DataSourceRead(provider, properties, dataType)
	} else if requestType == "Create" {
		if isImportMode(properties) {
			physicalResourceID = constantPhysicalResourceID
			state, err = crud.Import(provider, importID.(string), resourceType)
			if state != nil {
				err = utils.CompareAttributes(state.Attributes, properties, event.LogicalResourceID)
			}
		} else {
			state, err = crud.Create(provider, resourceType, properties)
		}
	} else if requestType == "Update" {
		// facts: iam username must be the ID as determined by d.Id()
		state, err = crud.Update(provider, resourceID, resourceType, event.OldResourceProperties, properties)
	} else if requestType == "Delete" {
		state, err = crud.Delete(provider, resourceID, resourceType, event.OldResourceProperties, event.ResourceProperties)
	} else {
		panic("Unrecognised request type: " + requestType)
	}
	newData := map[string]interface{}{}
	if state != nil {
		// state is mysteriously nil when created an aws_iam_group_membership
		for key, value := range state.Attributes {
			newData[key] = value
		}
		if requestType == "Delete" {
			if err != nil {
				err = fmt.Errorf("WARN: delete failed; haven't figured out a reliable stateless way to tell CFN to skip resource replacement yet as it has already been handled by Terraform.", err)
				/*_, ok := state.Attributes["id"]
				if ok {
					err = nil
				}*/
			}
		}
		physicalResourceID += separator + typeOfResource + separator + state.ID
	} else {
		if requestType == "Create" && isImportMode(properties) {
			if err == nil {
				err = fmt.Errorf("Import failed, likely because %s does not exist", importID)
			}
			return physicalResourceID, map[string]interface{}{}, err
		}
	}
	if requestType == "Delete" {
		physicalResourceID = event.PhysicalResourceID
	}
	fmt.Println("operation returned state", state, "err", err, "newData", newData)
	if err == nil {
		return physicalResourceID, newData, nil
	}
	if requestType == "Create" {
		physicalResourceID = createFailed
	}
	return physicalResourceID, map[string]interface{}{}, err
}

func isImportMode(properties map[string]interface{}) bool {
	value, ok := properties["TFBRIDGE_MODE"]
	if !ok {
		return false
	}
	return strings.Index(strings.ToLower(value.(string)), "import") == 0
}

func main() {
	lambda.Start(cfn.LambdaWrap(tfResource))
}
