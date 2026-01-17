# ConfigKit Implementation Plan

## User Stories

### Epic: Core Configuration Management

#### Story 1: Basic Configuration Definition

**As a** developer  
**I want to** define configuration parameters with fluent API  
**So that** I can easily set up configuration for my service

**Acceptance Criteria:**

- Can define string, int64, float64, bool, duration, URL, string slice, and int64 slice types
- Can set environment variable names and command-line flags
- Can set default values
- Can mark fields as required
- Can add descriptions

**Implementation Tasks:**

- [x] Create `types.go` with ValueType constants and parseValue function
- [x] Create `definition.go` with Definition struct and DefinitionBuilder
- [x] Implement all type setter methods (String(), Int64(), etc.)
- [x] Implement source setters (Env(), Flag())
- [x] Implement behavior setters (Required(), Default(), Description())

#### Story 2: Configuration Value Resolution

**As a** developer  
**I want to** resolve configuration values from multiple sources  
**So that** I can use flags (highest priority), environment variables, or defaults

**Acceptance Criteria:**

- Command-line flags override environment variables
- Environment variables override defaults
- Proper error handling when required values are missing
- Support for optional values

**Implementation Tasks:**

- [x] Create `config.go` with Config struct
- [x] Implement resolveValue() method with priority logic
- [x] Add flag registration and parsing
- [x] Handle missing required values

#### Story 3: Type-Safe Value Retrieval

**As a** developer  
**I want to** retrieve configuration values with type safety  
**So that** I can avoid runtime type errors

**Acceptance Criteria:**

- Generic Get[T]() function for type-safe retrieval
- Panic with clear error messages for wrong types
- Convenience methods (GetString(), GetInt64(), etc.)
- Has() method to check if value exists

**Implementation Tasks:**

- [x] Create `get.go` with generic Get function
- [x] Implement type checking and panic messages
- [x] Add convenience typed getter methods
- [x] Implement Has() and Keys() methods

### Epic: Validation System

#### Story 4: Built-in Validation Rules

**As a** developer  
**I want to** validate configuration values with common rules  
**So that** I can ensure data integrity and provide helpful error messages

**Acceptance Criteria:**

- Required field validation
- Numeric range validation (min, max, range)
- String length validation (minLength, maxLength, lengthRange)
- Regular expression validation
- OneOf validation for enums
- Duration range validation
- Array item count validation (minItems, maxItems, itemsRange)

**Implementation Tasks:**

- [x] Create `validation.go` with Validation struct
- [x] Implement all built-in validation functions
- [x] Add validation to DefinitionBuilder
- [x] Integrate validation into resolveValue()

#### Story 5: Custom Validation

**As a** developer  
**I want to** add custom validation rules  
**So that** I can enforce domain-specific constraints

**Acceptance Criteria:**

- Custom() method on DefinitionBuilder
- Accepts validation name and check function
- Integrates with existing validation system

**Implementation Tasks:**

- [x] Add Custom() method to DefinitionBuilder
- [x] Ensure custom validations run with built-in ones

### Epic: Security & Secrets

#### Story 6: Secure Secret Handling

**As a** developer  
**I want to** handle sensitive configuration values securely  
**So that** secrets are protected in memory and properly masked

**Acceptance Criteria:**

- Secret() method to mark fields as sensitive
- Integration with memguard for secure storage
- Secrets never stored in regular values map
- Automatic masking in error messages
- Destroy() method to wipe secrets from memory

**Implementation Tasks:**

- [x] Add memguard dependency
- [x] Create `secret.go` with Secret struct and SecretStore
- [x] Implement secure storage and retrieval
- [x] Add secret masking in errors
- [x] Implement Destroy() cleanup

#### Story 7: Secret Access Control

**As a** developer  
**I want to** access secrets through dedicated API  
**So that** I can't accidentally expose them

**Acceptance Criteria:**

- GetSecret() method for secure access
- Generic Get() panics for secret keys
- Secret size and existence checks
- Safe byte/string access with cleanup

**Implementation Tasks:**

- [x] Modify Get() to panic for secrets
- [x] Implement GetSecret() method
- [x] Add Secret utility methods (Size(), IsSet(), etc.)

### Epic: Error Handling & UX

#### Story 8: Beautiful Error Formatting

**As a** developer  
**I want to** see clear, formatted error messages  
**So that** I can quickly identify and fix configuration issues

**Acceptance Criteria:**

- Collected all errors, not just first failure
- Boxed format with clear visual hierarchy
- Source information (env, flag, default, none)
- Masked secret values in errors
- Total error count

**Implementation Tasks:**

- [x] Create `errors.go` with ConfigError struct
- [x] Implement formatErrors() with box drawing
- [x] Add secret masking function
- [x] Modify Process() to collect all errors

#### Story 9: Help Generation

**As a** developer  
**I want to** generate help documentation automatically  
**So that** users can see all available configuration options

**Acceptance Criteria:**

- GenerateHelp() method
- Lists all configuration keys with types
- Shows environment variables and flags
- Indicates required fields and secrets
- Shows defaults and descriptions
- Lists validation rules

**Implementation Tasks:**

- [x] Implement GenerateHelp() method
- [x] Format output with clear sections
- [x] Include all relevant metadata

#### Story 10: Configuration Dumping

**As a** developer  
**I want to** dump current configuration for debugging  
**So that** I can see what values are actually being used

**Acceptance Criteria:**

- Dump() method returning map[string]string
- Secrets masked with byte count
- Shows "[not set]" for missing values
- Useful for debugging and logging

**Implementation Tasks:**

- [x] Implement Dump() method
- [x] Handle secret masking
- [x] Handle unset values

### Epic: Advanced Features

#### Story 11: Array Support

**As a** developer  
**I want to** handle array configuration values  
**So that** I can configure lists like CORS origins

**Acceptance Criteria:**

- StringSlice and Int64Slice types
- Configurable delimiter (default comma)
- Empty string results in empty array
- Whitespace trimming
- Array validation (minItems, maxItems)

**Implementation Tasks:**

- [x] Add slice types to ValueType
- [x] Implement slice parsing in parseValue()
- [x] Add slice type setters to DefinitionBuilder
- [x] Add array validation rules

#### Story 12: URL Validation

**As a** developer  
**I want to** validate URL configuration values  
**So that** I can ensure proper URL format

**Acceptance Criteria:**

- URL type with validation
- Checks for scheme and host
- Stores as string after validation
- Clear error messages for invalid URLs

**Implementation Tasks:**

- [x] Add URL type to ValueType
- [x] Implement URL parsing and validation
- [x] Add URL() method to DefinitionBuilder

### Epic: Integration & Testing

#### Story 13: Complete Example Implementation

**As a** developer  
**I want to** see a complete working example  
**So that** I can understand how to use the library

**Acceptance Criteria:**

- Full AuthForge example with all features
- Server, database, JWT, CORS, OAuth, rate limiting configs
- Shows both secret and non-secret usage
- Demonstrates error handling

**Implementation Tasks:**

- [x] Create comprehensive example
- [x] Include all major features
- [x] Add comments explaining usage

#### Story 14: Comprehensive Testing

**As a** developer  
**I want to** have confidence the library works correctly  
**So that** I can rely on it in production

**Acceptance Criteria:**

- Unit tests for all major functions
- Integration tests for complete workflows
- Error condition testing
- Secret handling tests

**Implementation Tasks:**

- [x] Create config_test.go
- [x] Test all validation rules
- [x] Test secret handling
- [x] Test error formatting
- [x] Test edge cases

### Epic: Documentation & Release

#### Story 15: Documentation

**As a** developer  
**I want to** have complete documentation  
**So that** I can easily understand and use the library

**Acceptance Criteria:**

- README with quick start
- API documentation
- Examples for all features
- Installation instructions

**Implementation Tasks:**

- [x] Write comprehensive README
- [x] Document all public APIs
- [x] Add usage examples
- [x] Include installation guide

#### Story 16: Package Publishing

**As a** developer  
**I want to** publish the library to Go package registry  
**So that** others can easily use it

**Acceptance Criteria:**

- Proper go.mod setup
- Version tagging
- Clean API surface
- No breaking changes in v1

**Implementation Tasks:**

- [x] Set up go.mod with correct module path
- [x] Ensure clean public API
- [ ] Add version tagging strategy
- [ ] Prepare for release

### Epic: Command System Foundation

#### Story 17: Command Definition API

**As a** developer  
**I want to** define commands with fluent API  
**So that** I can create CLI interfaces consistent with configuration

**Acceptance Criteria:**

- Command() method on Config returning CommandBuilder
- Func(), ShortHelp(), LongHelp(), Aliases() methods
- Config() method for command-specific configuration
- SubCommand() method for nested commands
- Same fluent API as configuration definition

**Implementation Tasks:**

- [x] Create Command struct and CommandBuilder
- [x] Implement Command() method on Config
- [x] Add all CommandBuilder methods
- [x] Create CommandConfig wrapper
- [x] Add command storage to Config struct

#### Story 18: Command Context & Execution

**As a** developer  
**I want to** execute commands with context  
**So that** I can access merged configuration and command info

**Acceptance Criteria:**

- CommandContext struct with Args, Config, Command, SubCommand
- Execute() method on Config for command line parsing
- Command function signature with CommandContext
- Command parsing and routing logic
- Support for command aliases

**Implementation Tasks:**

- [x] Create CommandContext struct
- [x] Implement Execute() method with argument parsing
- [x] Add command routing and alias resolution
- [x] Create command execution flow
- [x] Handle command not found cases

### Epic: Command Configuration

#### Story 19: Command-Specific Configuration

**As a** developer  
**I want to** define configuration for specific commands  
**So that** each command can have its own settings

**Acceptance Criteria:**

- CommandConfig with same Define() API as global
- Command-specific flags and environment variables
- Configuration merging (global + command)
- Flag override warnings for conflicts
- Priority resolution for command vs global

**Implementation Tasks:**

- [x] Implement CommandConfig struct
- [x] Add command-specific definition storage
- [x] Create config merging logic
- [x] Add override warning system
- [x] Implement priority resolution

#### Story 20: Configuration Priority System

**As a** developer  
**I want to** have clear configuration priority rules  
**So that** I understand which values take precedence

**Acceptance Criteria:**

- Command flags > Command env > Global flags > Global env > Command defaults > Global defaults
- Clear documentation of priority order
- Warnings for flag overrides
- Consistent behavior across all sources

**Implementation Tasks:**

- [x] Implement priority resolution algorithm
- [x] Add override detection and warnings
- [x] Document priority rules
- [x] Test priority scenarios
- [x] Handle edge cases

### Epic: Help System

#### Story 21: Auto-Generated Help

**As a** developer  
**I want to** auto-generate help from definitions  
**So that** help is always consistent and up-to-date

**Acceptance Criteria:**

- ShowGlobalHelp() method listing all commands
- ShowCommandHelp() method for specific command help
- --help flag support for commands
- Help includes options, descriptions, defaults
- Help shows subcommands and aliases

**Implementation Tasks:**

- [x] Create help generation system
- [x] Implement ShowGlobalHelp() method
- [x] Implement ShowCommandHelp() method
- [x] Add --help flag handling
- [x] Format help output clearly

#### Story 22: Smart Command Suggestions

**As a** developer  
**I want to** get suggestions for unknown commands  
**So that** I can quickly find the right command

**Acceptance Criteria:**

- Levenshtein distance algorithm for suggestions
- "Did you mean?" messages for unknown commands
- Threshold for suggestion quality
- No suggestions when no close matches exist

**Implementation Tasks:**

- [x] Implement Levenshtein distance algorithm
- [x] Create suggestion logic
- [x] Add "Did you mean?" formatting
- [x] Set appropriate distance threshold
- [x] Test with various command names

### Epic: Advanced Command Features

#### Story 23: Subcommands

**As a** developer  
**I want to** create nested subcommands  
**So that** I can organize complex CLI interfaces

**Acceptance Criteria:**

- SubCommand() method on CommandBuilder
- Nested command configuration
- Subcommand aliases
- Help for subcommands
- Execution routing to subcommands

**Implementation Tasks:**

- [x] Add subcommand storage to Command struct
- [x] Implement SubCommand() method
- [x] Create subcommand routing logic
- [x] Add subcommand help generation
- [x] Handle subcommand aliases

#### Story 24: Command Error Handling

**As a** developer  
**I want to** have clear error messages for commands  
**So that** I can quickly fix command issues

**Acceptance Criteria:**

- Beautiful formatted errors for command config
- Missing required flag messages with usage
- Unknown command errors with suggestions
- Command execution error handling
- Consistent error formatting with config errors

**Implementation Tasks:**

- [x] Extend error formatting for commands
- [x] Add missing required flag handling
- [x] Implement unknown command errors
- [x] Handle command execution errors
- [x] Test all error scenarios

### Epic: Command Integration & Testing

#### Story 25: Complete Command Example

**As a** developer  
**I want to** see a complete command system example  
**So that** I can understand how to use all command features

**Acceptance Criteria:**

- Complete example with global + command config
- Multiple commands with aliases
- Subcommands demonstration
- Help system usage
- Error handling examples

**Implementation Tasks:**

- [x] Create comprehensive command example
- [x] Include all major features
- [x] Add detailed comments
- [x] Show best practices
- [ ] Document edge cases

#### Story 26: Command System Testing

**As a** developer  
**I want to** have comprehensive command system tests  
**So that** I can rely on command functionality

**Acceptance Criteria:**

- Unit tests for all command components
- Integration tests for complete workflows
- Error condition testing
- Help system testing
- Configuration merging tests

**Implementation Tasks:**

- [x] Create command_test.go
- [x] Test command definition and execution
- [x] Test configuration merging
- [x] Test help generation
- [x] Test error handling and suggestions

### Epic: Configuration Files

#### Story 27: Configuration File Support

**As a** developer  
**I want to** load configuration from JSON, YAML, and TOML files  
**So that** I can manage configuration in files with highest priority

**Acceptance Criteria:**

- Support for JSON, YAML, and TOML file formats
- LoadFile() and LoadFiles() methods
- LoadFromEnv() method to get file path from environment
- Configuration files have highest priority in source chain
- Environment-specific configuration support

**Implementation Tasks:**

- [x] Add file format parsers (JSON, YAML, TOML)
- [x] Create files.go with file loading logic
- [x] Implement LoadFile() and LoadFiles() methods
- [x] Add LoadFromEnv() method
- [x] Implement environment-specific overrides

#### Story 28: Configuration File Priority & Merging

**As a** developer  
**I want to** have clear priority rules for configuration files  
**So that** I understand which values take precedence

**Acceptance Criteria:**

- Config files > Flags > Env Vars > Defaults priority
- Multiple file support with override logic
- Environment-specific configuration merging
- Clear documentation of priority order

**Implementation Tasks:**

- [x] Implement file priority resolution
- [x] Add multi-file merging logic
- [x] Create environment-specific override system
- [x] Document priority rules
- [x] Test priority scenarios

### Epic: Command Middleware

#### Story 29: Command Middleware System

**As a** developer  
**I want to** add middleware to commands  
**So that** I can implement common functionality like logging and authentication

**Acceptance Criteria:**

- Middleware interface and types
- UseMiddleware() method for global middleware
- Command-specific middleware support
- Middleware execution order (registration order)
- Built-in middleware (logging, auth, error handling)

**Implementation Tasks:**

- [x] Create middleware.go with middleware types (in command.go)
- [x] Implement UseMiddleware() method
- [x] Add command-specific middleware support
- [x] Create built-in middleware implementations
- [x] Implement middleware execution chain

#### Story 30: Advanced Middleware Features

**As a** developer  
**I want to** have advanced middleware capabilities  
**So that** I can apply middleware conditionally and share context

**Acceptance Criteria:**

- Conditional middleware for specific commands
- Subcommand-specific middleware
- Context sharing between middleware
- Middleware for command groups
- Custom middleware examples

**Implementation Tasks:**

- [x] Add UseMiddlewareForCommands() method
- [x] Add UseMiddlewareForSubcommands() method
- [x] Implement context sharing in CommandContext
- [x] Create middleware examples and documentation
- [x] Test advanced middleware scenarios

## Implementation Order

1. **Phase 1: Core Foundation**
   - Story 1: Basic Configuration Definition
   - Story 2: Configuration Value Resolution
   - Story 3: Type-Safe Value Retrieval

2. **Phase 2: Validation**
   - Story 4: Built-in Validation Rules
   - Story 5: Custom Validation

3. **Phase 3: Security**
   - Story 6: Secure Secret Handling
   - Story 7: Secret Access Control

4. **Phase 4: User Experience**
   - Story 8: Beautiful Error Formatting
   - Story 9: Help Generation
   - Story 10: Configuration Dumping

5. **Phase 5: Advanced Features**
   - Story 11: Array Support
   - Story 12: URL Validation

6. **Phase 6: Configuration Files**
   - Story 27: Configuration File Support
   - Story 28: Configuration File Priority & Merging

7. **Phase 7: Command Foundation**
   - Story 17: Command Definition API
   - Story 18: Command Context & Execution

8. **Phase 8: Command Configuration**
   - Story 19: Command-Specific Configuration
   - Story 20: Configuration Priority System

9. **Phase 9: Command Middleware**
   - Story 29: Command Middleware System
   - Story 30: Advanced Middleware Features

10. **Phase 10: Help System**
    - Story 21: Auto-Generated Help
    - Story 22: Smart Command Suggestions

11. **Phase 11: Advanced Commands**
    - Story 23: Subcommands
    - Story 24: Command Error Handling

12. **Phase 12: Quality Assurance**
    - Story 13: Complete Example Implementation
    - Story 14: Comprehensive Testing
    - Story 25: Complete Command Example
    - Story 26: Command System Testing

13. **Phase 13: Release**
    - Story 15: Documentation
    - Story 16: Package Publishing

## Success Criteria

- All user stories completed with acceptance criteria met
- Comprehensive test coverage (>90%)
- Clean, documented API
- Working example demonstrating all features
- Ready for production use in MicroSaaS services
