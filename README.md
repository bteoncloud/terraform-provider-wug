# Terraform WhatsUp Gold Provider

This is the repository for the Terraform WhatsUp Gold Provider, which one can use
with Terraform to work with Veeam server.

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

The provider is useful to add a device to WUG
### Example
```hcl
# Configure the Veeam Provider
provider "wug" {
  url = "http://ipaddress:9644/api/v1"
  user = "${var.user}"
  password = "${var.password}"
}

resource  "wug_device" "my_vm"{
  name = "VM-WINDOWS-01" # display name on WUG
  options = "l2" # set of options for applying the template (either l2 or basic)

  device_type = "Windows 2016 Server"
  snmp_oid = "1.3.6.1.4.1.311.1.1.3.1.2"
  primary_role = "Windows Server"
  subroles = [
    "Windows",
    "Windows Infrastructure",
    "Windows Server"
  ]
  os = "Windows Server 2016"
  brand = "VMware, Inc."

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

  credential {
    type = "SNMP"
    name = "Boostv2"
  }

  active_monitor {
    name = "Ping" # monitor name
    argument = ""
    comment = "mon ping"
    critical = true # is critical
    polling_order = 10
  }

  performance_monitor {
    name = "Disk"
  }
}

```



# Building The Provider

**NOTE:** Unless you are [developing][7] or require a pre-release bugfix or feature,
you will want to use the officially released version of the provider (see [the
section above][8]).
[7]: #developing-the-provider
[8]: #using-the-provider

# Cloning the Project
First, you will want to clone the repository to `$GOPATH/src/github.com/terraform-providers/terraform-provider-veeam`:

```
mkdir -p $GOPATH/src/github.com/terraform-providers
cd $GOPATH/src/github.com/terraform-providers
git clone git@github.com:terraform-providers/terraform-provider-veeam

```

# Running the Build
After the clone has been completed, you can enter the provider directory and build the provider.

```
cd $GOPATH/src/github.com/terraform-providers/terraform-provider-veeam
make build

```

# Installing the Local Plugin
After the build is complete, copy the `terraform-provider-veeam` binary into the same path as your `terraform` binary, and re-run `terraform init`.

After this, your project-local `.terraform/plugins/ARCH/lock.json` (where `ARCH` matches the architecture of your machine) file should contain a SHA256 sum that matches the local plugin. Run `shasum -a 256` on the binary to verify the values match.

# Developing the Provider
If you wish to work on the provider, you'll first need [Go][9] installed on your machine (version 1.9+ is required). You'll also need to correctly setup a [GOPATH][10], as well as adding `$GOPATH/bin` to your `$PATH`.

[9]:https://golang.org/
[10]:https://golang.org/doc/code.html#GOPATH
See [Building the Provider][11] for details on building the provider.
[11]: #building-the-provider


# Testing the Provider
NOTE: Testing the VEEAM provider is currently a complex operation as it requires having a VEEAM backup Server to test against.

# Configuring Environment Variables
Most of the tests in this provider require a comprehensive list of environment variables to run. See the individual `*_test.go` files in the [`veeam/`][12] directory for more details. The next section also describes how you can manage a configuration file of the test environment variables.

[12]: https://github.com/GSLabDev/terraform-provider-veeam/tree/master/veeam 

# Running the Acceptance Tests

After this is done, you can run the acceptance tests by running:

```
$ make testacc

```

If you want to run against a specific set of tests, run `make testacc` with the `TESTARGS` parameter containing the run mask as per below:

```
make testacc TESTARGS="-run=TestAccAddVMToJob_Basic"

```

This following example would run all of the acceptance tests matching `TestAccAddVMToJob_Basic`
