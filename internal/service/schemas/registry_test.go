package schemas_test

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfschemas "github.com/hashicorp/terraform-provider-aws/internal/service/schemas"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_schemas_registry", &resource.Sweeper{
		Name: "aws_schemas_registry",
		F:    sweepRegistries,
	})
}

func sweepRegistries(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SchemasConn
	input := &schemas.ListRegistriesInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListRegistriesPages(input, func(page *schemas.ListRegistriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, registry := range page.Registries {
			registryName := aws.StringValue(registry.RegistryName)

			input := &schemas.ListSchemasInput{
				RegistryName: aws.String(registryName),
			}

			err = conn.ListSchemasPages(input, func(page *schemas.ListSchemasOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, schema := range page.Schemas {
					schemaName := aws.StringValue(schema.SchemaName)
					if strings.HasPrefix(schemaName, "aws.") {
						continue
					}

					r := tfschemas.ResourceSchema()
					d := r.Data(nil)
					d.SetId(tfschemas.SchemaCreateResourceID(schemaName, registryName))
					err = r.Delete(d, client)

					if err != nil {
						log.Printf("[ERROR] %s", err)
						sweeperErrs = multierror.Append(sweeperErrs, err)
						continue
					}
				}

				return !lastPage
			})

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge Schemas Schemas: %w", err))
			}

			if strings.HasPrefix(registryName, "aws.") {
				continue
			}

			r := tfschemas.ResourceRegistry()
			d := r.Data(nil)
			d.SetId(registryName)
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge Schemas Registry sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge Schemas Registries: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccSchemasRegistry_basic(t *testing.T) {
	var v schemas.DescribeRegistryOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(schemas.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, schemas.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "schemas", fmt.Sprintf("registry/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchemasRegistry_disappears(t *testing.T) {
	var v schemas.DescribeRegistryOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(schemas.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, schemas.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfschemas.ResourceRegistry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSchemasRegistry_description(t *testing.T) {
	var v schemas.DescribeRegistryOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(schemas.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, schemas.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryDescriptionConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRegistryDescriptionConfig(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
			{
				Config: testAccRegistryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccSchemasRegistry_tags(t *testing.T) {
	var v schemas.DescribeRegistryOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(schemas.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, schemas.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRegistryTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRegistryTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasRegistryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckRegistryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_schemas_registry" {
			continue
		}

		_, err := tfschemas.FindRegistryByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EventBridge Schemas Registry %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSchemasRegistryExists(n string, v *schemas.DescribeRegistryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EventBridge Schemas Registry ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasConn

		output, err := tfschemas.FindRegistryByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRegistryConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRegistryDescriptionConfig(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccRegistryTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRegistryTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
