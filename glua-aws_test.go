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
		local results = aws.listEC2Instances("eu-west-1", "thesaiyankiwi")
		
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
		local result = aws.createCloudfrontInvalidation("eu-west-1", "thesaiyankiwi", "EDFDVBD6EXAMPLE", {"/images/*", "/index.html"})
		
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

func TestUploadToS3(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("aws", Loader)
	if err := L.DoString(`
		local aws = require("aws")
		
		-- Create a temp file to upload
		local f = io.open("test_upload.txt", "w")
		f:write("This is test content for S3 upload")
		f:close()
		
		-- Upload file to S3
		local success = aws.uploadToS3("us-west-2", "default", "test-bucket", "test-key.txt", "test_upload.txt")
		
		-- Basic validation
		assert(success == true, "Upload should return true on success")
		
		-- Clean up temp file
		os.remove("test_upload.txt")
		
		print("S3 upload test completed successfully")
	`); err != nil {
		t.Error(err)
	}
}

func TestListS3Files(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("aws", Loader)
	if err := L.DoString(`
		local aws = require("aws")
		
		-- List files in S3 bucket
		local files = aws.listS3Files("us-west-2", "default", "test-bucket")
		
		-- Basic validation of return value
		assert(type(files) == "table", "Should return a table of files")
		
		-- Print and validate files
		print("Files in S3 bucket:")
		for i, fileName in ipairs(files) do
			assert(type(fileName) == "string", "Each file name should be a string")
			print("  -", fileName)
		end
		
		print("S3 list files test completed successfully")
	`); err != nil {
		t.Error(err)
	}
}

func TestDownloadFromS3(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("aws", Loader)
	if err := L.DoString(`
		local aws = require("aws")
		
		-- Download file from S3
		local success = aws.downloadFromS3("us-west-2", "default", "test-bucket", "test-key.txt", "test_download.txt")
		
		-- Basic validation
		assert(success == true, "Download should return true on success")
		
		-- Verify the file exists
		local f = io.open("test_download.txt", "r")
		assert(f ~= nil, "Downloaded file should exist")
		
		-- Read and verify content if needed
		local content = f:read("*all")
		print("Downloaded content length:", #content)
		f:close()
		
		-- Clean up
		os.remove("test_download.txt")
		
		print("S3 download test completed successfully")
	`); err != nil {
		t.Error(err)
	}
}
