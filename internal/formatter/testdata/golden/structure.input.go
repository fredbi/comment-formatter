package golden

// Config holds settings. The fields are documented below with a list and a code example that must be preserved verbatim.
//
// Example:
//
//	c := Config{Width: 80}
//	_ = c
//
// Supported options:
//
//   - Width sets the column.
//   - MaxLines sets the threshold.
type Config struct {
	Width    int // width in columns, a trailing comment that must not move
	MaxLines int
}

func use() {
	// a plain free-floating comment that is fairly long and certainly exceeds the configured width so it should wrap onto multiple lines nicely
	_ = 0
}
