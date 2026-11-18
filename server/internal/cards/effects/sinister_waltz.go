package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sinister Waltz — Sorcery {3}{B}{R} (Innistrad: Crimson Vow):
//
//	"Choose three target creature cards in your graveyard. Return two of
//	 them at random to the battlefield and put the other on the bottom of
//	 your library."
//
// The two that return are drawn from the keyed random stream (undo
// rewinds it) among the targets still legal at resolution (CR 608.2b). A
// target that has left is skipped: with only two left both return and
// nothing goes to the bottom. The two enter as one simultaneous entry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36a7d1a6-4eb8-4ddc-942c-00fd2acbf2ee",
		Name:         "Sinister Waltz",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("three target creature cards in your graveyard", Creature(), YouOwn()).WithCount(3, 3),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ids := legalTargetCardIDs(ctx)
			pick := ctx.Game.ChooseAtRandomForEffect(randomDraw(ctx), ids, 2)
			picked := make(map[uuid.UUID]bool, len(pick))
			for _, id := range pick {
				picked[id] = true
			}
			var rest []uuid.UUID
			for _, id := range ids {
				if !picked[id] {
					rest = append(rest, id)
				}
			}
			if err := ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneGraveyard, rest); err != nil {
				return err
			}
			return ReturnFromGraveyardTogether{Targets: pick}.Apply(ctx)
		},
	})
}
