---
subcategory: "AppStream"
layout: "aws"
page_title: "AWS: aws_appstream_image_builder"
description: |-
  Provides an AppStream image builder
---

# Resource: aws_appstream_image_builder

Provides an AppStream image builder.

## Example Usage

```terraform
resource "aws_appstream_image_builder" "test_fleet" {
  name                           = "Image Builder Name"
  description                    = "Description of a ImageBuilder"
  display_name                   = "Display name of a ImageBuilder"
  enable_default_internet_access = false
  image_name                     = "AppStream-WinServer2012R2-07-19-2021"
  instance_type                  = "stream.standard.large"

  vpc_config {
    subnet_ids = [aws_subnet.example.id]
  }

  tags = {
    Name = "Example Image Builder"
  }
}
```

## Argument Reference

The following arguments are required:

* `instance_type` - (Required) The instance type to use when launching the image builder.
* `name` - (Required) Unique name for the image builder.

The following arguments are optional:

* `access_endpoint` - (Optional) Set of interface VPC endpoint (interface endpoint) objects. Maximum of 4. See below.
* `appstream_agent_version` - (Optional) The version of the AppStream 2.0 agent to use for this image builder.
* `description` - (Optional) Description to display.
* `display_name` - (Optional) Human-readable friendly name for the AppStream image builder.
* `domain_join_info` - (Optional) Configuration block for the name of the directory and organizational unit (OU) to use to join the image builder to a Microsoft Active Directory domain. See below.
* `enable_default_internet_access` - (Optional) Enables or disables default internet access for the image builder.
* `iam_role_arn` - (Optional) ARN of the IAM role to apply to the image builder.
* `image_arn` - (Optional, Required if `image_name` not provided) ARN of the public, private, or shared image to use.
* `image_name` - (Optional, Required if `image_arn` not provided) Name of the image used to create the image builder.
* `vpc_config` - (Optional) Configuration block for the VPC configuration for the image builder. See below.
* `tags` - (Optional) A map of tags to assign to the instance. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `access_endpoint`

The `access_endpoint` block supports the following arguments:

* `endpoint_type` - (Required) Type of interface endpoint.
* `vpce_id` - (Optional) Identifier (ID) of the VPC in which the interface endpoint is used.

### `domain_join_info`

The `domain_join_info` block supports the following arguments:

* `directory_name` - (Optional) Fully qualified name of the directory (for example, corp.example.com).
* `organizational_unit_distinguished_name` - (Optional) Distinguished name of the organizational unit for computer accounts.

### `vpc_config`

The `vpc_config` block supports the following arguments:

* `security_group_ids` - (Optional) Identifiers of the security groups for the image builder or image builder.
* `subnet_ids` - (Optional) Identifiers of the subnets to which a network interface is attached from the image builder instance or image builder instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the appstream image builder.
* `created_time` -  Date and time, in UTC and extended RFC 3339 format, when the image builder was created.
* `id` - The name of the image builder.
* `state` - State of the image builder. Can be: `PENDING`, `UPDATING_AGENT`, `RUNNING`, `STOPPING`, `STOPPED`, `REBOOTING`, `SNAPSHOTTING`, `DELETING`, `FAILED`, `UPDATING`, `PENDING_QUALIFICATION`
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_appstream_image_builder` can be imported using the `name`, e.g.

```
$ terraform import aws_appstream_image_builder.example imageBuilderExample
```
