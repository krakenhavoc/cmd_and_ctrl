package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghirapur Grand Prix — Plane — Avishkar (punk, #4870):
//
//	"When you planeswalk here, all players start their engines! (If you
//	 have no speed, it starts at 1. It increases once on each of your turns
//	 when an opponent loses life. Max speed is 4.)
//	 At the beginning of your end step, put X +1/+1 counters on target
//	 creature you control, where X is your speed.
//	 When you planeswalk away from Ghirapur Grand Prix, each player with
//	 the highest speed among players creates three Treasure tokens."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// Starting engines is a verb every seat does in turn order (CR 702.179b); a
// player who already has speed keeps it. The end-step ability counts the
// planar controller's speed as it resolves (a player with none puts zero
// counters, CR 702.179f). The planeswalk-away ability (CR 901.11b) reads the
// speeds as it resolves: every player still in the game whose speed ties the
// highest creates three Treasures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6515e8d6-26ff-4716-a62d-a5847d0933e7",
		Name:         "Ghirapur Grand Prix",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThis("Ghirapur Grand Prix — all players start their engines",
				func(g *game.Game, item *game.StackItem) error {
					for _, p := range seatsFromActive(g) {
						g.StartEnginesForEffect(p)
					}
					return nil
				}),
			Targeting(AtYourEndStep("Ghirapur Grand Prix — X +1/+1 counters on target creature you control (X = your speed)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return AddCounter{Target: t, Kind: game.CounterPlusOne, N: YourSpeed(g, item.Controller)}.Apply(ctx)
				}), TargetCreature("target creature you control", YouControl())),
			WhenYouPlaneswalkAwayFromThis("Ghirapur Grand Prix — each player with the highest speed among players creates three Treasure tokens",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, p := range g.PlayersWithTheHighestSpeedForEffect() {
						if err := (CreateToken{Controller: p, Template: TreasureToken(), N: 3}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
