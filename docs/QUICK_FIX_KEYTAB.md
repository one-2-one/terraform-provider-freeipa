# Quick Fix: "illegal base64 data at input byte 4" Error

## The Problem

You created an encoded file:
```bash
base64 < HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab | tr -d '\n' > HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded
```

And you're getting this error:
```
Error: Failed to load keytab
Reason: failed to decode keytab_base64: illegal base64 data at input byte 4
```

## The Solution

**You're likely using `filebase64()` on the encoded file. Use `file()` instead!**

### ❌ Wrong (causes double encoding):
```hcl
provider "freeipa" {
  # ... other config ...
  keytab_base64 = filebase64("${path.module}/HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded")
}
```

### ✅ Correct (Option 1 - Read the encoded file):
```hcl
provider "freeipa" {
  host               = "your-host"
  kerberos_enabled   = true
  kerberos_principal = "your-principal"
  kerberos_realm     = "your-realm"
  keytab_base64      = file("${path.module}/HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded")
}
```

### ✅ Correct (Option 2 - Encode directly, recommended):
```hcl
provider "freeipa" {
  host               = "your-host"
  kerberos_enabled   = true
  kerberos_principal = "your-principal"
  kerberos_realm     = "your-realm"
  keytab_base64      = filebase64("${path.module}/HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab")
}
```

## Quick Test

Verify your encoded file is valid base64:
```bash
# Check first characters (should be base64 like "BQIAAABH...")
head -c 50 HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded

# Test decoding
base64 -d < HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded > /tmp/test.keytab
file /tmp/test.keytab
# Should show: /tmp/test.keytab: Kerberos Keytab
```

## Rule of Thumb

- **`file()`** = Read text/base64 content from a file
- **`filebase64()`** = Read binary file and encode it to base64

Since you already encoded the file, use `file()`!

