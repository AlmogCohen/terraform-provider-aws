package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/lex/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tflexmodelbuilding "github.com/hashicorp/terraform-provider-aws/internal/service/lexmodelbuilding"
	tflexmodelbuilding "github.com/hashicorp/terraform-provider-aws/internal/service/lexmodelbuilding"
	tflexmodelbuilding "github.com/hashicorp/terraform-provider-aws/internal/service/lexmodelbuilding"
)

const (
	//Lex Bot Statuses
	lexModeBuildingServicesStatusBuilding          = "BUILDING"
	lexModeBuildingServicesStatusFailed            = "FAILED"
	lexModeBuildingServicesStatusNotBuilt          = "NOT_BUILT"
	lexModeBuildingServicesStatusReady             = "READY"
	lexModeBuildingServicesStatusReadyBasicTesting = "READY_BASIC_TESTING"

	lexModelBuildingServiceStatusCreated  = "CREATED"
	lexModelBuildingServiceStatusNotFound = "NOTFOUND"
	lexModelBuildingServiceStatusUnknown  = "UNKNOWN"
)

func statusBotVersion(conn *lexmodelbuildingservice.LexModelBuildingService, name, version string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := tflexmodelbuilding.FindBotVersionByName(conn, name, version)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusLexSlotType(conn *lexmodelbuildingservice.LexModelBuildingService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetSlotTypeVersions(&lexmodelbuildingservice.GetSlotTypeVersionsInput{
			Name: aws.String(id),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}
		if err != nil {
			return nil, lexModelBuildingServiceStatusUnknown, err
		}

		if output == nil || len(output.SlotTypes) == 0 {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}

		return output, lexModelBuildingServiceStatusCreated, nil
	}
}

func statusLexIntent(conn *lexmodelbuildingservice.LexModelBuildingService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetIntentVersions(&lexmodelbuildingservice.GetIntentVersionsInput{
			Name: aws.String(id),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}
		if err != nil {
			return nil, lexModelBuildingServiceStatusUnknown, err
		}

		if output == nil || len(output.Intents) == 0 {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}

		return output, lexModelBuildingServiceStatusCreated, nil
	}
}

func statusLexBotAlias(conn *lexmodelbuildingservice.LexModelBuildingService, botAliasName, botName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(botAliasName),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}
		if err != nil {
			return nil, lexModelBuildingServiceStatusUnknown, err
		}
		if output == nil {
			return nil, lexModelBuildingServiceStatusNotFound, nil
		}

		return output, lexModelBuildingServiceStatusCreated, nil
	}
}