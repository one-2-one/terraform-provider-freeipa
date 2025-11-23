terraform {
  required_providers {
    freeipa = {
      source  = "one-2-one/freeipa"
      version = "~> 1.1"
    }
  }
}

# Example 1: Keytab authentication with explicit paths
provider "freeipa" {
  host = "ipa.example.com"

  kerberos_enabled   = true
  kerberos_principal = "terraform/ipa.example.com@EXAMPLE.COM"
  kerberos_realm     = "EXAMPLE.COM"
  keytab_path        = "/etc/krb5.keytab"
  krb5_conf_path     = "/etc/krb5.conf"

  # Optional: disable TLS verification for testing
  insecure = false
}

# Example resource using keytab authentication
resource "freeipa_user" "example" {
  uid       = "testuser"
  givenname = "Test"
  sn        = "User"
}
