# Terraform WhatsUp Gold Provider

This is the repository for the Terraform WhatsUp Gold Provider, which one can use
with Terraform to work with the WhatsUp Gold (WUG) server.

For general information about Terraform, visit the [official website][3] and the
[GitHub project page][4].

[3]: https://terraform.io/
[4]: https://github.com/hashicorp/terraform


# Using the Provider

The current version of this provider requires Terraform v0.12.9 or higher to
run.

Note that you need to run `terraform init` to fetch the provider before
deploying. Read about the provider split and other changes to TF v0.10.0 in the
official release announcement found [here][4].

[4]: https://www.hashicorp.com/blog/hashicorp-terraform-0-10/


## Full Provider Documentation

### Example
```hcl
# Configure the WUG Provider
provider "wug" {
  url = "http://ipaddress:9644/api/v1"
  user = "${var.user}"
  password = "${var.password}"
}


# Add a device to WUG
resource  "wug_device" "my_vm"{
  name 			= "VM-WINDOWS-01" # display name on WUG
  options 		= "l2" # set of options for applying the template (either l2 or basic)
  action_policy 	= "Mail Policy" # Check the WUG action library to get the exact policy name

  device_type 		= "Windows 2016 Server"
  snmp_oid 		= "1.3.6.1.4.1.311.1.1.3.1.2"
  primary_role 		= "Windows Server"
  os 			= "Windows Server 2016"
  brand 		= "VMware, Inc."
  
  subroles = [
    "Windows",
    "Windows Infrastructure",
    "Windows Server"
  ]

  groups {
    name = "" # Child group name where to put the device
    parents = [ # Child group parents list
      "ROOT",
      "PARENT1",
      "PARENT2"
    ]
  }

  # Interface to declare for the device
  interface {
    default = true # is the interface the default one
    network_name = "LAN"
    network_address = "127.0.0.1"
    poll_using_network_name = true # poll using network name or address
  }

  credential { # Check the WUG credential library to get the exact type and name
    type = "SNMP"
    name = "Boostv2"
  }
}


# Add a monitor to the device
## First, get the monitor IDs
data "wug_monitor" "my_monitor" {
  type 			= "active" # Either "active" or "performance"
  search	 	= "Ping" # Check the WUG monitor library to get the exact monitor name
}

## Second, assign the monitor to the device with its IDs
resource "wug_monitor" "my_monitor" {
  device_id 		= wug_device.my_monitor.id

  type 			= "active" # Either "active" or "performance"
  monitor_type_class_id = data.wug_monitor.my_monitor.class_id
  monitor_type_id 	= data.wug_monitor.my_monitor.id
  monitor_type_name 	= data.wug_monitor.my_monitor.monitor_name	# Re-using "Ping" in this example
  
  # Configure an "active" or "performance" block according to your monitor type
  active {
    critical_order 		= 0
    action_policy_name 		= "Mail Policy" # Check the WUG action library to get the exact policy name
    action_policy_id 		= "" # Only required if 2 monitors have the same name (but shouldn't be needed)
    comment 			= "mon ping"
    argument 			= ""
    polling_interval_seconds 	= 60
    interface_id 		= ""
  }

  performance {
    polling_interval_minutes	= 10
  }
}

```



# Building The Provider

**NOTE:** Unless you are [developing][7] or require a pre-release bugfix or feature,
you will want to use the officially released version of the provider (see [the
section above][8]).
[7]: (#Developing-The-Provider)
[8]: (#Using-the-Provider)

# Cloning the Project
First, you will want to clone the repository to `$GOPATH/src/github.com/terraform-providers/terraform-provider-wug`:

```
mkdir -p $GOPATH/src/github.com/terraform-providers
cd $GOPATH/src/github.com/terraform-providers
git clone git@github.com:terraform-providers/terraform-provider-wug

```

# Running the Build
After the clone has been completed, you can enter the provider directory and build the provider.

```
cd $GOPATH/src/github.com/terraform-providers/terraform-provider-wug
make build

```

# Installing the Local Plugin
After the build is complete, copy the `terraform-provider-wug` binary into the same path as your `terraform` binary, and re-run `terraform init`.

After this, your project-local `.terraform/plugins/ARCH/lock.json` (where `ARCH` matches the architecture of your machine) file should contain a SHA256 sum that matches the local plugin. Run `shasum -a 256` on the binary to verify the values match.

# Developing the Provider
If you wish to work on the provider, you'll first need [Go][9] installed on your machine (version 1.9+ is required). You'll also need to correctly setup a [GOPATH][10], as well as adding `$GOPATH/bin` to your `$PATH`.

[9]:https://golang.org/
[10]:https://golang.org/doc/code.html#GOPATH
See [Building the Provider][11] for details on building the provider.
[11]: (#Building-The-Provider)


# Testing the Provider
NOTE: Testing the WUG provider is currently a complex operation as it requires having a WUG backup Server to test against.

# Configuring Environment Variables
Most of the tests in this provider require a comprehensive list of environment variables to run. Individual `*_test.go` files in the [`wug/`][12] directory have not been built yet. 
Here is an example of how to manage a configuration file of the test environment variables within another provider.

[12]: https://github.com/nerimcloud/terraform-provider-wug

# Running the Acceptance Tests

After this is done, you can run the acceptance tests by running:

```
$ make testacc

```
