package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kithkin Zealot — Creature — Kithkin Cleric {1}{W} 1/3:
//
//	"When this creature enters, you gain 1 life for each black and/or red
//	 permanent target opponent controls."
//
// A permanent that is black, red or both counts once. The count is of the
// target opponent's permanents as the trigger resolves (CR 608.2c); a target
// that has left the game makes the trigger do nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7445122d-a943-493c-b7b6-3699cc07db0e",
		Name:         "Kithkin Zealot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Kithkin Zealot — you gain 1 life for each black and/or red permanent target opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						n := effects.CountControlled(g, t.ID, func(c game.Card) bool { return c.HasColorWord("B") || c.HasColorWord("R") })
						return effects.GainLife{Player: item.Controller, Amount: n}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
