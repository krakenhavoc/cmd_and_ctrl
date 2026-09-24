package coverage

import (
	"embed"
	"regexp"
	"sync"
)

// printed.go — the oracle fixture as a value the running server can
// read, and the caveat phrase matcher as a function other packages can
// call.
//
// Both exist for internal/roadmap (ADR 0092), which answers "which
// fully automated cards use this mechanic" at runtime. Its most honest
// answer for a mechanic with no declared Spec slot — cascade, ward,
// hideaway — is "a card whose PRINTED text has the keyword and whose
// card file declares it complete", and that needs the printed text in
// the binary rather than in a test's working directory.
//
// The embed costs about half a megabyte. Nothing serves it: the
// roadmap publishes card names only, never oracle text (ADR 0092
// Decision 1).

//go:embed testdata/oracle/*.json
var oracleFixtureFS embed.FS

var (
	printedOnce sync.Once
	printed     map[string]OracleCard
)

// PrintedOracle returns the checked-in oracle fixture, keyed by base
// oracle ID, parsed once per binary. The map is shared: callers must
// not modify it.
//
// It is the same directory LoadOracleFixture reads, so the same
// freshness guard covers it (TestOracleFixtureIsCurrent, nightly). A
// card missing from the fixture simply has no printed text here; it is
// never an error.
func PrintedOracle() map[string]OracleCard {
	printedOnce.Do(func() {
		out, _, err := LoadOracleFixtureFS(oracleFixtureFS, OracleFixtureDir)
		if err != nil {
			// The fixture is compiled in and parsed by the oracle
			// tests on every CI run, so this cannot happen on a
			// build that passed them. An empty map errs towards
			// "no examples found", never towards a false one.
			out = map[string]OracleCard{}
		}
		printed = out
	})
	return printed
}

// PhraseMatcher compiles phrases into the predicate the caveat guards
// use: a case-insensitive, word-bounded match of any one of them. It
// is the one definition of "this caveat names that phrase", shared by
// the curated mechanic table here and by internal/roadmap, so the two
// cannot disagree about which caveats a phrase reaches.
func PhraseMatcher(phrases []string) func(string) bool {
	res := make([]*regexp.Regexp, 0, len(phrases))
	for _, p := range phrases {
		res = append(res, regexp.MustCompile(`(?i)\b`+regexp.QuoteMeta(p)+`\b`))
	}
	return func(s string) bool {
		for _, re := range res {
			if re.MatchString(s) {
				return true
			}
		}
		return false
	}
}
