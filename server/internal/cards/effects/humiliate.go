package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Humiliate — Sorcery {W}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card. Put a +1/+1 counter on a
//	 creature you control."
//
// Thoughtseize's pick (ADR 0116), then the counter. The creature is
// chosen as Humiliate resolves, after the hand has been revealed, and is
// not targeted (the 2021-04-16 rulings), so it is a choice among the
// creatures you control at that moment. Humiliate can be cast with no
// creature; the counter then goes nowhere (CR 609.3).
//
// The counter is printed after the discard and its prompt is queued on
// the next line, after the pick (ADR 0116 §6): it neither reads the
// chosen card nor changes which cards may be chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf18a1ca-1fb5-4c90-b20a-6ae2fb54a9bb",
		Name:         "Humiliate",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: Nonland(),
				Label:  "nonland card",
			}).Apply(ctx); err != nil {
				return err
			}
			return ChoosePermanents{
				Question:   "Humiliate — put a +1/+1 counter on a creature you control",
				Candidates: humiliateCreatures,
				Then: func(ctx *Context, picked game.PromptedPicks) error {
					for _, id := range picked.Cards() {
						if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			}.Apply(ctx)
		},
	})
}

// humiliateCreatures offers exactly one of the creatures `of` controls,
// or nothing when they control none.
func humiliateCreatures(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	ids := creaturesControlledByPlayer(g, of)
	if len(ids) == 0 {
		return nil, 0, 0
	}
	return ids, 1, 1
}
