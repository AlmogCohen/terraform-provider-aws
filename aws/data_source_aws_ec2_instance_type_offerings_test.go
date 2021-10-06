package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSEc2InstanceTypeOfferingsDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2InstanceTypeOfferings(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2InstanceTypeOfferingsDataSourceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2InstanceTypeOfferingsInstanceTypes(dataSourceName),
				),
			},
		},
	})
}

func TestAccAWSEc2InstanceTypeOfferingsDataSource_LocationType(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2InstanceTypeOfferings(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2InstanceTypeOfferingsDataSourceConfigLocationType(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2InstanceTypeOfferingsInstanceTypes(dataSourceName),
					testAccCheckEc2InstanceTypeOfferingsLocations(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckEc2InstanceTypeOfferingsInstanceTypes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["instance_types.#"]; v == "0" {
			return fmt.Errorf("expected at least one instance_types result, got none")
		}

		if v := rs.Primary.Attributes["locations.#"]; v == "0" {
			return fmt.Errorf("expected at least one locations result, got none")
		}

		if v := rs.Primary.Attributes["location_types.#"]; v == "0" {
			return fmt.Errorf("expected at least one location_types result, got none")
		}

		return nil
	}
}

func testAccCheckEc2InstanceTypeOfferingsLocations(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["locations.#"]; v == "0" {
			return fmt.Errorf("expected at least one locations result, got none")
		}

		return nil
	}
}

func testAccPreCheckAWSEc2InstanceTypeOfferings(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeInstanceTypeOfferingsInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeInstanceTypeOfferings(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSEc2InstanceTypeOfferingsDataSourceConfigFilter() string {
	return `
data "aws_ec2_instance_type_offerings" "test" {
  filter {
    name   = "instance-type"
    values = ["t2.micro", "t3.micro"]
  }
}
`
}

func testAccAWSEc2InstanceTypeOfferingsDataSourceConfigLocationType() string {
	return acctest.ConfigAvailableAZsNoOptIn() + `
data "aws_ec2_instance_type_offerings" "test" {
  filter {
    name   = "location"
    values = [data.aws_availability_zones.available.names[0]]
  }

  location_type = "availability-zone"
}
`
}
