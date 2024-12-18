# Vale Testing Cheat Sheet

## Testing Frameworks and Libraries

- Jest
- Go's built-in testing package

## Mocking and Stubbing

- No explicit mocking libraries used
- Manual stubs and fake implementations are created for testing

## Test File Structure

- Test files are named with `_test.go` suffix
- Test functions are prefixed with `Test`

## Common Testing Patterns

### Table-Driven Tests

```go
var cases = []struct {
    input    string
    expected string
}{
    {"input1", "expected1"},
    {"input2", "expected2"},
}

for _, c := range cases {
    result := functionUnderTest(c.input)
    if result != c.expected {
        t.Errorf("expected = %v, got = %v", c.expected, result)
    }
}
```

### Benchmark Tests

```go
func BenchmarkFunction(b *testing.B) {
    for n := 0; n < b.N; n++ {
        functionUnderTest()
    }
}
```

### Helper Functions

```go
func initLinter() (*Linter, error) {
    // Setup code
}

func benchmarkLint(b *testing.B, path string) {
    b.Helper()
    // Benchmark setup and execution
}
```

## Assertions

- Using standard Go testing package assertions
- Custom error messages with `t.Errorf` and `t.Fatalf`

```go
if result != expected {
    t.Errorf("expected = %v, got = %v", expected, result)
}
```

## Test Setup and Teardown

- No explicit setup/teardown functions
- Setup is often done within the test function or using helper functions

## Environment Management

- Using `os.Setenv` for setting environment variables in tests
- `filepath.Join` for cross-platform file path handling

## File System Interactions

- `os.MkdirTemp` for creating temporary directories
- `os.WriteFile` and `os.Remove` for file operations in tests

## Test Coverage

- No explicit mention of coverage tools, but Go's built-in coverage tool is likely used

## Best Practices

1. Use descriptive test names
2. Group related tests in the same file
3. Use table-driven tests for multiple test cases
4. Create helper functions for common setup tasks
5. Use benchmarks for performance-critical code
6. Test both valid and invalid inputs
7. Use `t.Parallel()` for concurrent test execution when possible

## Fake Implementations

- Custom implementations of interfaces or structs are created for testing purposes
- Example: `type globTest struct { query string; match bool }`

## Error Handling in Tests

- Use `t.Fatal` for critical errors that should stop the test
- Use `t.Error` for non-critical failures that allow the test to continue

## Testing Configuration

- Custom configuration structs are often used to set up test environments
- Example: `core.Config` and `core.CLIFlags` structs