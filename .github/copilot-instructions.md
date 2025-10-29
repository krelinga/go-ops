## Tests

When generating tests, use the following style:
- Test functions should be named `Test<FunctionName>`.
- When generating tests for `<filename>.go`, the generated test functions should be placed in a file named `<filename>_test.go`.
- Test functions should be placed in the `<package_name>_test` package.
- Prefer table-drive tests for functions with multiple test cases.
- Prefer subtests (t.Run) for grouping related test cases.
- Use descriptive names for test cases to clarify their purpose.
- Ensure tests are isolated and do not depend on external state.
