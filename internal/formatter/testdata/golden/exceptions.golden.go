// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Fred. This line stays put and keeps no trailing period even though it is long enough to wrap

package golden

// Handler does the thing.
//
// See the reference below for the full protocol details that we follow here.
//
// [spec]: https://example.com/very/long/path/to/specification/document/that/is/quite/long
func Handler() {}

// Options to configure the widget behavior across the whole subsystem here:
//   - first option does something useful and has a fairly long description that goes on
//   - second option
func Configure() {
	//nolint:errcheck
	doThing()
	//go:noinline
}

// Deprecated: use NewThing instead because this old API is going away soon in
// the next release.
func OldThing() {}

// Sections documents the headings.
//
// # Overview
//
// ## Details
//
// ### Sub heading without a trailing period
func Sections() {}

func doThing() {}
