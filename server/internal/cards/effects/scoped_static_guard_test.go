package effects

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// scoped_static_guard_test.go is ADR 0041 phase 3's interim lint
// (Decision P6, #1497). A continuous effect from a resolution is a
// data record now — ScopedEffectFor over the game.*Mod vocabulary —
// and a closure-bearing ScopedStatic freezes its table's restore point
// for as long as it lives. So no NEW file may reach the closure API.
//
// The files that still do are listed here, and the list only shrinks:
// tier 3a moves the until-end-of-turn builders and the remaining raw
// abilities onto the record and then deletes StaticForDuration,
// StaticUntilEOT and RegisterScopedStaticForEffect outright — at which
// point this test has nothing left to guard and goes with them.
var legacyScopedStaticCallers = map[string]string{
	// The builders themselves (tier 3a moves their bodies; their
	// signatures, and so every card calling them, stay).
	"until_end_of_turn.go": "BoostUntilEOT, GrantKeywordUntilEOT, StaticUntilEOT",
	"durations.go":         "StaticForDuration itself",
	"restrictions.go":      "RestrictUntilEOT",
	"tribal.go":            "GrantAllCreatureTypesUntilEOT",
	"vehicles.go":          "BecomeArtifactCreature / crew",
	// The raw abilities tier 3a rewrites as mods.
	"cerulean_wisps.go":           "setColors",
	"coercive_recruiter.go":       "addSubtypes",
	"katara_water_tribes_hope.go": "setBasePower / setBaseToughness",
	"pupu_ufo.go":                 "setBasePower",
	"shadowspear.go":              "removeKeywords — pinned at resolution when it moves (owner decision 2)",
	"sudden_spoiling.go":          "loseAllAbilities + setBasePower / setBaseToughness",
}

var legacyScopedStaticCall = regexp.MustCompile(`StaticUntilEOT\{|StaticForDuration\{|RegisterScopedStaticForEffect\(`)

func TestNoNewClosureScopedStatics(t *testing.T) {
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
		if legacyScopedStaticCall.Match(src) {
			using[name] = true
		}
	}
	var fresh, stale []string
	for name := range using {
		if _, ok := legacyScopedStaticCallers[name]; !ok {
			fresh = append(fresh, name)
		}
	}
	for name := range legacyScopedStaticCallers {
		if !using[name] {
			stale = append(stale, name)
		}
	}
	sort.Strings(fresh)
	sort.Strings(stale)
	for _, name := range fresh {
		t.Errorf(`%s registers a closure-bearing scoped static (StaticForDuration,
StaticUntilEOT or RegisterScopedStaticForEffect).

That keeps the table off the restore path for as long as the effect
lives (ADR 0041 phase 3, #1497). Write it as a data record instead:

	ScopedEffectFor{Target: id, Mods: []game.Mod{game.ModifyPTMod(2, 2)},
	    Duration: DurationUntilEndOfTurn(ctx), Label: "…"}.Apply(ctx)

If none of the game.*Mod constructors can say what the card does, that
is a new mod kind (game/scoped_effects.go), not a closure.`, name)
	}
	for _, name := range stale {
		t.Errorf("%s no longer registers a closure-bearing scoped static — delete it from legacyScopedStaticCallers (that is the ratchet working)", name)
	}
}
