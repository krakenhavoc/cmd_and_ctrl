package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viscera Seer — Creature — Vampire Wizard {B}, 1/1:
//
//	"Sacrifice a creature: Scry 1."
//
// The cheapest free sacrifice outlet in the format, and the reason it
// is played over strictly stronger payoffs: no mana, no tap, no limit
// per turn. Any number of creatures can go through it at instant speed
// during someone else's turn, which is what makes a persist or
// undying loop work at all.
//
// The scry is almost incidental — the card is in every aristocrats
// list for the outlet — but it does mean a sacrifice loop also fixes
// your draws, and it stacks: sacrifice five creatures, scry five
// times, one card at a time.
//
// It can eat itself (Sacrifice a creature, and it is a creature), so
// the last activation is always available.
func init() {
	Register(Spec{
		OracleID: "f82a4e85-526d-4456-b700-7760043a31be",
		Name:     "Viscera Seer",
		Activated: []ActivatedAbility{{
			Label: "Sacrifice a creature: Scry 1.",
			Cost:  SacrificeACreature(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
