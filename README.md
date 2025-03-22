# glua-aws

This is a lua module for aws services

## Usage

To use the AWS module in your Lua code:

```lua
local aws = require("aws")

-- List EC2 instances by specifying region and profile
local instances = aws.listEC2Instances("us-west-2", "default")

-- Iterate through instances
for _, instance in ipairs(instances) do
    -- Access instance properties
    print("Instance ID:", instance.instanceId)
    print("Instance Type:", instance.instanceType)
    print("State:", instance.state)
    
    -- Access IP addresses (if available)
    if instance.privateIp then
        print("Private IP:", instance.privateIp)
    end
    if instance.publicIp then
        print("Public IP:", instance.publicIp)
    end
    
    -- Access instance tags
    for tagKey, tagValue in pairs(instance.tags) do
        print("Tag:", tagKey, "=", tagValue)
    end
end

-- Create CloudFront invalidation
local invalidation = aws.createCloudfrontInvalidation("us-east-1", "default", "EDFDVBD6EXAMPLE", {"/images/*", "/index.html"})

-- Access invalidation properties
print("Invalidation ID:", invalidation.id)
print("Status:", invalidation.status)

-- Access invalidation paths
print("Paths to invalidate:")
for _, path in ipairs(invalidation.paths) do
    print("  -", path)
end

-- Upload a file to S3
local uploadSuccess = aws.uploadToS3("us-west-2", "default", "my-bucket", "path/to/remote-file.txt", "path/to/local-file.txt")
print("Upload success:", uploadSuccess)

-- List files in an S3 bucket
local files = aws.listS3Files("us-west-2", "default", "my-bucket")
local files = aws.listS3Files("us-west-2", "default", "my-bucket:/prefix/")
print("Files in bucket:")
for _, fileName in ipairs(files) do
    print("  -", fileName)
end

-- Download a file from S3
local downloadSuccess = aws.downloadFromS3("us-west-2", "default", "my-bucket", "path/to/remote-file.txt", "path/to/local-destination.txt")
print("Download success:", downloadSuccess)
```

Each instance object contains the following properties:
- `instanceId`: The EC2 instance ID
- `instanceType`: The instance type (e.g., t2.micro)
- `state`: Current instance state
- `privateIp`: Private IP address (if available)
- `publicIp`: Public IP address (if available)
- `tags`: Table containing instance tags

Each invalidation object contains the following properties:
- `id`: The CloudFront invalidation ID
- `status`: Current invalidation status
- `paths`: Table containing paths that are being invalidated

S3 functions return:
- `uploadToS3`: Boolean indicating success or failure
- `listS3Files`: Table of strings containing file names in the bucket
- `downloadFromS3`: Boolean indicating success or failure

## Tasks

### tag

Add a new tag and create latest tag if not exists

Inputs: MSG

```
TAG_VERSION=$(convco version -b)
git tag -a v$TAG_VERSION -m "$MSG"
git push origin v$TAG_VERSION
git tag -f latest -m "$MSG"
git push -f origin latest
```

### changelog

Generate changelog

```
convco changelog > CHANGELOG.md
git add CHANGELOG.md
git commit -m "Update CHANGELOG.md"
```

### refresh-dependencies

Refresh dependencies

```
go mod tidy
```