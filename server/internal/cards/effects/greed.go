package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greed — Enchantment {3}{B} (EDHREC rank 2172):
//
//	"{B}, Pay 2 life: Draw a card."
//
// The life-for-cards enchantment. One activated ability whose cost
// is a mana component plus a life component (CR 118.8); both are
// validated before either is paid, so a player below 2 life cannot
// activate it (CR 119.4 — a player may pay life only up to their
// life total, so paying the last 2 is legal and lethal, as printed).
// The draw goes on the stack and can be responded to.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ff62220-be95-4901-b8d8-812b9a1a1b0a",
		Name:         "Greed",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{B}, Pay 2 life: Draw a card",
			Cost:  Plus(ManaCost("{B}"), PayLife(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
