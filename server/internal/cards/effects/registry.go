package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// registry is the package-level card-effect catalog. Populated at
// init() time by each per-card file calling Register(Spec{...}).
// Reads (Lookup / All) are zero-lock because Go's init() ordering
// serialises writes before any reader touches the map. Post-init
// mutation is disallowed — tests that need to inject specs call
// registerForTest, which is only visible within the test binary.
var registry = map[string]Spec{}

// Register adds a catalog entry. Called once per card at init()
// time. Panics on duplicate OracleID — a collision means a card
// file was copy-pasted without updating the key, which is a logic
// bug we want to surface loudly at server boot rather than
// silently letting one spec win.
//
// Empty OracleID is rejected for the same reason: a spec without
// a key would shadow Lookup() results for the empty string and
// mask bugs in the caller.
func Register(spec Spec) {
	if spec.OracleID == "" {
		panic(fmt.Sprintf("effects.Register: empty OracleID on %q", spec.Name))
	}
	if existing, ok := registry[spec.OracleID]; ok {
		panic(fmt.Sprintf("effects.Register: duplicate OracleID %s (existing %q, new %q)",
			spec.OracleID, existing.Name, spec.Name))
	}
	if spec.Modes != nil {
		if spec.Targets != nil {
			panic(fmt.Sprintf("effects.Register: %q declares both Targets and Modes — put the target clause on the mode", spec.Name))
		}
		targeted := 0
		for _, o := range spec.Modes.Options {
			if o.Targets != nil {
				targeted++
			}
		}
		if spec.Modes.Max > 1 && targeted > 1 {
			panic(fmt.Sprintf("effects.Register: %q has %d targeted modes with Max %d — per-mode target slots are unsupported (S20 sub-PR 4)",
				spec.Name, targeted, spec.Modes.Max))
		}
	}
	// S22: an alternative cost is claimed by name on the wire, so a
	// blank or duplicated key is unaddressable — the cast either
	// can't name it or names two of them. Both are copy-paste
	// mistakes, and both fail loudly at boot rather than as a
	// mysteriously-rejected cast mid-game.
	seenAlt := make(map[string]bool, len(spec.AlternativeCosts))
	for _, ac := range spec.AlternativeCosts {
		if ac.Key == "" {
			panic(fmt.Sprintf("effects.Register: %q declares an alternative cost with no Key", spec.Name))
		}
		if seenAlt[ac.Key] {
			panic(fmt.Sprintf("effects.Register: %q declares two alternative costs keyed %q", spec.Name, ac.Key))
		}
		seenAlt[ac.Key] = true
	}
	// S22: a tap-permanents cost with no pool of legal permanents can
	// never be paid, and one whose extra cost doesn't parse would
	// silently charge nothing — both are copy-paste mistakes in a
	// card file, and both fail at boot rather than mid-game.
	if tc := spec.TapCost; tc != nil {
		if tc.Key == "" || tc.Spec == nil {
			panic(fmt.Sprintf("effects.Register: %q declares a tap cost with no key or no legal permanents — build it with Convoke() or Waterbend()", spec.Name))
		}
		if tc.Extra != "" {
			if _, err := game.ParseCost(tc.Extra); err != nil {
				panic(fmt.Sprintf("effects.Register: %q declares an unparseable tap cost %q: %v", spec.Name, tc.Extra, err))
			}
		}
	}
	registry[spec.OracleID] = spec
}

// Lookup returns the Spec for a given oracle ID. The second return
// is false when the ID is not in the catalog — that's the signal
// for the resolution path to fall back to manual sandbox behaviour.
// Callers MUST check the second return; the zero Spec{} is
// semantically distinct from a real registered spec (no OnResolve,
// no OnETB), which means a missed check would silently apply
// nothing rather than triggering the manual fallback correctly.
func Lookup(oracleID string) (Spec, bool) {
	s, ok := registry[oracleID]
	return s, ok
}

// All returns a snapshot slice of every registered Spec. Order is
// undefined (map iteration). Used by the "auto"-bit serialiser on
// CardView (sub-PR 3) and by tests that want to iterate the whole
// catalog. The returned slice is freshly allocated — callers can
// mutate it freely without leaking into the registry.
func All() []Spec {
	out := make([]Spec, 0, len(registry))
	for _, s := range registry {
		out = append(out, s)
	}
	return out
}

// Has reports whether the catalog knows the given oracle ID.
// Convenience wrapper for the auto-badge bit (sub-PR 3).
func Has(oracleID string) bool {
	_, ok := registry[oracleID]
	return ok
}
