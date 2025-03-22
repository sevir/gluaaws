package gluaaws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	lua "github.com/yuin/gopher-lua"
)

func Loader(L *lua.LState) int {
	mod := L.NewTable()
	L.SetFuncs(mod, map[string]lua.LGFunction{
		"listEC2Instances":             listEC2Instances,
		"createCloudfrontInvalidation": createCloudfrontInvalidation,
	})
	L.Push(mod)
	return 1
}

func listEC2Instances(L *lua.LState) int {
	region := L.CheckString(1)
	profile := L.CheckString(2)

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Create EC2 client
	client := ec2.NewFromConfig(cfg)

	// Describe instances
	resp, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Create result table
	resultTable := L.NewTable()

	// Process instances
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instanceTable := L.NewTable()

			// Add basic instance information
			instanceTable.RawSetString("instanceId", lua.LString(*instance.InstanceId))
			if instance.InstanceType != "" {
				instanceTable.RawSetString("instanceType", lua.LString(instance.InstanceType))
			}
			if instance.State != nil {
				instanceTable.RawSetString("state", lua.LString(instance.State.Name))
			}
			if instance.PrivateIpAddress != nil {
				instanceTable.RawSetString("privateIp", lua.LString(*instance.PrivateIpAddress))
			}
			if instance.PublicIpAddress != nil {
				instanceTable.RawSetString("publicIp", lua.LString(*instance.PublicIpAddress))
			}

			// Add tags
			tagsTable := L.NewTable()
			for _, tag := range instance.Tags {
				if tag.Key != nil && tag.Value != nil {
					tagsTable.RawSetString(*tag.Key, lua.LString(*tag.Value))
				}
			}
			instanceTable.RawSetString("tags", tagsTable)

			resultTable.Append(instanceTable)
		}
	}

	L.Push(resultTable)
	return 1
}

func createCloudfrontInvalidation(L *lua.LState) int {
	region := L.CheckString(1)
	profile := L.CheckString(2)
	distributionID := L.CheckString(3)
	pathsTable := L.CheckTable(4)

	// Convert Lua table to Go slice
	var paths []string
	pathsTable.ForEach(func(_ lua.LValue, value lua.LValue) {
		paths = append(paths, value.String())
	})

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Create CloudFront client
	client := cloudfront.NewFromConfig(cfg)

	// Create unique reference identifier based on timestamp
	reference := fmt.Sprintf("glua-cf-invalidation-%d", time.Now().Unix())

	// Build paths for invalidation
	items := make([]string, len(paths))
	for i, path := range paths {
		items[i] = path
	}

	// Create invalidation request
	input := &cloudfront.CreateInvalidationInput{
		DistributionId: &distributionID,
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: &reference,
			Paths: &cloudfront.Paths{
				Quantity: int32(len(items)),
				Items:    items,
			},
		},
	}

	// Execute invalidation request
	result, err := client.CreateInvalidation(context.TODO(), input)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Create result table with invalidation details
	resultTable := L.NewTable()

	if result.Invalidation != nil {
		if result.Invalidation.Id != nil {
			resultTable.RawSetString("id", lua.LString(*result.Invalidation.Id))
		}
		if result.Invalidation.Status != nil {
			resultTable.RawSetString("status", lua.LString(*result.Invalidation.Status))
		}

		// Add paths information
		if result.Invalidation.InvalidationBatch != nil &&
			result.Invalidation.InvalidationBatch.Paths != nil {
			pathsResult := L.NewTable()
			if result.Invalidation.InvalidationBatch.Paths.Items != nil {
				for _, path := range result.Invalidation.InvalidationBatch.Paths.Items {
					pathsResult.Append(lua.LString(path))
				}
			}
			resultTable.RawSetString("paths", pathsResult)
		}
	}

	L.Push(resultTable)
	return 1
}
