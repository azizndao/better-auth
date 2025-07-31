# Transport Package

A modular HTTP transport layer for Better Auth with multi-language validation support.

## File Structure

### 📁 **Core Files**

- **`transport.go`** - Main constructor (`NewDefault()`)
- **`types.go`** - Interface definitions, types, and constants

### 📁 **Feature Modules**

- **`json.go`** - JSON encoding/decoding operations
- **`errors.go`** - Error response handling  
- **`validation.go`** - Validation error processing
- **`localization.go`** - Multi-language translation setup
- **`context.go`** - User context management
- **`auth.go`** - Token extraction and cookie operations

### 📁 **Testing**

- **`transport_test.go`** - Comprehensive test suite

## Architecture Benefits

### ✅ **Modularity**
- **Single Responsibility**: Each file handles one specific concern
- **Easy Navigation**: Find functionality quickly by file name
- **Independent Testing**: Test modules in isolation

### ✅ **Maintainability** 
- **Smaller Files**: No more 415-line monolith
- **Clear Boundaries**: Logical separation of concerns
- **Easy Debugging**: Isolated functionality per file

### ✅ **Extensibility**
- **Add New Features**: Create new files for new functionality
- **Modify Existing**: Change only relevant files
- **Team Development**: Multiple developers can work on different files

## File Responsibilities

| File | Purpose | Key Functions |
|------|---------|---------------|
| `types.go` | Core definitions | `Transport` interface, `ValidationError`, constants |
| `json.go` | JSON operations | `DecodeJSON()`, `DecodeJSONWithLocale()`, `RespondJSON()` |
| `errors.go` | Error responses | `RespondError()`, `RespondValidationError()` |
| `validation.go` | Validation logic | `buildValidationError()`, `isSensitiveField()` |
| `localization.go` | Multi-language | `detectLocale()`, `parseAcceptLanguage()`, translator setup |
| `context.go` | User context | `GetUserContext()`, `SetUserContext()` |
| `auth.go` | Authentication | `ExtractToken()`, `SetCookie()`, `GetCookie()` |

## Usage

The public interface remains unchanged. Use the package exactly as before:

```go
import "better-auth/pkg/transport"

// Create transport instance
t := transport.NewDefault()

// Use all the same methods
err := t.DecodeJSON(req, &model)
t.RespondJSON(w, 200, data)
t.RespondValidationError(w, validationErr)
```

## Supported Languages

- **English**: `en`, `en-US`, `en-GB`, `en-CA`, `en-AU`
- **Spanish**: `es`, `es-ES`, `es-MX`, `es-AR`, `es-CO`  
- **French**: `fr`, `fr-FR`, `fr-CA`, `fr-BE`, `fr-CH`

## Security Features

- **Sensitive Field Detection**: Automatically hides `password`, `secret`, `token`, `key` fields
- **Secure Cookie Defaults**: HttpOnly, SameSite, configurable Secure flags
- **Professional Error Messages**: Using go-playground/validator translations

The modular structure maintains 100% backward compatibility while providing a cleaner, more maintainable codebase.