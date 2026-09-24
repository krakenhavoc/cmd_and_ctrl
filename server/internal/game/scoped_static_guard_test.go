package game

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// scoped_static_guard_test.go is the engine half of ADR 0041 phase 3's
// interim lint (#1497; the catalog half is
// cards/effects/scoped_static_guard_test.go). A continuous effect from
// a resolution is a data record now — RegisterScopedEffectForEffect
// over the *Mod vocabulary — and a closure-bearing ScopedStatic keeps
// its table off the restore path for as long as it lives. So no NEW
// engine file may register one.
//
// The list only shrinks. Tier 3a moves prowess onto the record and
// then deletes the closure registry, and this test with it.
var legacyScopedStaticEngineCallers = map[string]string{
	"scoped_statics.go": "the legacy registry itself",
	"prowess.go":        "prowess's +1/+1 until end of turn — tier 3a",
}

var legacyScopedStaticEngineCall = regexp.MustCompile(`registerScopedStaticLocked\(|RegisterScopedStaticForEffect\(`)

func TestNoNewEngineClosureScopedStatics(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	using := map[string]bool{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if legacyScopedStaticEngineCall.Match(src) {
			using[name] = true
		}
	}
	var fresh, stale []string
	for name := range using {
		if _, ok := legacyScopedStaticEngineCallers[name]; !ok {
			fresh = append(fresh, name)
		}
	}
	for name := range legacyScopedStaticEngineCallers {
		if !using[name] {
			stale = append(stale, name)
		}
	}
	sort.Strings(fresh)
	sort.Strings(stale)
	for _, name := range fresh {
		t.Errorf(`%s registers a closure-bearing scoped static (registerScopedStaticLocked or
RegisterScopedStaticForEffect). That keeps the table off the restore
path for as long as the effect lives (ADR 0041 phase 3, #1497). Register
a ScopedEffect instead — RegisterScopedEffectForEffect with the *Mod
constructors in scoped_effects.go — and if none of them can say it,
that is a new mod kind, not a closure.`, name)
	}
	for _, name := range stale {
		t.Errorf("%s no longer registers a closure-bearing scoped static — delete it from legacyScopedStaticEngineCallers (that is the ratchet working)", name)
	}
}
