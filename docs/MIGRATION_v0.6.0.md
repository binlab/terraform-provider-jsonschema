# Migration Guide: v0.5.x to v0.6.0

This guide helps you migrate from v0.5.x to v0.6.0, which introduces breaking changes to support multi-format validation (YAML, TOML).

## Breaking Changes Summary

| Change                | v0.5.x                   | v0.6.0                     |
| --------------------- | ------------------------ | -------------------------- |
| **document field**    | Raw content via `file()` | File path (string)         |
| **Output field**      | `validated`              | `valid_json`               |
| **Supported formats** | JSON, JSON5              | JSON, JSON5, YAML, TOML    |
| **Format detection**  | N/A                      | Auto-detect from extension |
| **Force override**    | N/A                      | `force_filetype` field     |

## Why These Changes?

### File Path API

- **Cleaner syntax**: No need for `file()` wrapper function
- **Format detection**: Automatically detect YAML/TOML from extension
- **Better error messages**: Can report file path in errors
- **Consistency**: Schema already used file path, now document does too

### Renamed Output Field

- **Clarity**: `valid_json` makes it clear the output is JSON format (not boolean)
- **Consistency**: Output is always JSON regardless of input format (YAML/TOML → JSON)

## Step-by-Step Migration

### Step 1: Update Provider Version

```hcl
terraform {
  required_providers {
    jsonschema = {
      source  = "binlab/jsonschema"
-     version = "0.5.0"
+     version = "0.6.0"
    }
  }
}
```

### Step 2: Remove `file()` Wrapper from `document`

**Before (v0.5.x):**

```hcl
data "jsonschema_validator" "config" {
  document = file("${path.module}/config.json")
  schema   = "${path.module}/config.schema.json"
}
```

**After (v0.6.0):**

```hcl
data "jsonschema_validator" "config" {
  document = "${path.module}/config.json"
  schema   = "${path.module}/config.schema.json"
}
```

### Step 3: Rename `validated` to `valid_json`

**Before (v0.5.x):**

```hcl
locals {
  config = jsondecode(data.jsonschema_validator.config.validated)
}

output "validated_config" {
  value = data.jsonschema_validator.config.validated
}
```

**After (v0.6.0):**

```hcl
locals {
  config = jsondecode(data.jsonschema_validator.config.valid_json)
}

output "validated_config" {
  value = data.jsonschema_validator.config.valid_json
}
```

### Step 4: Update All Data Sources

Use find/replace across your `.tf` files:

1. **Remove file() wrapper:**

   - Find: `document = file("`
   - Replace: `document = "`

2. **Rename validated field:**
   - Find: `.validated`
   - Replace: `.valid_json`

## Verification Checklist

After migration, verify:

- [ ] All `document` fields use file paths (no `file()` wrapper)
- [ ] All references changed from `.validated` to `.valid_json`
- [ ] Run `terraform init -upgrade` to download v0.6.0
- [ ] Run `terraform plan` to check for issues
- [ ] Run `terraform apply` in non-production first
- [ ] Verify existing validations still work
- [ ] Check that output values haven't changed

## New Features You Can Use

### YAML Document Validation

Now you can validate YAML files directly:

```hcl
data "jsonschema_validator" "k8s_manifest" {
  document = "${path.module}/deployment.yaml"  # Auto-detected!
  schema   = "${path.module}/k8s-schema.json"
}

locals {
  deployment = jsondecode(data.jsonschema_validator.k8s_manifest.valid_json)
}
```

### TOML Configuration Validation

Validate TOML configuration files:

```hcl
data "jsonschema_validator" "app_config" {
  document = "${path.module}/config.toml"  # Auto-detected!
  schema   = "${path.module}/config-schema.json"
}

locals {
  config = jsondecode(data.jsonschema_validator.app_config.valid_json)
}
```

### Force File Type Override

When file extension doesn't match content:

```hcl
# .txt file containing YAML
data "jsonschema_validator" "custom" {
  document       = "${path.module}/data.txt"
  schema         = "${path.module}/schema.json"
  force_filetype = "yaml"  # Override auto-detection
}

# Force JSON5 parser for relaxed JSON syntax
data "jsonschema_validator" "relaxed" {
  document       = "${path.module}/config.json"
  schema         = "${path.module}/schema.json"
  force_filetype = "json5"  # Allow trailing commas, comments
}
```

## Format Detection Rules

| File Extension  | Detected Format | Parser                            |
| --------------- | --------------- | --------------------------------- |
| `.json`         | JSON            | JSON5 (backward compatible)       |
| `.json5`        | JSON5           | JSON5 (comments, trailing commas) |
| `.yaml`, `.yml` | YAML            | YAML 1.2 (superset of JSON)       |
| `.toml`         | TOML            | TOML v1.0.0                       |
| Other           | JSON5           | Fallback for unknown extensions   |

**Note:** YAML parser can parse JSON (YAML ⊃ JSON), but JSON parser cannot parse YAML.

## Common Issues & Solutions

### Issue: "Error: Invalid function argument"

**Error:**

```
Error: Invalid function argument
  on main.tf line 10, in data "jsonschema_validator" "config":
  10:   document = file("${path.module}/config.json")
```

**Solution:** Remove the `file()` wrapper:

```hcl
- document = file("${path.module}/config.json")
+ document = "${path.module}/config.json"
```

### Issue: "Attribute 'validated' not found"

**Error:**

```
Error: Unsupported attribute
  on main.tf line 15, in locals:
  15:   config = jsondecode(data.jsonschema_validator.config.validated)
```

**Solution:** Rename to `valid_json`:

```hcl
- config = jsondecode(data.jsonschema_validator.config.validated)
+ config = jsondecode(data.jsonschema_validator.config.valid_json)
```

### Issue: YAML file parsed as JSON fails

**Problem:** YAML document being parsed as JSON

**Solution 1:** Use `.yaml` or `.yml` extension (auto-detected)

```hcl
# Rename file: config.json → config.yaml
document = "${path.module}/config.yaml"
```

**Solution 2:** Use `force_filetype` override

```hcl
document       = "${path.module}/config.txt"
force_filetype = "yaml"
```

## Rollback Plan

If you need to rollback:

1. **Restore file() wrapper:**

   ```hcl
   document = file("${path.module}/config.json")
   ```

2. **Restore validated field:**

   ```hcl
   locals {
     config = jsondecode(data.jsonschema_validator.config.validated)
   }
   ```

3. **Downgrade provider version:**

   ```hcl
   terraform {
     required_providers {
       jsonschema = {
         source  = "binlab/jsonschema"
         version = "0.5.0"
       }
     }
   }
   ```

4. **Reinitialize:**
   ```bash
   terraform init -upgrade
   terraform plan
   ```

## Testing Strategy

### 1. Test in Non-Production First

```bash
# In development/staging environment
terraform init -upgrade
terraform plan
terraform apply
```

### 2. Validate Existing Behavior

Ensure validation results haven't changed:

```bash
# Before migration (v0.5.0)
terraform output validated_config > before.json

# After migration (v0.6.0)
terraform output validated_config > after.json

# Compare (should be identical)
diff before.json after.json
```

### 3. Test New Formats

Add YAML/TOML validation to your tests:

```hcl
# Test YAML validation
data "jsonschema_validator" "test_yaml" {
  document = "${path.module}/test.yaml"
  schema   = "${path.module}/test.schema.json"
}

# Test TOML validation
data "jsonschema_validator" "test_toml" {
  document = "${path.module}/test.toml"
  schema   = "${path.module}/test.schema.json"
}
```

## Need Help?

- **Documentation**: See [docs/data-sources/jsonschema_validator.md](data-sources/jsonschema_validator.md)
- **Examples**: Check [examples/](../examples/) directory
- **Issues**: Report issues at https://github.com/binlab/terraform-provider-jsonschema/issues

## Summary

| Task                    | Command/Action                                |
| ----------------------- | --------------------------------------------- |
| Update provider version | Change `version = "0.5.0"` to `"0.6.0"`       |
| Remove file() wrapper   | `document = file("...")` → `document = "..."` |
| Rename output field     | `.validated` → `.valid_json`                  |
| Reinitialize Terraform  | `terraform init -upgrade`                     |
| Test changes            | `terraform plan` then `terraform apply`       |

The migration is straightforward: **remove `file()` wrapper** and **rename `.validated` to `.valid_json`**. The benefits are cleaner syntax and support for YAML/TOML validation!
