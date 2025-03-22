package gluaaws

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestListEC2Instances(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("aws", Loader)
	if err := L.DoString(`
		local aws = require("aws")
		
		-- List EC2 instances with region and profile
		local results = aws.listEC2Instances("us-west-2", "default")
		
		-- Basic validation of return value
		assert(type(results) == "table")
		
		-- Print and validate each instance
		for _, instance in ipairs(results) do
			assert(type(instance) == "table")
			assert(type(instance.instanceId) == "string")
			assert(type(instance.tags) == "table")
			
			-- Print instance details for debugging
			print("Instance ID:", instance.instanceId)
			print("Instance Type:", instance.instanceType)
			print("State:", instance.state)
			
			-- Print IP addresses if available
			if instance.privateIp then
				print("Private IP:", instance.privateIp)
			end
			if instance.publicIp then
				print("Public IP:", instance.publicIp)
			end
			
			-- Print all tags
			for tagKey, tagValue in pairs(instance.tags) do
				print("Tag:", tagKey, "=", tagValue)
			end
		end
	`); err != nil {
		t.Error(err)
	}
}

func TestCreateCloudfrontInvalidation(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("aws", Loader)
	if err := L.DoString(`
		local aws = require("aws")
		
		-- Create CloudFront invalidation
		local result = aws.createCloudfrontInvalidation("us-east-1", "default", "EDFDVBD6EXAMPLE", {"/images/*", "/index.html"})
		
		-- Basic validation of return value
		assert(type(result) == "table")
		
		-- Validate structure of returned invalidation data
		print("Invalidation ID:", result.id)
		print("Status:", result.status)
		
		-- Check paths
		assert(type(result.paths) == "table")
		print("Paths to invalidate:")
		for _, path in ipairs(result.paths) do
			assert(type(path) == "string")
			print("  -", path)
		end
	`); err != nil {
		t.Error(err)
	}
}
