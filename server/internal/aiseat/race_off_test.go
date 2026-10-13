//go:build !race

package aiseat_test

// raceEnabled reports whether the test binary was built with -race.
// See race_on_test.go.
const raceEnabled = false
