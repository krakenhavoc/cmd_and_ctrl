package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reinforcements — Instant {W} (Alliances, S82 OLD batch 820):
//
//	"Put up to three target creature cards from your graveyard on top of your
//	 library."
//
// You put them there in the order you choose (CR 401.4, the same pile put
// back "in any order" Misinformation names), and a target gone by
// resolution is simply left out.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff27ff96-afd2-45df-b799-3046179e27c6",
		Name:         "Reinforcements",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("up to three target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return effects.PutInLibraryInAnyOrder{
				Cards:     ids,
				From:      game.ZoneGraveyard,
				Placement: game.LibraryPlaceTop,
				Label:     "Reinforcements — put the cards on top of your library in any order",
			}.Apply(ctx)
		},
	})
}
