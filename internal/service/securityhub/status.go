package securityhub

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/securityhub/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
)

const (
	// AdminStatus NotFound
	adminStatusNotFound = "NotFound"

	// AdminStatus Unknown
	adminStatusUnknown = "Unknown"

	standardsStatusNotFound = "NotFound"
)

// statusAdminAccountAdmin fetches the AdminAccount and its AdminStatus
func statusAdminAccountAdmin(conn *securityhub.SecurityHub, adminAccountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		adminAccount, err := tfsecurityhub.FindAdminAccount(conn, adminAccountID)

		if err != nil {
			return nil, adminStatusUnknown, err
		}

		if adminAccount == nil {
			return adminAccount, adminStatusNotFound, nil
		}

		return adminAccount, aws.StringValue(adminAccount.Status), nil
	}
}

func statusStandardsSubscription(conn *securityhub.SecurityHub, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := tfsecurityhub.FindStandardsSubscriptionByARN(conn, arn)

		if tfresource.NotFound(err) {
			// Return a fake result and status to deal with the INCOMPLETE subscription status
			// being a target for both Create and Delete.
			return "", standardsStatusNotFound, nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.StandardsStatus), nil
	}
}
