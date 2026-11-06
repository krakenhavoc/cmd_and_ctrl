package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hog-Monkey Rampage — Instant {1}{R/G} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Choose target creature you control and target creature an opponent
//	 controls. Put a +1/+1 counter on the creature you control if it has
//	 power 4 or greater. Then those creatures fight each other. (Each
//	 deals damage equal to its power to the other.)"
//
// Savage Stomp's two clauses, in printed order. The power is read through
// the layers as the spell resolves, before the counter, so a 3-power
// creature gets no counter and a 4-power one fights as a 5. If the
// creature you control is no longer a legal target it gets no counter
// (CR 608.2b); if either target is illegal, neither creature fights
// (CR 701.14b), but a legal creature you control still gets its counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e5df175-5f91-47f7-957c-63bac31f0858",
		Name:         "Hog-Monkey Rampage",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				ctx.Game.RecomputeLayersIfStaleLocked()
				c, ok := ctx.Game.LookupCardForEffect(mine.ID)
				if !ok || c.CurrentPower() < 4 {
					return nil
				}
				return AddCounter{Target: mine.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			Instruction(fightTargetClauses),
		),
	})
}
