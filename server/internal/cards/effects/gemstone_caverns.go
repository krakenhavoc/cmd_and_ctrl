package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gemstone Caverns — Legendary Land:
//
//	"If this card is in your opening hand and you're not the starting
//	 player, you may begin the game with Gemstone Caverns on the
//	 battlefield with a luck counter on it. If you do, exile a card
//	 from your hand.
//	 {T}: Add {C}. If Gemstone Caverns has a luck counter on it,
//	 instead add one mana of any color."
//
// Two sentences, and only the second has a seam to stand on. The mana
// ability is one `ProducedFunc` reading the source's counters: a luck
// counter turns the colourless slot into a five-colour pick, and
// "instead" means it REPLACES the {C} rather than adding to it.
//
// The counter is read at ACTIVATION, not declared once, so a Caverns
// that gains a luck counter some other way (a Proliferate, a Vampire
// Hexmage removing one) answers correctly on the next tap.
//
// DECLARED SIMPLIFICATION, weaker than printed: the opening-hand
// start is not offered. "Begin the game with this on the battlefield"
// is a CR 103.6 pre-game action and the engine has no pre-game
// decision point — no mulligan-time prompt, no way to put a permanent
// onto the battlefield before turn one, and no exile of the
// compensating card. So this Caverns is a colourless land you play
// off a land drop, and it will only ever make coloured mana if
// something else puts a luck counter on it. The free turn-zero
// acceleration is the card's whole reason for being played, and its
// absence is the direction #259 allows; the seam is a pre-game
// opening-hand action, recorded on the batch issue.
func init() {
	Register(Spec{
		OracleID:     "c0adbddc-b070-4c5f-afe0-0474c72a9251",
		Name:         "Gemstone Caverns",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Starting the game with this on the battlefield from your opening hand isn't implemented — you play it as an ordinary land, and without a luck counter it only makes {C}.",
		},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: gemstoneCavernsProduced,
			Label:        "Add {C}, or one mana of any color with a luck counter",
		}},
	})
}

// gemstoneCavernsProduced is the "instead" clause: a five-colour pick
// while the land has a luck counter, plain {C} otherwise.
func gemstoneCavernsProduced(g *game.Game, _, source uuid.UUID) string {
	if c, ok := g.LookupCardForEffect(source); ok && c.Counters["luck"] > 0 {
		return "{W|U|B|R|G}"
	}
	return "{C}"
}
