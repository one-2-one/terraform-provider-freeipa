# FreeIPA Provider - Keytab Authentication Example

This example demonstrates how to configure the FreeIPA Terraform provider to use Kerberos/keytab authentication instead of username/password.

## Prerequisites

1. A valid keytab file with credentials for a principal that has permissions to manage FreeIPA resources
2. A properly configured `krb5.conf` file
3. Network access to your FreeIPA server

## Configuration

### Using Explicit Configuration

```hcl
provider "freeipa" {
  host = "ipa.example.com"

  kerberos_enabled   = true
  kerberos_principal = "terraform/ipa.example.com@EXAMPLE.COM"
  kerberos_realm     = "EXAMPLE.COM"
  keytab_path        = "/etc/krb5.keytab"
  krb5_conf_path     = "/etc/krb5.conf"
}
```

### Using Environment Variables

Alternatively, you can use environment variables:

```bash
export FREEIPA_HOST="ipa.example.com"
export FREEIPA_KERBEROS_ENABLED="true"
export FREEIPA_KERBEROS_PRINCIPAL="terraform/ipa.example.com@EXAMPLE.COM"
export FREEIPA_KERBEROS_REALM="EXAMPLE.COM"
export FREEIPA_KEYTAB="/etc/krb5.keytab"
export FREEIPA_KRB5_CONF="/etc/krb5.conf"

terraform apply
```

## Provider Arguments

| Argument              | Environment Variable           | Default              | Description                                                    |
|-----------------------|--------------------------------|----------------------|----------------------------------------------------------------|
| `host`                | `FREEIPA_HOST`                 | -                    | FreeIPA server hostname (required)                             |
| `kerberos_enabled`    | `FREEIPA_KERBEROS_ENABLED`     | `false`              | Enable Kerberos/keytab authentication                          |
| `kerberos_principal`  | `FREEIPA_KERBEROS_PRINCIPAL`   | -                    | Kerberos principal (required when kerberos_enabled is true)    |
| `kerberos_realm`      | `FREEIPA_KERBEROS_REALM`       | -                    | Kerberos realm (required when kerberos_enabled is true)        |
| `keytab_path`         | `FREEIPA_KEYTAB`               | `/etc/krb5.keytab`   | Path to keytab file (required when kerberos_enabled is true)   |
| `krb5_conf_path`      | `FREEIPA_KRB5_CONF`            | `/etc/krb5.conf`     | Path to krb5.conf file                                         |
| `insecure`            | `FREEIPA_INSECURE`             | `false`              | Disable TLS certificate verification                           |

## Creating a Keytab for Terraform

To create a dedicated principal and keytab for Terraform:

```bash
# On the FreeIPA server, create a service principal
ipa service-add terraform/ipa.example.com

# Generate a keytab
ipa-getkeytab -s ipa.example.com \
  -p terraform/ipa.example.com@EXAMPLE.COM \
  -k /path/to/terraform.keytab

# Grant necessary permissions (adjust based on your requirements)
ipa role-add-member "User Administrator" --services=terraform/ipa.example.com
```

## Running the Example

```bash
# Initialize Terraform
terraform init

# Plan your changes
terraform plan

# Apply the configuration
terraform apply
```

## Troubleshooting

### Authentication Fails

1. Verify your keytab is valid:
   ```bash
   klist -kt /path/to/keytab
   ```

2. Test Kerberos authentication manually:
   ```bash
   kinit -kt /path/to/keytab principal@REALM
   klist
   ```

3. Check connectivity to FreeIPA:
   ```bash
   curl -k https://ipa.example.com/ipa/json
   ```

### Permission Denied

Ensure the principal has the necessary RBAC permissions in FreeIPA to perform the operations Terraform is attempting.
