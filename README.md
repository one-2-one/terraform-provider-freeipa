FreeIPA Terraform Provider
==========================

[![Terraform Registry Version](https://img.shields.io/badge/dynamic/json?color=blue&label=registry&query=%24.version&url=https%3A%2F%2Fregistry.terraform.io%2Fv1%2Fproviders%2Fcamptocamp%2Ffreeipa)](https://registry.terraform.io/providers/camptocamp/freeipa)
[![Go Report Card](https://goreportcard.com/badge/github.com/camptocamp/terraform-provider-freeipa)](https://goreportcard.com/report/github.com/camptocamp/terraform-provider-freeipa)
[![Build Status](https://travis-ci.org/camptocamp/terraform-provider-freeipa.svg?branch=master)](https://travis-ci.org/camptocamp/terraform-provider-freeipa)
[![By Camptocamp](https://img.shields.io/badge/by-camptocamp-fb7047.svg)](http://www.camptocamp.com)

This provider adds integration between Terraform and FreeIPA.

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.10


Building The Provider
---------------------

Download the provider source code

```sh
$ go get github.com/camptocamp/terraform-provider-freeipa
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/camptocamp/terraform-provider-freeipa
$ make build
```

Installing the provider
-----------------------

After building the provider, install it using the Terraform instructions for [installing a third party provider](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins).

Example
----------------------

### Password Authentication

```hcl
provider "freeipa" {
  host     = "ipa.example.test"   # or set $FREEIPA_HOST
  username = "admin"              # or set $FREEIPA_USERNAME
  password = "P@S5sw0rd"          # or set $FREEIPA_PASSWORD
}

resource "freeipa_user" "john" {
  name       = "jdoe"
  first_name = "John"
  last_name  = "Doe"
  email_address = ["john.doe@example.test"]
}

resource "freeipa_hostgroup" "web_servers" {
  name        = "web-servers"
  description = "Web server hosts"
}
```

### Keytab Authentication

```hcl
provider "freeipa" {
  host               = "ipa.example.test"     # or set $FREEIPA_HOST
  kerberos_enabled   = true                   # or set $FREEIPA_KERBEROS_ENABLED=true
  kerberos_principal = "terraform/ipa.example.test@EXAMPLE.TEST"  # or set $FREEIPA_KERBEROS_PRINCIPAL
  kerberos_realm     = "EXAMPLE.TEST"         # or set $FREEIPA_KERBEROS_REALM
  keytab_path        = "/etc/krb5.keytab"     # or set $FREEIPA_KEYTAB (default: /etc/krb5.keytab)
  krb5_conf_path     = "/etc/krb5.conf"       # or set $FREEIPA_KRB5_CONF (default: /etc/krb5.conf)
}

resource "freeipa_dns_zone" "example" {
  zone_name = "example.test."
  
  dynamic_updates = true
  default_ttl     = 3600
  
  zone_forwarders = [
    "8.8.8.8",
    "8.8.4.4",
  ]
}

resource "freeipa_sudo_rule" "admin_rule" {
  name            = "admin-full-access"
  description     = "Full sudo access for admins"
  enabled         = true
  usercategory    = "all"
  hostcategory    = "all"
  commandcategory = "all"
}
```

### Creating a Keytab for Terraform

To use keytab authentication, create a service principal for Terraform:

```bash
# On the FreeIPA server
ipa service-add terraform/ipa.example.test

# Generate a keytab
ipa-getkeytab -s ipa.example.test \
  -p terraform/ipa.example.test@EXAMPLE.TEST \
  -k /path/to/terraform.keytab

# Grant necessary permissions
ipa role-add-member "User Administrator" --services=terraform/ipa.example.test
```

Usage
----------------------


Troubleshooting
---------------

If you encounter errors with keytab authentication, especially the "illegal base64 data" error, see the [Troubleshooting Guide](docs/TROUBLESHOOTING.md) for detailed solutions.

Common issues:
- **"illegal base64 data at input byte 4"**: Usually caused by data URI prefixes or invalid base64 format. Ensure your `keytab_base64` is a plain base64 string without any prefixes.
- **Keytab not found**: Verify the keytab file path is correct and the file exists.
- **Authentication failures**: Ensure the keytab matches the principal specified in `kerberos_principal`.


Import
------

DNS records can be imported using the record name and the zone name from <record_name>/<zone_name>/\<type\>

```
$ terraform import freeipa_dns_record.foo foo/example.tld./A
```
