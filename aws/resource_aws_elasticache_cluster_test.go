package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	RegisterServiceErrorCheckFunc(elasticache.EndpointsID, testAccErrorCheckSkipElasticache)

	resource.AddTestSweepers("aws_elasticache_cluster", &resource.Sweeper{
		Name: "aws_elasticache_cluster",
		F:    testSweepElasticacheClusters,
		Dependencies: []string{
			"aws_elasticache_replication_group",
		},
	})
}

func testSweepElasticacheClusters(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).elasticacheconn

	var sweeperErrs *multierror.Error

	input := &elasticache.DescribeCacheClustersInput{
		ShowCacheClustersNotInReplicationGroups: aws.Bool(true),
	}
	err = conn.DescribeCacheClustersPages(input, func(page *elasticache.DescribeCacheClustersOutput, lastPage bool) bool {
		if len(page.CacheClusters) == 0 {
			log.Print("[DEBUG] No ElastiCache Replicaton Groups to sweep")
			return false
		}

		for _, cluster := range page.CacheClusters {
			id := aws.StringValue(cluster.CacheClusterId)

			log.Printf("[INFO] Deleting ElastiCache Cluster: %s", id)
			err := deleteElasticacheCacheCluster(conn, id, "")
			if err != nil {
				log.Printf("[ERROR] Failed to delete ElastiCache Cache Cluster (%s): %s", id, err)
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting ElastiCache Cache Cluster (%s): %w", id, err))
			}
			_, err = waiter.CacheClusterDeleted(conn, id, waiter.CacheClusterDeletedTimeout)
			if err != nil {
				log.Printf("[ERROR] Failed waiting for ElastiCache Cache Cluster (%s) to be deleted: %s", id, err)
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting ElastiCache Cache Cluster (%s): waiting for completion: %w", id, err))
			}
		}
		return !lastPage
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping ElastiCache Cluster sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error retrieving ElastiCache Clusters: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func testAccErrorCheckSkipElasticache(t *testing.T) resource.ErrorCheckFunc {
	return testAccErrorCheckSkipMessagesContaining(t,
		"is not suppored in this region",
	)
}

func TestAccAWSElasticacheCluster_Engine_Memcached(t *testing.T) {
	var ec elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_Engine_Memcached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_address"),
					resource.TestCheckResourceAttr(resourceName, "engine", "memcached"),
					resource.TestCheckResourceAttr(resourceName, "port", "11211"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
				},
			},
		},
	})
}

func TestAccAWSElasticacheCluster_Engine_Redis(t *testing.T) {
	var ec elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_Engine_Redis(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "port", "6379"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
				},
			},
		},
	})
}

func TestAccAWSElasticacheCluster_Port_Redis_Default(t *testing.T) {
	var ec elasticache.CacheCluster
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_RedisDefaultPort,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.test", &ec),
					resource.TestCheckResourceAttr("aws_security_group_rule.test", "to_port", "6379"),
					resource.TestCheckResourceAttr("aws_security_group_rule.test", "from_port", "6379"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_ParameterGroupName_Default(t *testing.T) {
	var ec elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_ParameterGroupName(rName, "memcached", "1.4.34", "default.memcached1.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "engine", "memcached"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.34"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "1.4.34"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.memcached1.4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
				},
			},
		},
	})
}

func TestAccAWSElasticacheCluster_Port(t *testing.T) {
	var ec elasticache.CacheCluster
	port := 11212
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_Port(rName, port),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_address"),
					resource.TestCheckResourceAttr(resourceName, "engine", "memcached"),
					resource.TestCheckResourceAttr(resourceName, "port", strconv.Itoa(port)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
				},
			},
		},
	})
}

func TestAccAWSElasticacheCluster_SecurityGroup_Ec2Classic(t *testing.T) {
	var ec elasticache.CacheCluster
	resourceName := "aws_elasticache_cluster.test"
	resourceSecurityGroupName := "aws_elasticache_security_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_SecurityGroup_Ec2Classic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSecurityGroupExists(resourceSecurityGroupName),
					testAccCheckAWSElasticacheClusterEc2ClassicExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_address"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_snapshotsWithUpdates(t *testing.T) {
	var ec elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_snapshots(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.test", &ec),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "snapshot_window", "05:00-09:00"),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "snapshot_retention_limit", "3"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_snapshotsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.test", &ec),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "snapshot_window", "07:00-09:00"),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "snapshot_retention_limit", "7"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NumCacheNodes_Decrease(t *testing.T) {
	var ec elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_NumCacheNodes(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "3"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NumCacheNodes(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "1"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NumCacheNodes_Increase(t *testing.T) {
	var ec elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_NumCacheNodes(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "1"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NumCacheNodes(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "3"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NumCacheNodes_IncreaseWithPreferredAvailabilityZones(t *testing.T) {
	var ec elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_NumCacheNodesWithPreferredAvailabilityZones(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "1"),
					resource.TestCheckResourceAttr(resourceName, "preferred_availability_zones.#", "1"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NumCacheNodesWithPreferredAvailabilityZones(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "3"),
					resource.TestCheckResourceAttr(resourceName, "preferred_availability_zones.#", "3"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_vpc(t *testing.T) {
	var csg elasticache.CacheSubnetGroup
	var ec elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterInVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSubnetGroupExists("aws_elasticache_subnet_group.test", &csg),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.test", &ec),
					testAccCheckAWSElasticacheClusterAttributes(&ec),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_multiAZInVpc(t *testing.T) {
	var csg elasticache.CacheSubnetGroup
	var ec elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterMultiAZInVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSubnetGroupExists("aws_elasticache_subnet_group.test", &csg),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.test", &ec),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "availability_zone", "Multiple"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_AZMode_Memcached(t *testing.T) {
	var cluster elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Memcached(rName, "unknown"),
				ExpectError: regexp.MustCompile(`expected az_mode to be one of .*, got unknown`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Memcached(rName, "cross-az"),
				ExpectError: regexp.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_AZMode_Memcached(rName, "single-az"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "az_mode", "single-az"),
				),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Memcached(rName, "cross-az"),
				ExpectError: regexp.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_AZMode_Redis(t *testing.T) {
	var cluster elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Redis(rName, "unknown"),
				ExpectError: regexp.MustCompile(`expected az_mode to be one of .*, got unknown`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Redis(rName, "cross-az"),
				ExpectError: regexp.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_AZMode_Redis(rName, "single-az"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "az_mode", "single-az"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_EngineVersion_Memcached(t *testing.T) {
	var pre, mid, post elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Memcached(rName, "1.4.33"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.33"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "1.4.33"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Memcached(rName, "1.4.24"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &mid),
					testAccCheckAWSElasticacheClusterRecreated(&pre, &mid),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.24"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "1.4.24"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Memcached(rName, "1.4.34"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &post),
					testAccCheckAWSElasticacheClusterNotRecreated(&mid, &post),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.34"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "1.4.34"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_EngineVersion_Redis(t *testing.T) {
	var v1, v2, v3, v4, v5 elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Redis(rName, "3.2.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.6"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "3.2.6"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Redis(rName, "3.2.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &v2),
					testAccCheckAWSElasticacheClusterRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.4"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "3.2.4"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Redis(rName, "3.2.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &v3),
					testAccCheckAWSElasticacheClusterNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.10"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "3.2.10"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Redis(rName, "6.x"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &v4),
					testAccCheckAWSElasticacheClusterNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "6.x"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.[[:digit:]]+\.[[:digit:]]+$`)),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Redis(rName, "5.0.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &v5),
					testAccCheckAWSElasticacheClusterRecreated(&v4, &v5),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.0.6"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "5.0.6"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NodeTypeResize_Memcached(t *testing.T) {
	var pre, post elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_NodeType_Memcached(rName, "cache.t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.small"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NodeType_Memcached(rName, "cache.t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &post),
					testAccCheckAWSElasticacheClusterRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.medium"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NodeTypeResize_Redis(t *testing.T) {
	var pre, post elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_NodeType_Redis(rName, "cache.t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.small"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NodeType_Redis(rName, "cache.t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &post),
					testAccCheckAWSElasticacheClusterNotRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.medium"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NumCacheNodes_Redis(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_NumCacheNodes_Redis(rName, 2),
				ExpectError: regexp.MustCompile(`engine "redis" does not support num_cache_nodes > 1`),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_ReplicationGroupID_AvailabilityZone(t *testing.T) {
	var cluster elasticache.CacheCluster
	var replicationGroup elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	clusterResourceName := "aws_elasticache_cluster.test"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_ReplicationGroupID_AvailabilityZone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(replicationGroupResourceName, &replicationGroup),
					testAccCheckAWSElasticacheClusterExists(clusterResourceName, &cluster),
					testAccCheckAWSElasticacheClusterReplicationGroupIDAttribute(&cluster, &replicationGroup),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_ReplicationGroupID_SingleReplica(t *testing.T) {
	var cluster elasticache.CacheCluster
	var replicationGroup elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	clusterResourceName := "aws_elasticache_cluster.test.0"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_ReplicationGroupID_Replica(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(replicationGroupResourceName, &replicationGroup),
					testAccCheckAWSElasticacheClusterExists(clusterResourceName, &cluster),
					testAccCheckAWSElasticacheClusterReplicationGroupIDAttribute(&cluster, &replicationGroup),
					resource.TestCheckResourceAttr(clusterResourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(clusterResourceName, "node_type", "cache.t3.medium"),
					resource.TestCheckResourceAttr(clusterResourceName, "port", "6379"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_ReplicationGroupID_MultipleReplica(t *testing.T) {
	var cluster1, cluster2 elasticache.CacheCluster
	var replicationGroup elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	clusterResourceName1 := "aws_elasticache_cluster.test.0"
	clusterResourceName2 := "aws_elasticache_cluster.test.1"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_ReplicationGroupID_Replica(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(replicationGroupResourceName, &replicationGroup),

					testAccCheckAWSElasticacheClusterExists(clusterResourceName1, &cluster1),
					testAccCheckAWSElasticacheClusterReplicationGroupIDAttribute(&cluster1, &replicationGroup),
					resource.TestCheckResourceAttr(clusterResourceName1, "engine", "redis"),
					resource.TestCheckResourceAttr(clusterResourceName1, "node_type", "cache.t3.medium"),
					resource.TestCheckResourceAttr(clusterResourceName1, "port", "6379"),

					testAccCheckAWSElasticacheClusterExists(clusterResourceName2, &cluster2),
					testAccCheckAWSElasticacheClusterReplicationGroupIDAttribute(&cluster2, &replicationGroup),
					resource.TestCheckResourceAttr(clusterResourceName2, "engine", "redis"),
					resource.TestCheckResourceAttr(clusterResourceName2, "node_type", "cache.t3.medium"),
					resource.TestCheckResourceAttr(clusterResourceName2, "port", "6379"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_Memcached_FinalSnapshot(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_Memcached_FinalSnapshot(rName),
				ExpectError: regexp.MustCompile(`engine "memcached" does not support final_snapshot_identifier`),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_Redis_FinalSnapshot(t *testing.T) {
	var cluster elasticache.CacheCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_Redis_FinalSnapshot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "final_snapshot_identifier", rName),
				),
			},
		},
	})
}

func testAccCheckAWSElasticacheClusterAttributes(v *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v.NotificationConfiguration == nil {
			return fmt.Errorf("Expected NotificationConfiguration for ElastiCache Cluster (%s)", *v.CacheClusterId)
		}

		if strings.ToLower(*v.NotificationConfiguration.TopicStatus) != "active" {
			return fmt.Errorf("Expected NotificationConfiguration status to be 'active', got (%s)", *v.NotificationConfiguration.TopicStatus)
		}

		return nil
	}
}

func testAccCheckAWSElasticacheClusterReplicationGroupIDAttribute(cluster *elasticache.CacheCluster, replicationGroup *elasticache.ReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if cluster.ReplicationGroupId == nil {
			return errors.New("expected cluster ReplicationGroupId to be set")
		}

		if aws.StringValue(cluster.ReplicationGroupId) != aws.StringValue(replicationGroup.ReplicationGroupId) {
			return errors.New("expected cluster ReplicationGroupId to equal replication group ID")
		}

		return nil
	}
}

func testAccCheckAWSElasticacheClusterNotRecreated(i, j *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CacheClusterCreateTime).Equal(aws.TimeValue(j.CacheClusterCreateTime)) {
			return errors.New("ElastiCache Cluster was recreated")
		}

		return nil
	}
}

func testAccCheckAWSElasticacheClusterRecreated(i, j *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CacheClusterCreateTime).Equal(aws.TimeValue(j.CacheClusterCreateTime)) {
			return errors.New("ElastiCache Cluster was not recreated")
		}

		return nil
	}
}

func testAccCheckAWSElasticacheClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_cluster" {
			continue
		}
		_, err := finder.CacheClusterByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("ElastiCache Cache Cluster (%s) still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckAWSElasticacheClusterExists(n string, v *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No cache cluster ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
		resp, err := conn.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
			CacheClusterId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("ElastiCache error: %v", err)
		}

		for _, c := range resp.CacheClusters {
			if *c.CacheClusterId == rs.Primary.ID {
				*v = *c
			}
		}

		return nil
	}
}

func testAccCheckAWSElasticacheClusterEc2ClassicExists(n string, v *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No cache cluster ID is set")
		}

		conn := testAccProviderEc2Classic.Meta().(*AWSClient).elasticacheconn

		input := &elasticache.DescribeCacheClustersInput{
			CacheClusterId: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeCacheClusters(input)

		if err != nil {
			return fmt.Errorf("error describing ElastiCache Cluster (%s): %w", rs.Primary.ID, err)
		}

		for _, c := range output.CacheClusters {
			if aws.StringValue(c.CacheClusterId) == rs.Primary.ID {
				*v = *c

				return nil
			}
		}

		return fmt.Errorf("ElastiCache Cluster (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSElasticacheClusterConfig_Engine_Memcached(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = "%s"
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
}
`, rName)
}

func testAccAWSElasticacheClusterConfig_Engine_Redis(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = "%s"
  engine          = "redis"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
}
`, rName)
}

func testAccAWSElasticacheClusterConfig_ParameterGroupName(rName, engine, engineVersion, parameterGroupName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id           = %q
  engine               = %q
  engine_version       = %q
  node_type            = "cache.t3.small"
  num_cache_nodes      = 1
  parameter_group_name = %q
}
`, rName, engine, engineVersion, parameterGroupName)
}

func testAccAWSElasticacheClusterConfig_Port(rName string, port int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = "%s"
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
  port            = %d
}
`, rName, port)
}

func testAccAWSElasticacheClusterConfig_SecurityGroup_Ec2Classic(rName string) string {
	return composeConfig(
		testAccEc2ClassicRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSElastiCacheCluster_basic"
  }
}

resource "aws_elasticache_security_group" "test" {
  name                 = %[1]q
  description          = "tf-test-security-group-descr"
  security_group_names = [aws_security_group.test.name]
}

resource "aws_elasticache_cluster" "test" {
  cluster_id = %[1]q
  engine     = "memcached"

  # tflint-ignore: aws_elasticache_cluster_previous_type
  node_type            = "cache.m3.medium"
  num_cache_nodes      = 1
  port                 = 11211
  security_group_names = [aws_elasticache_security_group.test.name]
}
`, rName))
}

func testAccAWSElasticacheClusterConfig_snapshots(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id               = "tf-%s"
  engine                   = "redis"
  node_type                = "cache.t3.small"
  num_cache_nodes          = 1
  port                     = 6379
  snapshot_window          = "05:00-09:00"
  snapshot_retention_limit = 3
}
`, rName)
}

func testAccAWSElasticacheClusterConfig_snapshotsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id               = "tf-%s"
  engine                   = "redis"
  node_type                = "cache.t3.small"
  num_cache_nodes          = 1
  port                     = 6379
  snapshot_window          = "07:00-09:00"
  snapshot_retention_limit = 7
  apply_immediately        = true
}
`, rName)
}

func testAccAWSElasticacheClusterConfig_NumCacheNodes(rName string, numCacheNodes int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = "%s"
  engine            = "memcached"
  node_type         = "cache.t3.small"
  num_cache_nodes   = %d
}
`, rName, numCacheNodes)
}

func testAccAWSElasticacheClusterConfig_NumCacheNodesWithPreferredAvailabilityZones(rName string, numCacheNodes int) string {
	preferredAvailabilityZones := make([]string, numCacheNodes)
	for i := range preferredAvailabilityZones {
		preferredAvailabilityZones[i] = `data.aws_availability_zones.available.names[0]`
	}

	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elasticache_cluster" "test" {
  apply_immediately            = true
  cluster_id                   = "%s"
  engine                       = "memcached"
  node_type                    = "cache.t3.small"
  num_cache_nodes              = %d
  preferred_availability_zones = [%s]
}
`, rName, numCacheNodes, strings.Join(preferredAvailabilityZones, ","))
}

func testAccAWSElasticacheClusterInVPCConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-cluster-in-vpc"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-cluster-in-vpc"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = "tf-test-cache-subnet-group-descr"
  subnet_ids  = [aws_subnet.test.id]
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_cluster" "test" {
  # Including uppercase letters in this name to ensure
  # that we correctly handle the fact that the API
  # normalizes names to lowercase.
  cluster_id             = %[1]q
  node_type              = "cache.t3.small"
  num_cache_nodes        = 1
  engine                 = "redis"
  engine_version         = "2.8.19"
  port                   = 6379
  subnet_group_name      = aws_elasticache_subnet_group.test.name
  security_group_ids     = [aws_security_group.test.id]
  parameter_group_name   = "default.redis2.8"
  notification_topic_arn = aws_sns_topic.test.arn
  availability_zone      = data.aws_availability_zones.available.names[0]
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAWSElasticacheClusterMultiAZInVPCConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-cluster-multi-az-in-vpc"
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-cluster-multi-az-in-vpc-foo"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.16.0/20"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-elasticache-cluster-multi-az-in-vpc-bar"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = "tf-test-cache-subnet-group-descr"
  subnet_ids = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
  ]
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_cluster" "test" {
  cluster_id         = %[1]q
  engine             = "memcached"
  node_type          = "cache.t3.small"
  num_cache_nodes    = 2
  port               = 11211
  subnet_group_name  = aws_elasticache_subnet_group.test.name
  security_group_ids = [aws_security_group.test.id]
  az_mode            = "cross-az"
  preferred_availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1]
  ]
}
`, rName)
}

var testAccAWSElasticacheClusterConfig_RedisDefaultPort = `
resource "aws_security_group" "test" {
  name        = "tf-test-security-group"
  description = "tf-test-security-group-descr"
}

resource "aws_security_group_rule" "test" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = aws_elasticache_cluster.test.port
  protocol          = "tcp"
  security_group_id = aws_security_group.test.id
  to_port           = aws_elasticache_cluster.test.port
  type              = "ingress"
}

resource "aws_elasticache_cluster" "test" {
  cluster_id           = "foo-cluster"
  engine               = "redis"
  engine_version       = "5.0.4"
  node_type            = "cache.t2.micro"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis5.0"
}
`

func testAccAWSElasticacheClusterConfig_AZMode_Memcached(rName, azMode string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  az_mode           = "%[2]s"
  cluster_id        = "%[1]s"
  engine            = "memcached"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = 1
  port              = 11211
}
`, rName, azMode)
}

func testAccAWSElasticacheClusterConfig_AZMode_Redis(rName, azMode string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  az_mode           = "%[2]s"
  cluster_id        = "%[1]s"
  engine            = "redis"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = 1
  port              = 6379
}
`, rName, azMode)
}

func testAccAWSElasticacheClusterConfig_EngineVersion_Memcached(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = "%[1]s"
  engine            = "memcached"
  engine_version    = "%[2]s"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = 1
  port              = 11211
}
`, rName, engineVersion)
}

func testAccAWSElasticacheClusterConfig_EngineVersion_Redis(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = "%[1]s"
  engine            = "redis"
  engine_version    = "%[2]s"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = 1
  port              = 6379
}
`, rName, engineVersion)
}

func testAccAWSElasticacheClusterConfig_NodeType_Memcached(rName, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = "%[1]s"
  engine            = "memcached"
  node_type         = "%[2]s"
  num_cache_nodes   = 1
  port              = 11211
}
`, rName, nodeType)
}

func testAccAWSElasticacheClusterConfig_NodeType_Redis(rName, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = "%[1]s"
  engine            = "redis"
  node_type         = "%[2]s"
  num_cache_nodes   = 1
  port              = 6379
}
`, rName, nodeType)
}

func testAccAWSElasticacheClusterConfig_NumCacheNodes_Redis(rName string, numCacheNodes int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = "%[1]s"
  engine            = "redis"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = %[2]d
  port              = 6379
}
`, rName, numCacheNodes)
}

func testAccAWSElasticacheClusterConfig_ReplicationGroupID_AvailabilityZone(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_description = "Terraform Acceptance Testing"
  replication_group_id          = "%[1]s"
  node_type                     = "cache.t3.medium"
  number_cache_clusters         = 1
  port                          = 6379

  lifecycle {
    ignore_changes = [number_cache_clusters]
  }
}

resource "aws_elasticache_cluster" "test" {
  availability_zone    = data.aws_availability_zones.available.names[0]
  cluster_id           = "%[1]s1"
  replication_group_id = aws_elasticache_replication_group.test.id
}
`, rName)
}

func testAccAWSElasticacheClusterConfig_ReplicationGroupID_Replica(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_description = "Terraform Acceptance Testing"
  replication_group_id          = "%[1]s"
  node_type                     = "cache.t3.medium"
  number_cache_clusters         = 1
  port                          = 6379

  lifecycle {
    ignore_changes = [number_cache_clusters]
  }
}

resource "aws_elasticache_cluster" "test" {
  count                = %[2]d
  cluster_id           = "%[1]s${count.index}"
  replication_group_id = aws_elasticache_replication_group.test.id
}
`, rName, count)
}

func testAccAWSElasticacheClusterConfig_Memcached_FinalSnapshot(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1

  final_snapshot_identifier = %[1]q
}
`, rName)
}

func testAccAWSElasticacheClusterConfig_Redis_FinalSnapshot(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "redis"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1

  final_snapshot_identifier = %[1]q
}
`, rName)
}
