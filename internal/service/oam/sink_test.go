package oam_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/oam"
	"github.com/aws/aws-sdk-go-v2/service/oam/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfoam "github.com/hashicorp/terraform-provider-aws/internal/service/oam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccObservabilityAccessManagerSink_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var sink oam.GetSinkOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_oam_sink.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkExists(resourceName, &sink),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "oam", regexp.MustCompile(`sink/+.`)),
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

func TestAccObservabilityAccessManagerSink_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var sink oam.GetSinkOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_oam_sink.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkExists(resourceName, &sink),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfoam.ResourceSink(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccObservabilityAccessManagerSink_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var sink oam.GetSinkOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_oam_sink.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkExists(resourceName, &sink),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "oam", regexp.MustCompile(`sink/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSinkConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkExists(resourceName, &sink),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "oam", regexp.MustCompile(`sink/+.`)),
				),
			},
			{
				Config: testAccSinkConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkExists(resourceName, &sink),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "oam", regexp.MustCompile(`sink/+.`)),
				),
			},
		},
	})
}

func testAccCheckSinkDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAccessManagerClient()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_oam_sink" {
			continue
		}

		input := &oam.GetSinkInput{
			Identifier: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetSink(ctx, input)
		if err != nil {
			var nfe *types.ResourceNotFoundException
			if errors.As(err, &nfe) {
				return nil
			}
			return err
		}

		return create.Error(names.ObservabilityAccessManager, create.ErrActionCheckingDestroyed, tfoam.ResNameSink, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckSinkExists(name string, sink *oam.GetSinkOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ObservabilityAccessManager, create.ErrActionCheckingExistence, tfoam.ResNameSink, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ObservabilityAccessManager, create.ErrActionCheckingExistence, tfoam.ResNameSink, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAccessManagerClient()
		ctx := context.Background()
		resp, err := conn.GetSink(ctx, &oam.GetSinkInput{
			Identifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.ObservabilityAccessManager, create.ErrActionCheckingExistence, tfoam.ResNameSink, rs.Primary.ID, err)
		}

		*sink = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAccessManagerClient()
	ctx := context.Background()

	input := &oam.ListSinksInput{}
	_, err := conn.ListSinks(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSinkConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_oam_sink" "test" {
  name = %[1]q
}
`, rName)
}

func testAccSinkConfigTags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_oam_sink" "test" {
  name = %[1]q
  
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccSinkConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_oam_sink" "test" {
  name = %[1]q
  
  tags = {
    %[2]q = %[3]q
	%[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
