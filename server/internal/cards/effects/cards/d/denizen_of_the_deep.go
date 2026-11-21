package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Denizen of the Deep — Creature — Serpent {6}{U}{U}, 11/11 (Dominaria
// Remastered, #5948):
//
//	"When this creature enters, return each other creature you control to
//	 its owner's hand."
//
// Every other creature its controller controls goes back at once as the
// trigger resolves; the Denizen stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc1bb5b9-5e36-4154-b683-e62f4f66f2b6",
		Name:         "Denizen of the Deep",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Denizen of the Deep — return each other creature you control to its owner's hand", effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				self := ctx.Source()
				return effects.BounceAllMatching{Match: func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
					return c.IsCreature() && c.Controller == caster && c.InstanceID != self
				}}.Apply(ctx)
			}))),
		},
	})
}
