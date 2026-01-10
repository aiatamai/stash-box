# Backend Resolver Code - Restore After Code Generation

## What Happened

The Settings page backend resolver code has been temporarily removed because it references Go types that need to be generated from the GraphQL schema. The project won't compile until gqlgen generates these types.

## Files Saved for Restoration

The complete, working resolver code has been saved in these files:
- `internal/api/resolver.go.PENDING_CODE_GENERATION` - Updated GetConfig resolver
- `internal/api/resolver_mutation_config.go.PENDING_CODE_GENERATION` - New UpdateConfig mutation resolver

## Steps to Complete Implementation

### 1. Generate Go Types from GraphQL Schema

Run code generation in an environment with Go 1.25+:

```bash
# Using Docker (recommended):
docker run --rm -v $(pwd):/app -w /app golang:1.25 sh -c "go generate ./..."

# OR using local Go 1.25+:
make generate-backend
```

This will generate the required types in `internal/models/generated_models.go`:
- `ConfigUpdateInput` struct
- Extended `StashBoxConfig` struct with all new fields

### 2. Restore the Resolver Code

After successful code generation:

```bash
# Restore the GetConfig resolver
cp internal/api/resolver.go.PENDING_CODE_GENERATION internal/api/resolver.go

# Restore the UpdateConfig mutation resolver
cp internal/api/resolver_mutation_config.go.PENDING_CODE_GENERATION internal/api/resolver_mutation_config.go

# Clean up the temporary files
rm internal/api/*.PENDING_CODE_GENERATION
```

### 3. Verify and Build

```bash
# Verify the code compiles
make build

# Run tests if applicable
make test
```

### 4. Commit the Restored Code

```bash
git add internal/api/resolver.go internal/api/resolver_mutation_config.go
git commit -m "Restore Settings page backend resolvers after code generation"
```

## What the Resolvers Do

### resolver.go - GetConfig()
Returns all configuration fields from the config package, mapping them to the GraphQL StashBoxConfig type. Includes:
- General settings (title, host_url, etc.)
- Email configuration
- Image settings
- S3 configuration
- Database settings
- All other config values

### resolver_mutation_config.go - UpdateConfig()
Handles the `updateConfig` mutation:
1. Accepts a `ConfigUpdateInput` with all optional fields
2. Updates viper configuration for any provided fields
3. Writes changes to `stash-box-config.yml`
4. Reloads configuration
5. Returns the updated config

## Current State

✅ GraphQL schema updated with all fields
✅ Frontend fully implemented and validated
✅ Resolver code written and saved
⚠️ Resolver code temporarily removed (waiting for code generation)
⚠️ Backend won't compile until types are generated and code is restored

## Questions?

See `SETTINGS_IMPLEMENTATION.md` in the project root for full documentation.
