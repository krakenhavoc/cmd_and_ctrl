package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cut a Deal — Sorcery, {2}{W} (EDHREC rank 996):
//
//	"Each opponent draws a card, then you draw a card for each
//	 opponent who drew a card this way."
//
// White's group-hug draw three: everyone gets a card, and at a full
// table the caster gets three. The opponents draw in seat order,
// each an ordinary draw (Sheoldred and Razorkin Needlehead see every
// one), and the caster draws the count of opponents who actually
// had a card to draw — an opponent with an empty library draws
// nothing and earns nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b5cffd5-5db3-4151-9275-d91387554412",
		Name:         "Cut a Deal",
		Completeness: CompletenessFull,
		// The caster's draw waits for a prompt an opponent's asks (a
		// dredge offer; CR 608.2c), and counts the opponents who drew
		// in the event log after the step record (#4163): a draw a
		// dredge replaced is not one.
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				ctx.RecordStep(cutADealRecord, nil, int(b25LastEventSeq(ctx.Game)))
				_, err := b08EachOpponentDraws(ctx.Game, ctx.Item)
				return err
			}),
			Instruction(func(ctx *Context) error {
				after := uint64(ctx.StepRecord(cutADealRecord).N())
				drew := 0
				for _, opp := range ctx.Opponents() {
					for _, ev := range ctx.Game.Events {
						if ev.Seq > after && ev.Kind == game.EventDrawCard && ev.Actor == opp {
							drew++
							break
						}
					}
				}
				return DrawCards{N: drew}.Apply(ctx)
			}),
		),
	})
}

// cutADealRecord is the last event logged before the opponents drew.
const cutADealRecord = "cut-a-deal/before-draws"
