// Code generated by "aws/internal/generators/listpages/main.go -function=DescribeCapacityProviders -paginator=NextToken github.com/aws/aws-sdk-go/service/ecs"; DO NOT EDIT.

package lister

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DescribeCapacityProvidersPages(conn *ecs.ECS, input *ecs.DescribeCapacityProvidersInput, fn func(*ecs.DescribeCapacityProvidersOutput, bool) bool) error {
	return DescribeCapacityProvidersPagesWithContext(context.Background(), conn, input, fn)
}

func DescribeCapacityProvidersPagesWithContext(ctx context.Context, conn *ecs.ECS, input *ecs.DescribeCapacityProvidersInput, fn func(*ecs.DescribeCapacityProvidersOutput, bool) bool) error {
	for {
		output, err := conn.DescribeCapacityProvidersWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}