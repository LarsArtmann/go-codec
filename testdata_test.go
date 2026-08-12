//nolint:testpackage // tests internal helpers
package codec

// Shared test fixtures used across multiple test files. Extracting these as
// named constants satisfies goconst and makes test data intentions explicit.

const (
	testName     = "Alice"
	testEmail    = "alice@example.com"
	testUserName = "alice"
	testGreeting = "hello"
	testValue    = "test"
	testMapKey   = "key"
	testMapVal   = "value"
	testField    = "name"
	testFieldE   = "email"
	testCount    = "count"
)
