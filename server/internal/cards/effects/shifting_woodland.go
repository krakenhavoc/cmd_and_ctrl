package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shifting Woodland — Land:
//
//	"This land enters tapped unless you control a Forest.
//	 {T}: Add {G}.
//	 Delirium — {2}{G}{G}: This land becomes a copy of target permanent
//	 card in your graveyard until end of turn. Activate only if there
//	 are four or more card types among cards in your graveyard."
//
// The enters-tapped clause and the mana ability are the ordinary
// checkland shape (EntersTappedUnless, one Forest, not "another" —
// this land has no printed subtype of its own to exclude).
//
// S49 sandbox simplification: the Delirium activated ability isn't
// implemented. "Becomes a copy of target permanent card ... until end
// of turn" is a CR 707 layer-1 copy effect applied to a permanent
// that is ALREADY on the battlefield, with its own duration and
// reversion at cleanup — a different shape from the ETB
// "enters as a copy" replacement the catalog has (EntersAsCopyOf),
// which only ever fires once, at entry, and never reverts. There's no
// primitive for a temporary in-place copy yet, so the ability is left
// off entirely; the land only ever taps for {G}.
func init() {
	Register(Spec{
		OracleID:     "7c2a4fe5-43e8-4e20-bef2-0278d18afc4b",
		Name:         "Shifting Woodland",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The \"Delirium — {2}{G}{G}: This land becomes a copy of target permanent card in your graveyard until end of turn\" ability isn't implemented — the land can only tap for green mana."},
		Replacements: []game.ReplacementEffect{EntersTappedUnless(otherLandsWithSubtypeAtLeast("forest", 1))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
