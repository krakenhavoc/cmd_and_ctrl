package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seasoned Consultant — Creature — Human Detective {1}{W}, 1/3 (Murders
// at Karlov Manor, #4469):
//
//	"Whenever you attack with three or more creatures, this creature
//	 gets +2/+0 until end of turn."
//
// Meddling Youths' trigger (the EventAttack events of one declaration by
// you, settled as a batch, one trigger when there are at least three; the
// Consultant need not be attacking), with the pump on itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ee64b00-6619-451e-bb48-10af107dcb2f",
		Name:         "Seasoned Consultant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WithAtBatchEnd(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Seasoned Consultant — gets +2/+0 (you attacked with three or more creatures)",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.BoostUntilEOT{Target: ctx.Item.SourceCardID, Power: 2,
						Label: "Seasoned Consultant — +2/+0 until end of turn"}.Apply(ctx)
				}))),
				effects.ThatManyEventsAtLeast(3)),
		},
	})
}
