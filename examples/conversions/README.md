# CommandKit Type Conversion TDD Test Suite

Tests all `Get[T]()` type conversions across all configuration sources.

## Usage

```bash
cd examples/conversions
go run main.go
```

## What It Tests

**23 types** × **4 sources** = **92 test cases**

### Sources
- **DEFAULT** — only default values, no env/flag/file
- **ENV** — environment variables override defaults
- **FLAG** — command-line flags override defaults
- **FILE** — JSON config file override defaults

### Types Covered
| Category | Types |
|----------|-------|
| Basic | `string`, `bool` |
| Integers | `int64`, `int`, `uint`, `uint8`, `uint16`, `uint32`, `uint64` |
| Floats | `float64`, `float32` |
| Time | `time.Duration`, `time.Time` |
| Special | `url` (string), `uuid.UUID`, `ip` (string), `os.FileMode`, `path` (string) |
| Slices | `[]string`, `[]int64`, `[]int`, `[]float64`, `[]bool` |

## Output Format

```
Source  | Key                       | Expected Type   | Returned Type   | Value                | OK
--------------------------------------------------------------------------------------------------------------
DEFAULT | TEST_STRING               | string          | string          | default_string       | ✅
ENV     | TEST_INT64                | int64           | int64           | 42                   | ✅
FLAG    | TEST_UUID                 | uuid.UUID       | uuid.UUID       | 550e8400-e29b-4...   | ✅
```

## Exit Code

- `0` — all tests passed
- `1` — one or more tests failed


