package gluaaws

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	lua "github.com/yuin/gopher-lua"
)

func Loader(L *lua.LState) int {
	mod := L.NewTable()
	L.SetFuncs(mod, map[string]lua.LGFunction{
		"listEC2Instances":             listEC2Instances,
		"createCloudfrontInvalidation": createCloudfrontInvalidation,
		"uploadToS3":                   uploadToS3,
		"listS3Files":                  listS3Files,
		"downloadFromS3":               downloadFromS3,
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
	copy(items, paths)

	// Create invalidation request
	input := &cloudfront.CreateInvalidationInput{
		DistributionId: &distributionID,
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: &reference,
			Paths: &types.Paths{
				Quantity: aws.Int32(int32(len(items))),
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

func uploadToS3(L *lua.LState) int {
	region := L.CheckString(1)
	profile := L.CheckString(2)
	bucket := L.CheckString(3)
	key := L.CheckString(4)
	filePath := L.CheckString(5)

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

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	// Open file for reading
	file, err := os.Open(filePath)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer file.Close()

	// Upload file to S3
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   file,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LBool(true))
	return 1
}

func listS3Files(L *lua.LState) int {
	region := L.CheckString(1)
	profile := L.CheckString(2)
	bucketPath := L.CheckString(3)

	// Parse bucket and path
	var bucket, prefix string
	for i, c := range bucketPath {
		if c == ':' && i+1 < len(bucketPath) && bucketPath[i+1] == '/' {
			bucket = bucketPath[:i]
			prefix = bucketPath[i+2:] // Skip the ':/'
			break
		}
	}

	if bucket == "" {
		// No prefix format found, assume the whole string is the bucket name
		bucket = bucketPath
		prefix = ""
	}

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

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	// List objects in bucket with prefix
	resp, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Create result table
	resultTable := L.NewTable()
	for _, item := range resp.Contents {
		if item.Key != nil {
			resultTable.Append(lua.LString(*item.Key))
		}
	}

	L.Push(resultTable)
	return 1
}

func downloadFromS3(L *lua.LState) int {
	region := L.CheckString(1)
	profile := L.CheckString(2)
	bucket := L.CheckString(3)
	key := L.CheckString(4)
	destPath := L.CheckString(5)

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

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	// Get object from S3
	resp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer resp.Body.Close()

	// Create destination file
	file, err := os.Create(destPath)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer file.Close()

	// Copy S3 object content to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LBool(true))
	return 1
}
