# Troubleshooting Guide

## Keytab Base64 Decoding Error

### Error Message

```
Error: Failed to load keytab
Reason: failed to decode keytab_base64: illegal base64 data at input byte 4
```

### What This Error Means

This error occurs when the `keytab_base64` value contains invalid base64 data. The provider expects a **plain base64-encoded string** with no prefixes, headers, or data URI schemes.

### Common Causes

1. **Double Encoding**: Using `filebase64()` on a file that already contains base64-encoded data (most common!)
2. **Data URI Prefix**: The base64 string includes a data URI prefix like `data:application/octet-stream;base64,`
3. **Incomplete Base64 String**: The base64 string was truncated or copied incorrectly
4. **Invalid Characters**: The string contains non-base64 characters that aren't whitespace
5. **Wrong Encoding**: The keytab file wasn't properly base64-encoded
6. **File Reading Issue**: Using `file()` on the encoded file but the file has trailing newlines or BOM characters

### Solutions

#### Solution 1: Fix Double Encoding (Most Common Issue!)

**If you created an encoded file like this:**
```bash
base64 < keytab.keytab | tr -d '\n' > keytab.encoded
```

**You MUST use `file()` NOT `filebase64()` in Terraform:**

```hcl
# ❌ WRONG - This will double-encode and cause "illegal base64 data at input byte 4"!
keytab_base64 = filebase64("${path.module}/keytab.encoded")

# ✅ CORRECT - Use file() to read the already-encoded string
keytab_base64 = file("${path.module}/keytab.encoded")
```

**Or better yet, use `filebase64()` directly on the original binary keytab file:**

```hcl
# ✅ CORRECT - This encodes the binary keytab file directly
keytab_base64 = filebase64("${path.module}/HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab")
```

**Quick diagnostic:** Check the first few characters of your encoded file:
```bash
head -c 20 HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded
```

- If it starts with `BQIAAABH` or similar (valid base64 characters), use `file()` ✅
- If it starts with binary characters (non-printable), use `filebase64()` ✅

**For your specific case:**
```hcl
provider "freeipa" {
  host               = "your-host"
  kerberos_enabled   = true
  kerberos_principal = "your-principal"
  kerberos_realm     = "your-realm"
  # Option A: Read the encoded file (use file(), not filebase64!)
  keytab_base64      = file("${path.module}/HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded")
  
  # Option B: Or better, encode directly from the original keytab
  # keytab_base64      = filebase64("${path.module}/HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab")
}
```

#### Solution 2: Verify Your Base64 String Format

The `keytab_base64` value must be a **plain base64 string** without any prefixes. It should look like:

```
BQIAAABHAAIADXRlcnJhZm9ybS5leGFtcGxlLmNvbQAAAAEAAAAA...
```

**NOT** like:
```
data:application/octet-stream;base64,BQIAAABHAAIADXRlcnJhZm9ybS5leGFtcGxlLmNvbQAAAAEAAAAA...
```

#### Solution 3: Properly Encode Your Keytab

Use one of these methods to generate a correct base64 string:

**Method A: Using `base64` command (recommended)**
```bash
base64 < /path/to/terraform.keytab | tr -d '\n'
```

**Method B: Using Terraform's `filebase64()` function**
```hcl
variable "freeipa_keytab_base64" {
  description = "Base64 encoded keytab contents"
  type        = string
  sensitive   = true
  default     = filebase64("${path.module}/terraform.keytab")
}

provider "freeipa" {
  host               = "ipa.example.com"
  kerberos_enabled   = true
  kerberos_principal = "terraform/ipa.example.com@EXAMPLE.COM"
  kerberos_realm     = "EXAMPLE.COM"
  keytab_base64      = var.freeipa_keytab_base64
}
```

**Method C: Using environment variable**
```bash
export FREEIPA_KEYTAB_BASE64="$(base64 < /path/to/terraform.keytab | tr -d '\n')"
```

#### Solution 4: Remove Data URI Prefixes

If you copied the base64 string from a data URI or another source that includes a prefix, remove everything before the actual base64 data:

**Before (incorrect):**
```
data:application/octet-stream;base64,BQIAAABHAAIADXRlcnJhZm9ybS5leGFtcGxlLmNvbQAAAAEAAAAA...
```

**After (correct):**
```
BQIAAABHAAIADXRlcnJhZm9ybS5leGFtcGxlLmNvbQAAAAEAAAAA...
```

#### Solution 5: Verify Base64 String Validity

You can verify your base64 string is valid using:

**On Linux/macOS:**
```bash
echo "YOUR_BASE64_STRING" | base64 -d > /tmp/test.keytab
file /tmp/test.keytab
# Should show: /tmp/test.keytab: Kerberos Keytab
```

**Using Python:**
```python
import base64
import sys

try:
    decoded = base64.b64decode(sys.argv[1])
    print(f"Valid base64! Decoded {len(decoded)} bytes")
except Exception as e:
    print(f"Invalid base64: {e}")
```

#### Solution 6: Clean Up Encoded File (Remove Trailing Newlines)

If you saved the base64 string to a file, ensure it has no trailing newlines:

```bash
# Remove any trailing newlines
base64 < terraform.keytab | tr -d '\n' | tr -d '\r' > keytab.encoded

# Verify it's a single line
wc -l keytab.encoded
# Should output: 1 keytab.encoded
```

Then in Terraform:
```hcl
keytab_base64 = file("${path.module}/keytab.encoded")
```

#### Solution 7: Use Keytab Path Instead

If you're having persistent issues with base64 encoding, consider using `keytab_path` instead:

```hcl
provider "freeipa" {
  host               = "ipa.example.com"
  kerberos_enabled   = true
  kerberos_principal = "terraform/ipa.example.com@EXAMPLE.COM"
  kerberos_realm     = "EXAMPLE.COM"
  keytab_path        = "/path/to/terraform.keytab"
  krb5_conf_path     = "/etc/krb5.conf"
}
```

### Common Scenario: Using a Pre-Encoded File

If you've already created an encoded file like this:
```bash
base64 < HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab | tr -d '\n' > HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded
```

**Use `file()` to read it (NOT `filebase64()`):**

```hcl
provider "freeipa" {
  host               = "ipa.example.com"
  kerberos_enabled   = true
  kerberos_principal = "terraform/ipa.example.com@EXAMPLE.COM"
  kerberos_realm     = "EXAMPLE.COM"
  # ✅ CORRECT - file() reads the already-encoded string
  keytab_base64      = file("${path.module}/HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded")
  krb5_conf_path     = "/etc/krb5.conf"
}
```

**Or use environment variable:**
```bash
export FREEIPA_KEYTAB_BASE64="$(cat HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded | tr -d '\n' | tr -d '\r')"
```

### Complete Working Example

Here's a complete example that works correctly:

```hcl
terraform {
  required_providers {
    freeipa = {
      source  = "one-2-one/freeipa"
      version = "1.2.4"
    }
  }
}

# Option 1: Using filebase64() function
variable "freeipa_keytab_base64" {
  description = "Base64 encoded keytab contents"
  type        = string
  sensitive   = true
  default     = filebase64("${path.module}/terraform.keytab")
}

provider "freeipa" {
  host               = "ipa.example.com"
  kerberos_enabled   = true
  kerberos_principal = "terraform/ipa.example.com@EXAMPLE.COM"
  kerberos_realm     = "EXAMPLE.COM"
  keytab_base64      = var.freeipa_keytab_base64
  krb5_conf_path     = "/etc/krb5.conf"
}
```

Or using environment variables:

```bash
# Generate the base64 string correctly
export FREEIPA_KEYTAB_BASE64="$(base64 < terraform.keytab | tr -d '\n')"
export FREEIPA_HOST="ipa.example.com"
export FREEIPA_KERBEROS_ENABLED="true"
export FREEIPA_KERBEROS_PRINCIPAL="terraform/ipa.example.com@EXAMPLE.COM"
export FREEIPA_KERBEROS_REALM="EXAMPLE.COM"
export FREEIPA_KRB5_CONF="/etc/krb5.conf"
```

### Additional Notes

- The provider automatically strips whitespace (spaces, newlines, tabs) from the base64 string, but it's still best practice to provide a compact string
- When both `keytab_path` and `keytab_base64` are set, `keytab_base64` takes precedence
- The `keytab_base64` attribute is marked as sensitive, so it won't appear in Terraform logs
- Ensure your keytab file is valid and not corrupted before encoding

### Quick Diagnostic Commands

Run these commands to diagnose the issue:

**1. Check if your encoded file is valid base64:**
```bash
# Check first few characters (should be base64, not binary)
head -c 50 HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded

# Verify it decodes correctly
base64 -d < HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded > /tmp/test.keytab
file /tmp/test.keytab
# Should show: Kerberos Keytab
```

**2. Check for trailing newlines:**
```bash
# Should show 1 line (or 0 if empty)
wc -l HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded

# Check for carriage returns
od -c HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded | tail -1
```

**3. Verify the original keytab is valid:**
```bash
klist -kt HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab
```

**4. Re-encode if needed:**
```bash
# Clean re-encoding (removes all whitespace)
base64 < HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab | tr -d '\n' | tr -d '\r' > HTTP_atlantis-system.apps.os1.shared.ropot.priv.keytab.encoded
```

### Still Having Issues?

If you continue to experience problems:

1. Verify your keytab file is valid:
   ```bash
   klist -kt /path/to/terraform.keytab
   ```

2. Check that your base64 string doesn't contain any non-base64 characters (should only contain A-Z, a-z, 0-9, +, /, and =)

3. Ensure you're using the correct keytab file for the principal specified in `kerberos_principal`

4. Try using `keytab_path` instead to isolate whether the issue is with base64 encoding or the keytab itself

5. **Most importantly**: Verify you're using `file()` (not `filebase64()`) when reading an already-encoded file

