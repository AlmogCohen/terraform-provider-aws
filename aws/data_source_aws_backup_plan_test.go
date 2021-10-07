package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSBackupPlanDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_backup_plan.test"
	resourceName := "aws_backup_plan.test"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, backup.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsBackupPlanDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`Error getting Backup Plan`),
			},
			{
				Config: testAccAwsBackupPlanDataSourceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "version", resourceName, "version"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

const testAccAwsBackupPlanDataSourceConfig_nonExistent = `
data "aws_backup_plan" "test" {
  plan_id = "tf-acc-test-does-not-exist"
}
`

func testAccAwsBackupPlanDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%[1]d"
}

resource "aws_backup_plan" "test" {
  name = "tf_acc_test_backup_plan_%[1]d"

  rule {
    rule_name         = "tf_acc_test_backup_rule_%[1]d"
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"
  }

  tags = {
    Name = "Value%[1]d"
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}

data "aws_backup_plan" "test" {
  plan_id = aws_backup_plan.test.id
}
`, rInt)
}
