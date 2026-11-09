package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pull Through the Weft — Sorcery {3}{G}{G} (Edge of Eternities):
//
//	"Return up to two target nonland permanent cards from your graveyard
//	 to your hand, then return up to two target land cards from your
//	 graveyard to the battlefield tapped."
//
// Two target clauses in printed order (Decimate's chain), each "up to two"
// and judged on its own at resolution (CR 608.2b). The nonland cards go to
// hand first; then the lands enter together, tapped, as one event so a
// landfall trigger sees them all at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "37bd9e54-c93e-4fec-9b21-6940da75ab7a",
		Name:         "Pull Through the Weft",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: TargetCardInGraveyard("up to two target nonland permanent cards from your graveyard",
			YouOwn(), Permanent(), Nonland()).WithCount(0, 2).Then(
			TargetCardInGraveyard("up to two target land cards from your graveyard", YouOwn(), Land()).WithCount(0, 2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var hand, lands []uuid.UUID
			for _, t := range ctx.Targets() {
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if t.Slot == 0 {
					hand = append(hand, t.ID)
				} else {
					lands = append(lands, t.ID)
				}
			}
			for _, id := range hand {
				if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			if len(lands) == 0 {
				return nil
			}
			return ReturnFromGraveyardTogether{Targets: lands, Tapped: true}.Apply(ctx)
		},
	})
}
