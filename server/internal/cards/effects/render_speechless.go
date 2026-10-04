package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Render Speechless — Sorcery {2}{W}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card.
//	 Put two +1/+1 counters on up to one target creature."
//
// Two target clauses, announced together (CR 601.2c): the opponent,
// then up to one creature. Each half acts on its own clause's pick only
// while it is still legal (CR 608.2b), so a creature that leaves in
// response still leaves the discard, and an opponent who has left the
// game still leaves the counters. The pick is Thoughtseize's
// (ADR 0116); the counters are printed after it and run on the next
// line (ADR 0116 §6), since they neither read the chosen card nor
// change which cards may be chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6fdddab4-fcc9-4622-9c17-c69f9d4bca08",
		Name:         "Render Speechless",
		Completeness: CompletenessFull,
		Targets: TargetPlayer("target opponent", Opponent()).
			Then(TargetCreature("up to one target creature").WithCount(0, 1)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: Nonland(),
				Label:  "nonland card",
			}).Apply(ctx); err != nil {
				return err
			}
			t, ok := ctx.ClauseTarget(1)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
		},
	})
}
