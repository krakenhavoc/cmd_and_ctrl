package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scourge Devil — Creature — Devil {4}{R}, 3/3 (Duel Decks: Speed vs.
// Cunning, #5555):
//
//	"When this creature enters, creatures you control get +1/+0 until end of
//	 turn.
//	 Unearth {2}{R}"
//
// Overrun's mass pump as an enters trigger, over the creatures its
// controller controls as it resolves (CR 611.2c), and the keyword's own
// Unearth constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ee65210-547f-4375-8f9a-3484d963339f",
		Name:         "Scourge Devil",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Scourge Devil — creatures you control get +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()),
						Power: 1,
						Label: "Scourge Devil — creatures you control get +1/+0",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{2}{R}")},
	})
}
