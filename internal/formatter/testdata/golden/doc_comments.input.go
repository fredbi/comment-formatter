package golden

// Foo does the thing. It also does another thing here. And a third sentence which is fairly long and goes on for a little while past eighty columns surely. Then a final short one. One more sentence to push past the dense prose threshold here. And yet another.
func Foo() {}

// bar is a helper
func bar() {}

// Baz processes input
// across two lines without a period
func Baz() {}

//go:generate stringer -type=Kind
type Kind int

// Qux supports e.g. abbreviations and version v1.2.3 without splitting. The second sentence stays separate though.
func Qux() {}
