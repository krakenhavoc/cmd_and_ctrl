package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deeproot Warrior — Creature — Merfolk Warrior {1}{G}, 2/2 (Ixalan):
//
//	"Whenever this creature becomes blocked, it gets +1/+1 until end of turn."
//
// Ichorclaw Myr's trigger for +1/+1: once per blocked attacker however many
// creatures block it (CR 509.3c), and only while it is still the creature
// that was blocked (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2af1930c-a5f2-4e7e-910f-e84458b051ac",
		Name:         "Deeproot Warrior",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self, "Deeproot Warrior — gets +1/+1 until end of turn",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
						return nil
					}
					return effects.BoostUntilEOT{
						Target: ctx.Item.SourceCardID, Power: 1, Toughness: 1,
						Label: "Deeproot Warrior — +1/+1",
					}.Apply(ctx)
				}))),
		},
	})
}
