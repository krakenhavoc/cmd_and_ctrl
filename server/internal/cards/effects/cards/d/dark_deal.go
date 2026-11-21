package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Deal — Sorcery {2}{B} (EDHREC rank 1457):
//
//	"Each player discards all the cards in their hand, then draws
//	 that many cards minus one."
//
// The wheel that shrinks. Wheel of Fortune's body with the count
// read per player: every discard happens before any draw, so
// discard payoffs queue while the effect is still resolving, and
// each player then draws their own hand size less one. A player who
// had nothing draws nothing rather than minus one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c527eb80-ccac-40b8-8377-c31121613128",
		Name:         "Dark Deal",
		Completeness: effects.CompletenessFull,
		// The draws wait for a prompt a discard asks (a madness offer;
		// CR 608.2c), and read each player's count from the step record
		// (#4163).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				players := effects.TablePlayers(ctx)
				draws := make([]int, len(players))
				for i, id := range players {
					n, err := effects.DiscardWholeHand(ctx.Game, id)
					if err != nil {
						return err
					}
					draws[i] = n - 1
				}
				ctx.RecordStep(darkDealRecord, players, draws...)
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				r := ctx.StepRecord(darkDealRecord)
				for i, id := range r.IDs {
					if i >= len(r.Ns) {
						break
					}
					if err := (effects.DrawCards{Player: id, N: r.Ns[i]}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}

// darkDealRecord is each player, and the cards they draw (the hand
// they discarded, less one).
const darkDealRecord = "dark-deal/draws"
