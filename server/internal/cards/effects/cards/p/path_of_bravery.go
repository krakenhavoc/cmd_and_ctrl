package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Path of Bravery — Enchantment {2}{W} (Jumpstart, #5510):
//
//	"As long as your life total is greater than or equal to your starting
//	 life total, creatures you control get +1/+1.
//	 Whenever one or more creatures you control attack, you gain life equal
//	 to the number of attacking creatures."
//
// The pump is Righteous Valkyrie's life gate with no margin over the table's
// starting life (ADR 0075); the trigger is Ancestor Dragon's batch trigger
// counting every attacking creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc04bfa9-a8f9-478d-8164-34b1e6c85417",
		Name:         "Path of Bravery",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.LifeGatedPump(effects.YourCreaturesWhileYourLifeAboveStarting(0), 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Path of Bravery — you gain life equal to the number of attacking creatures",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Amount: len(effects.AttackingCreatures(g))}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
