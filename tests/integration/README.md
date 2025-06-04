# Integration Test Suite

## Test File Naming Convention
To support external integration testing, all test-related helper files in this module are named using the `test_*.go` prefix instead of the conventional `*_test.go`.

This change is intentional:
In Go, files ending with `*_test.go` are excluded from regular build contexts and only compiled during go test execution. 
As a result, such files are not included when **the package is imported by other modules** — for example, when downstream chains want to reuse our EVM test harness.

To work around this, we rename those files to `test_*.go`, which:
- Keeps them recognizable as test helpers, 
- Ensures they are compiled as part of the package during regular builds, 
- Allows external chains to import and invoke testing utilities like `NewKeeperTestSuite()` or `NewUnitTestNetwork(...)`.

> Note: These files are still test-focused and should not be used in production builds.

