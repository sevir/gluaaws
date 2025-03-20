package gluaaws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	lua "github.com/yuin/gopher-lua"
)

func Loader(L *lua.LState) int {
	mod := L.NewTable()
	L.SetFuncs(mod, map[string]lua.LGFunction{
		"listEC2Instances": listEC2Instances,
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
