package route53_test

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
)

// add sweeper to delete resources
func init() {
	resource.AddTestSweepers("aws_route53_key_signing_key", &resource.Sweeper{
		Name: "aws_route53_key_signing_key",
		F:    testSweepRoute53KeySigningKeys,
	})
}

func testSweepRoute53KeySigningKeys(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).Route53Conn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &route53.ListHostedZonesInput{}

	err = conn.ListHostedZonesPages(input, func(page *route53.ListHostedZonesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

	MAIN:
		for _, detail := range page.HostedZones {
			if detail == nil {
				continue
			}

			id := aws.StringValue(detail.Id)

			for _, domain := range hostedZonesToPreserve() {
				if strings.Contains(aws.StringValue(detail.Name), domain) {
					log.Printf("[DEBUG] Skipping Route53 Hosted Zone (%s): %s", domain, id)
					continue MAIN
				}
			}

			dnsInput := &route53.GetDNSSECInput{
				HostedZoneId: detail.Id,
			}

			output, err := conn.GetDNSSEC(dnsInput)

			if tfawserr.ErrMessageContains(err, route53.ErrCodeInvalidArgument, "private hosted zones") {
				continue
			}

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error getting Route53 DNS SEC for %s: %w", region, err))
			}

			for _, dns := range output.KeySigningKeys {
				r := ResourceKeySigningKey()
				d := r.Data(nil)
				d.SetId(id)
				d.Set("hosted_zone_id", id)
				d.Set("name", dns.Name)
				d.Set("status", dns.Status)

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}

		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error getting Route53 Key-Signing Keys for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestratorContext(context.Background(), sweepResources, 0*time.Millisecond, 1*time.Minute, 30*time.Second, 30*time.Second, 10*time.Minute); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Route53 Key-Signing Keys for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Route53 Key-Signing Keys sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAwsRoute53KeySigningKey_basic(t *testing.T) {
	kmsKeyResourceName := "aws_kms_key.test"
	route53ZoneResourceName := "aws_route53_zone.test"
	resourceName := "aws_route53_key_signing_key.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckRoute53KeySigningKey(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53KeySigningKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53KeySigningKeyConfig_Name(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53KeySigningKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "digest_algorithm_mnemonic", "SHA-256"),
					resource.TestCheckResourceAttr(resourceName, "digest_algorithm_type", "2"),
					resource.TestMatchResourceAttr(resourceName, "digest_value", regexp.MustCompile(`^[0-9A-F]+$`)),
					resource.TestMatchResourceAttr(resourceName, "dnskey_record", regexp.MustCompile(`^257 [0-9]+ [0-9]+ [a-zA-Z0-9+/]+={0,3}$`)),
					resource.TestMatchResourceAttr(resourceName, "ds_record", regexp.MustCompile(`^[0-9]+ [0-9]+ [0-9]+ [0-9A-F]+$`)),
					resource.TestCheckResourceAttr(resourceName, "flag", "257"),
					resource.TestCheckResourceAttrPair(resourceName, "hosted_zone_id", route53ZoneResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "key_management_service_arn", kmsKeyResourceName, "arn"),
					resource.TestMatchResourceAttr(resourceName, "key_tag", regexp.MustCompile(`^[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestMatchResourceAttr(resourceName, "public_key", regexp.MustCompile(`^[a-zA-Z0-9+/]+={0,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm_mnemonic", "ECDSAP256SHA256"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm_type", "13"),
					resource.TestCheckResourceAttr(resourceName, "status", tfroute53.KeySigningKeyStatusActive),
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

func TestAccAwsRoute53KeySigningKey_disappears(t *testing.T) {
	resourceName := "aws_route53_key_signing_key.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckRoute53KeySigningKey(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53KeySigningKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53KeySigningKeyConfig_Name(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53KeySigningKeyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceKeySigningKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsRoute53KeySigningKey_Status(t *testing.T) {
	resourceName := "aws_route53_key_signing_key.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckRoute53KeySigningKey(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53KeySigningKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53KeySigningKeyConfig_Status(rName, domainName, tfroute53.KeySigningKeyStatusInactive),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53KeySigningKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", tfroute53.KeySigningKeyStatusInactive),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsRoute53KeySigningKeyConfig_Status(rName, domainName, tfroute53.KeySigningKeyStatusActive),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53KeySigningKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", tfroute53.KeySigningKeyStatusActive),
				),
			},
			{
				Config: testAccAwsRoute53KeySigningKeyConfig_Status(rName, domainName, tfroute53.KeySigningKeyStatusInactive),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53KeySigningKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", tfroute53.KeySigningKeyStatusInactive),
				),
			},
		},
	})
}

func testAccCheckAwsRoute53KeySigningKeyDestroy(s *terraform.State) error {
	conn := testAccProviderRoute53KeySigningKey.Meta().(*conns.AWSClient).Route53Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_key_signing_key" {
			continue
		}

		keySigningKey, err := tfroute53.FindKeySigningKeyByResourceID(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
			continue
		}

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchKeySigningKey) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Route 53 Key Signing Key (%s): %w", rs.Primary.ID, err)
		}

		if keySigningKey != nil {
			return fmt.Errorf("Route 53 Key Signing Key (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsRoute53KeySigningKeyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has not set its id", resourceName)
		}

		conn := testAccProviderRoute53KeySigningKey.Meta().(*conns.AWSClient).Route53Conn

		keySigningKey, err := tfroute53.FindKeySigningKeyByResourceID(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading Route 53 Key Signing Key (%s): %w", rs.Primary.ID, err)
		}

		if keySigningKey == nil {
			return fmt.Errorf("Route 53 Key Signing Key (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAwsRoute53KeySigningKeyConfig_Base(rName, domainName string) string {
	return acctest.ConfigCompose(
		testAccRoute53KeySigningKeyRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  customer_master_key_spec = "ECC_NIST_P256"
  deletion_window_in_days  = 7
  key_usage                = "SIGN_VERIFY"
  policy = jsonencode({
    Statement = [
      {
        Action = [
          "kms:DescribeKey",
          "kms:GetPublicKey",
          "kms:Sign",
        ],
        Effect = "Allow"
        Principal = {
          Service = "api-service.dnssec.route53.aws.internal"
        }
        Sid = "Allow Route 53 DNSSEC Service"
      },
      {
        Action = "kms:*"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
    ]
    Version = "2012-10-17"
  })
}

resource "aws_route53_zone" "test" {
  name = %[2]q
}
`, rName, domainName))
}

func testAccAwsRoute53KeySigningKeyConfig_Name(rName, domainName string) string {
	return acctest.ConfigCompose(
		testAccAwsRoute53KeySigningKeyConfig_Base(rName, domainName),
		fmt.Sprintf(`
resource "aws_route53_key_signing_key" "test" {
  hosted_zone_id             = aws_route53_zone.test.id
  key_management_service_arn = aws_kms_key.test.arn
  name                       = %[1]q
}
`, rName))
}

func testAccAwsRoute53KeySigningKeyConfig_Status(rName, domainName, status string) string {
	return acctest.ConfigCompose(
		testAccAwsRoute53KeySigningKeyConfig_Base(rName, domainName),
		fmt.Sprintf(`
resource "aws_route53_key_signing_key" "test" {
  hosted_zone_id             = aws_route53_zone.test.id
  key_management_service_arn = aws_kms_key.test.arn
  name                       = %[1]q
  status                     = %[2]q
}
`, rName, status))
}

// Route 53 Key Signing Key can only be enabled with KMS Keys in specific regions,

// testAccRoute53KeySigningKeyRegion is the chosen Route 53 Key Signing Key testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccRoute53KeySigningKeyRegion string

// testAccProviderRoute53KeySigningKey is the Route 53 Key Signing Key provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckRoute53KeySigningKey(t) must be called before using this provider instance.
var testAccProviderRoute53KeySigningKey *schema.Provider

// testAccProviderRoute53KeySigningKeyConfigure ensures the provider is only configured once
var testAccProviderRoute53KeySigningKeyConfigure sync.Once

// testAccPreCheckRoute53KeySigningKey verifies AWS credentials and that Route 53 Key Signing Key is supported
func testAccPreCheckRoute53KeySigningKey(t *testing.T) {
	acctest.PreCheckPartitionHasService(route53.EndpointsID, t)

	region := testAccGetRoute53KeySigningKeyRegion()

	if region == "" {
		t.Skip("Route 53 Key Signing Key not available in this AWS Partition")
	}

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderRoute53KeySigningKeyConfigure.Do(func() {
		testAccProviderRoute53KeySigningKey = provider.Provider()

		config := map[string]interface{}{
			"region": region,
		}

		diags := testAccProviderRoute53KeySigningKey.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring Route 53 Key Signing Key provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccRoute53KeySigningKeyRegionProviderConfig is the Terraform provider configuration for Route 53 Key Signing Key region testing
//
// Testing Route 53 Key Signing Key assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccRoute53KeySigningKeyRegionProviderConfig() string {
	return acctest.ConfigRegionalProvider(testAccGetRoute53KeySigningKeyRegion())
}

// testAccGetRoute53KeySigningKeyRegion returns the Route 53 Key Signing Key region for testing
func testAccGetRoute53KeySigningKeyRegion() string {
	if testAccRoute53KeySigningKeyRegion != "" {
		return testAccRoute53KeySigningKeyRegion
	}

	// AWS Commercial: https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring-dnssec-cmk-requirements.html
	// AWS GovCloud (US) - not available yet: https://docs.aws.amazon.com/govcloud-us/latest/UserGuide/govcloud-r53.html
	// AWS China - not available yet: https://docs.amazonaws.cn/en_us/aws/latest/userguide/route53.html
	switch acctest.Partition() {
	case endpoints.AwsPartitionID:
		testAccRoute53KeySigningKeyRegion = endpoints.UsEast1RegionID
	}

	return testAccRoute53KeySigningKeyRegion
}