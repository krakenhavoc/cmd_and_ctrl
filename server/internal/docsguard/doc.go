// Package docsguard holds tests that keep the docs/ tree internally
// consistent. It carries no production code and is imported by
// nothing — the tests are the deliverable.
//
// The first guard is ADR numbering. AGENTS.md §4 has asked authors to
// check for a free number since the file was written, and that has
// not been enough: two branches that both grab "the next number"
// against a main that has neither collide, and because the filenames
// differ git reports no conflict. The collision is invisible in the
// diff, invisible in the merge, and only shows up when somebody lists
// the directory. It happened three times before this package existed.
package docsguard
