package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yawgmoth's Bargain — Enchantment {4}{B}{B}:
//
//	"Skip your draw step.
//	 Pay 1 life: Draw a card."
//
// Necropotence with the interesting half filed off, and the reason
// SkipYourDrawStep is a shared helper rather than a lambda inside
// necropotence.go: two cards print the identical clause and a second
// hand-written copy is a second place for the seat check to be
// forgotten.
//
// The difference from Necropotence is the difference between a draw
// and an exile, and it is the whole card. This one really draws, so
// the card is in hand immediately, "whenever you draw a card"
// triggers (Underworld Dreams, Psychosis Crawler's cousins) fire,
// and emptying the library to it loses the game at the next
// state-based check per CR 704.5b. Necropotence's ability does none
// of those things.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f7f76f39-a0de-4bda-86b6-0f291892fcec",
		Name:         "Yawgmoth's Bargain",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SkipYourDrawStep()},
		Activated: []ActivatedAbility{{
			Label: "Pay 1 life: Draw a card",
			Cost:  PayLife(1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
