package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Explosive Getaway — {3}{R}{W} Sorcery (Aetherdrift, #4464):
//
//	"Exile up to one target artifact or creature. Return it to the
//	 battlefield under its owner's control at the beginning of the next end
//	 step. Explosive Getaway deals 4 damage to each creature."
//
// Touch the Spirit Realm's channel exile (exileTargetsThenScheduleReturn):
// the exile is the first sentence and the damage the last (CR 608.2c), so
// the exiled permanent is not on the battlefield to take it, and it returns
// as a new object at the next end step, under its owner's control (CR 400.7),
// a token not at all (CR 111.8). The damage is dealt to each creature on the
// battlefield as it resolves, the source being the spell. "Up to one" lets
// the spell be cast with no target and still sweep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c844cc98-b8b9-447d-92ee-64f4a4906e7c",
		Name:         "Explosive Getaway",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("up to one target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())).WithCount(0, 1),
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.ExileTargetsThenScheduleReturn(ctx, "Explosive Getaway — return the exiled permanent"); err != nil {
				return err
			}
			return effects.DamageEachMatching(ctx, effects.Creature(), 4)
		},
	})
}
