package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scale Blessing — Instant {3}{W} (S82 GAP, #5948):
//
//	"Bolster 1, then put a +1/+1 counter on each creature you control with a
//	 +1/+1 counter on it. (To bolster 1, choose a creature with the least
//	 toughness among creatures you control and put a +1/+1 counter on it.)"
//
// Honor's Reward's bolster, then the second sentence is read once the bolster has
// resolved, so the creature just bolstered counts (the creatures with a counter
// are picked as that sentence is reached, CR 608.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7157ec5-9e6a-4771-b3db-6760640d6c6c",
		Name:         "Scale Blessing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: effects.ResolveSteps(
			effects.Bolster{N: 1},
			effects.Instruction(func(ctx *effects.Context) error {
				var ids []uuid.UUID
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == ctx.Controller() && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0 {
						ids = append(ids, c.InstanceID)
					}
				}
				for _, id := range ids {
					if !effects.OnBattlefield(ctx.Game, id) {
						continue
					}
					if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}
