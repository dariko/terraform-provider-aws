package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSpotInstanceRequest_basic(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckAWSSpotInstanceRequestAttributes(&sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "terminate"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behaviour", "terminate"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_tags(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestTagsConfig1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccAWSSpotInstanceRequestTagsConfig2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSpotInstanceRequestTagsConfig1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_KeyName(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	keyPairResourceName := "aws_key_pair.test"

	rName := acctest.RandomWithPrefix("tf-acc-test")
	publicKey, _, err := acctest.RandSSHKeyPair(testAccDefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_KeyName(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttrPair(resourceName, "key_name", keyPairResourceName, "key_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_withLaunchGroup(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_withLaunchGroup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckAWSSpotInstanceRequestAttributes(&sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_group", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_withBlockDuration(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_withBlockDuration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckAWSSpotInstanceRequestAttributes(&sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "block_duration_minutes", "60"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_vpc(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfigVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckAWSSpotInstanceRequestAttributes(&sir),
					testAccCheckAWSSpotInstanceRequestAttributesVPC(&sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_validUntil(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := testAccAWSSpotInstanceRequestValidUntil(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfigValidUntil(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckAWSSpotInstanceRequestAttributes(&sir),
					testAccCheckAWSSpotInstanceRequestAttributesValidUntil(&sir, validUntil),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_withoutSpotPrice(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_withoutSpotPrice(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckAWSSpotInstanceRequestAttributesCheckSIRWithoutSpot(&sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_SubnetAndSGAndPublicIpAddress(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_SubnetAndSGAndPublicIpAddress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckAWSSpotInstanceRequest_InstanceAttributes(&sir, rName),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_NetworkInterfaceAttributes(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_SubnetAndSGAndPublicIpAddress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckAWSSpotInstanceRequest_InstanceAttributes(&sir, rName),
					testAccCheckAWSSpotInstanceRequest_NetworkInterfaceAttributes(&sir),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_getPasswordData(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"

	rName := acctest.RandomWithPrefix("tf-acc-test")
	publicKey, _, err := acctest.RandSSHKeyPair(testAccDefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_getPasswordData(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttrSet(resourceName, "password_data"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment", "password_data", "get_password_data"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_disappears(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSpotInstanceRequest(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSSpotInstanceRequestValidUntil(t *testing.T) string {
	return testAccAWSSpotInstanceRequestTime(t, "12h")
}

func testAccAWSSpotInstanceRequestTime(t *testing.T, duration string) string {
	n := time.Now().UTC()
	d, err := time.ParseDuration(duration)
	if err != nil {
		t.Fatalf("err parsing time duration: %s", err)
	}
	return n.Add(d).Format(time.RFC3339)
}

func testAccCheckAWSSpotInstanceRequestDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_spot_instance_request" {
			continue
		}

		req := &ec2.DescribeSpotInstanceRequestsInput{
			SpotInstanceRequestIds: []*string{aws.String(rs.Primary.ID)},
		}

		resp, spotErr := conn.DescribeSpotInstanceRequests(req)
		// Verify the error is what we expect
		if !isAWSErr(spotErr, "InvalidSpotInstanceRequestID.NotFound", "") {
			return spotErr
		}
		var s *ec2.SpotInstanceRequest
		if spotErr == nil {
			for _, sir := range resp.SpotInstanceRequests {
				if sir.SpotInstanceRequestId != nil && *sir.SpotInstanceRequestId == rs.Primary.ID {
					s = sir
				}
				continue
			}
		}
		if s == nil {
			// not found
			continue
		}
		if aws.StringValue(s.State) == ec2.SpotInstanceStateCancelled || aws.StringValue(s.State) == ec2.SpotInstanceStateClosed {
			// Requests stick around for a while, so we make sure it's cancelled
			// or closed.
			continue
		}

		// Now check if the associated Spot Instance was also destroyed
		instanceID := rs.Primary.Attributes["spot_instance_id"]
		instance, instErr := resourceAwsInstanceFindByID(conn, instanceID)
		if instErr == nil {
			if instance != nil {
				return fmt.Errorf("instance %q still exists", instanceID)
			}
			continue
		}

		// Verify the error is what we expect
		if !isAWSErr(instErr, "InvalidInstanceID.NotFound", "") {
			return instErr
		}
	}

	return nil
}

func testAccCheckAWSSpotInstanceRequestExists(
	n string, sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS subscription with that ARN exists")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		params := &ec2.DescribeSpotInstanceRequestsInput{
			SpotInstanceRequestIds: []*string{&rs.Primary.ID},
		}
		resp, err := conn.DescribeSpotInstanceRequests(params)

		if err != nil {
			return err
		}

		if v := len(resp.SpotInstanceRequests); v != 1 {
			return fmt.Errorf("Expected 1 request returned, got %d", v)
		}

		*sir = *resp.SpotInstanceRequests[0]

		return nil
	}
}

func testAccCheckAWSSpotInstanceRequestAttributes(
	sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *sir.SpotPrice != "0.050000" {
			return fmt.Errorf("Unexpected spot price: %s", *sir.SpotPrice)
		}
		if *sir.State != ec2.SpotInstanceStateActive {
			return fmt.Errorf("Unexpected request state: %s", *sir.State)
		}
		if *sir.Status.Code != "fulfilled" {
			return fmt.Errorf("Unexpected bid status: %s", *sir.State)
		}
		return nil
	}
}

func testAccCheckAWSSpotInstanceRequestAttributesValidUntil(
	sir *ec2.SpotInstanceRequest, validUntil string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sir.ValidUntil.Format(time.RFC3339) != validUntil {
			return fmt.Errorf("Unexpected valid_until time: %s", sir.ValidUntil.String())
		}
		return nil
	}
}

func testAccCheckAWSSpotInstanceRequestAttributesCheckSIRWithoutSpot(
	sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *sir.State != ec2.SpotInstanceStateActive {
			return fmt.Errorf("Unexpected request state: %s", *sir.State)
		}
		if *sir.Status.Code != "fulfilled" {
			return fmt.Errorf("Unexpected bid status: %s", *sir.State)
		}
		return nil
	}
}

func testAccCheckAWSSpotInstanceRequest_InstanceAttributes(sir *ec2.SpotInstanceRequest, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		instance, err := resourceAwsInstanceFindByID(conn, aws.StringValue(sir.InstanceId))
		if err != nil {
			if isAWSErr(err, "InvalidInstanceID.NotFound", "") {
				return fmt.Errorf("Spot Instance %q not found", aws.StringValue(sir.InstanceId))
			}
			return err
		}

		// If nothing was found, then return no state
		if instance == nil {
			return fmt.Errorf("Spot Instance not found")
		}

		var sgMatch bool
		for _, s := range instance.SecurityGroups {
			// Hardcoded name for the security group that should be added inside the
			// VPC
			if *s.GroupName == rName {
				sgMatch = true
			}
		}

		if !sgMatch {
			return fmt.Errorf("Error in matching Spot Instance Security Group, expected %s, got %s", rName, instance.SecurityGroups)
		}

		return nil
	}
}

func testAccCheckAWSSpotInstanceRequest_NetworkInterfaceAttributes(
	sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		nis := sir.LaunchSpecification.NetworkInterfaces
		if nis == nil || len(nis) != 1 {
			return fmt.Errorf("Expected exactly 1 network interface, found %d", len(nis))
		}

		return nil
	}
}

func testAccCheckAWSSpotInstanceRequestAttributesVPC(
	sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		nis := sir.LaunchSpecification.NetworkInterfaces
		if nis == nil || len(nis) != 1 {
			return fmt.Errorf("Expected exactly 1 network interface, found %d", len(nis))
		}

		ni := nis[0]

		if ni.SubnetId == nil {
			return fmt.Errorf("Expected SubnetId not be non-empty for %s as the instance belongs to a VPC", *sir.InstanceId)
		}
		return nil
	}
}

func TestAccAWSSpotInstanceRequest_InterruptStop(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestInterruptConfig("stop"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "stop"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behaviour", "stop"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_InterruptHibernate(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestInterruptConfig("hibernate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "hibernate"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behaviour", "hibernate"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_InterruptUpdate(t *testing.T) {
	var sir1, sir2 ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestInterruptConfig("hibernate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir1),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "hibernate"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behaviour", "hibernate"),
				),
			},
			{
				Config: testAccAWSSpotInstanceRequestInterruptConfig("terminate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir2),
					testAccCheckSpotInstanceRequestRecreated(&sir1, &sir2),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "terminate"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behaviour", "terminate"),
				),
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_InterruptDeprecated(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestInterruptConfig_Deprecated("hibernate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "hibernate"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behaviour", "hibernate"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_InterruptFixDeprecated(t *testing.T) {
	var sir1, sir2 ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestInterruptConfig_Deprecated("hibernate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir1),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "hibernate"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behaviour", "hibernate"),
				),
			},
			{
				Config: testAccAWSSpotInstanceRequestInterruptConfig("hibernate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir2),
					testAccCheckSpotInstanceRequestNotRecreated(&sir1, &sir2),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "hibernate"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behaviour", "hibernate"),
				),
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_InterruptUpdateFromDeprecated(t *testing.T) {
	var sir1, sir2 ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestInterruptConfig_Deprecated("hibernate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir1),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "hibernate"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behaviour", "hibernate"),
				),
			},
			{
				Config: testAccAWSSpotInstanceRequestInterruptConfig("stop"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(resourceName, &sir2),
					testAccCheckSpotInstanceRequestRecreated(&sir1, &sir2),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "stop"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behaviour", "stop"),
				),
			},
		},
	})
}

func testAccCheckSpotInstanceRequestRecreated(before, after *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.InstanceId), aws.StringValue(after.InstanceId); before == after {
			return fmt.Errorf("Spot Instance (%s) not recreated", before)
		}

		return nil
	}
}

func testAccCheckSpotInstanceRequestNotRecreated(before, after *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.InstanceId), aws.StringValue(after.InstanceId); before != after {
			return fmt.Errorf("Spot Instance (%s/%s) recreated", before, after)
		}

		return nil
	}
}

func testAccAWSSpotInstanceRequestConfig() string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"), `
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  wait_for_fulfillment = true
}
`)
}

func testAccAWSSpotInstanceRequestTagsConfig1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  wait_for_fulfillment = true
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSSpotInstanceRequestTagsConfig2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  wait_for_fulfillment = true
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAWSSpotInstanceRequestConfigValidUntil(rName string, validUntil string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  valid_until          = %[2]q
  wait_for_fulfillment = true

  tags = {
    Name = %[1]q
  }
}
`, rName, validUntil))
}

func testAccAWSSpotInstanceRequestConfig_withoutSpotPrice(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  wait_for_fulfillment = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSSpotInstanceRequestConfig_KeyName(rName, publicKey string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  key_name             = aws_key_pair.test.key_name
  wait_for_fulfillment = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, publicKey))
}

func testAccAWSSpotInstanceRequestConfig_withLaunchGroup(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  wait_for_fulfillment = true
  launch_group         = %[1]q

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSSpotInstanceRequestConfig_withBlockDuration(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type          = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price             = "0.05"
  wait_for_fulfillment   = true
  block_duration_minutes = 60

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSSpotInstanceRequestConfigVPC(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  wait_for_fulfillment = true
  subnet_id            = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSSpotInstanceRequestConfig_SubnetAndSGAndPublicIpAddress(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price                  = "0.05"
  wait_for_fulfillment        = true
  subnet_id                   = aws_subnet.test.id
  vpc_security_group_ids      = [aws_security_group.test.id]
  associate_public_ip_address = true
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSSpotInstanceRequestConfig_getPasswordData(rName, publicKey string) string {
	return composeConfig(
		testAccLatestWindowsServer2016CoreAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.win2016core-ami.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  key_name             = aws_key_pair.test.key_name
  wait_for_fulfillment = true
  get_password_data    = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, publicKey))
}

func testAccAWSSpotInstanceRequestInterruptConfig(interruptionBehavior string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("c5.large", "c4.large"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                            = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type                  = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price                     = "0.07"
  wait_for_fulfillment           = true
  instance_interruption_behavior = %[1]q
}
`, interruptionBehavior))
}

func testAccAWSSpotInstanceRequestInterruptConfig_Deprecated(interruptionBehavior string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("c5.large", "c4.large"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                             = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type                   = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price                      = "0.07"
  wait_for_fulfillment            = true
  instance_interruption_behaviour = %[1]q
}
`, interruptionBehavior))
}
