package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Misinformation — Instant {B} (Alliances, S82 OLD batch 817):
//
//	"Put up to three target cards from an opponent's graveyard on top of
//	 their library in any order."
//
// "A single graveyard" is the shared clause, narrowed to cards the caster
// does not own; the caster names the order (CR 401.4 is who knows it
// afterwards), and each card goes to its owner's library, which is the one
// graveyard it came from.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03dc8820-ec10-42c1-8200-d323cfcbf2b0",
		Name:         "Misinformation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.UpToCardsFromASingleGraveyard("up to three target cards from an opponent's graveyard", 3, effects.OldB621OpponentOwns()),
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
				Label:     "Misinformation — put the cards on top of their library in any order",
			}.Apply(ctx)
		},
	})
}
