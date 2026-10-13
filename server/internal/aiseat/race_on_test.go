//go:build race

package aiseat_test

// raceEnabled reports whether the test binary was built with -race.
// Whole-game tests use it to scale wall-clock budgets: the detector
// makes a game roughly 12x slower (#3147).
const raceEnabled = true
