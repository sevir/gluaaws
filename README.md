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
```

Each instance object contains the following properties:
- `instanceId`: The EC2 instance ID
- `instanceType`: The instance type (e.g., t2.micro)
- `state`: Current instance state
- `privateIp`: Private IP address (if available)
- `publicIp`: Public IP address (if available)
- `tags`: Table containing instance tags
