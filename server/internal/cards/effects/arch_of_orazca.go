package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arch of Orazca — Land (EDHREC rank 2395):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 {T}: Add {C}.
//	 {5}, {T}: Draw a card. Activate only if you have the city's
//	 blessing."
//
// A colorless land that draws in a wide deck. "Activate only if you
// have the city's blessing" is the draw's activation condition (CR
// 602.1b, #743).
//
// SANDBOX GAP, weaker than printed — Tendershoot Dryad's and
// Illustrious Wanderglyph's: the city's blessing is a player designation
// that lasts the rest of the game once earned, and the engine has no
// per-player designation to keep it in. The condition therefore reads
// "you control ten or more permanents" live, so a board that shrinks
// below ten shuts the draw where the printed card would keep it open.
// Never stronger: the condition that earns the blessing is the same
// one, and ascend grants it the moment it holds (CR 702.131b).
func init() {
	Register(Spec{
		OracleID:     "3bb518ff-399b-4ce7-b9ad-a1d563dd7792",
		Name:         "Arch of Orazca",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The city's blessing isn't kept once earned — the draw ability works only while you control ten or more permanents."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{5}, {T}: Draw a card. Activate only if you have the city's blessing.",
			Cost:      Plus(ManaCost("{5}"), TapCost()),
			Condition: archOfOrazcaCitysBlessing,
			Effect:    b36DrawOne,
		}},
	})
}

// archOfOrazcaCitysBlessing is the city's blessing, read live.
func archOfOrazcaCitysBlessing(g *game.Game, controller, _ uuid.UUID) bool {
	return b11PermanentsControlled(g, controller) >= 10
}
