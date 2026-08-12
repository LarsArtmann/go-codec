package codec_test

// Shared test fixtures for the external test package (codec_test). These mirror
// the constants in testdata_test.go (package codec) so both test packages can
// use the same fixture names.

const (
	testName      = "Alice"
	testEmail     = "alice@example.com"
	testUserName  = "alice"
	testGreeting  = "hello"
	testValue     = "test"
	testMapKey    = "key"
	testMapVal    = "value"
	testField     = "name"
	testFieldE    = "email"
	testCount     = "count"
	testBob       = "Bob"
	testEventType = "user.created"
	testAlpha     = "alpha"
	testBeta      = "beta"
	testGamma     = "gamma"
	testNested    = "nested"
	testJSON      = "json"
	testCBOR      = "cbor"
)
