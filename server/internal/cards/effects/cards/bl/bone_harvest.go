package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bone Harvest — Instant {2}{B} (Mirage, S82 OLD batch 812, #5633):
//
//	"Put any number of target creature cards from your graveyard on top of your
//	 library.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Any number (zero too) of creature cards from the caster's graveyard are put
// on top of the library in the order the caster chooses (PutInLibraryInAnyOrder,
// CR 401.4); each target that has left the graveyard is skipped (CR 608.2b).
// The delayed draw (CR 603.7) is Mishra's Bauble's, scheduled whatever was put.
//
// No simplification.
func init() {
	const label = "Bone Harvest — draw a card at the beginning of the next turn's upkeep"
	effects.Register(effects.Spec{
		OracleID:     "0cc38d20-5ae8-413a-9f26-ff13cd92f24c",
		Name:         "Bone Harvest",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("any number of target creature cards from your graveyard", effects.Creature(), effects.YouOwn()).WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.S82OldB812DelayedDrawAtNextUpkeep(ctx, label); err != nil {
				return err
			}
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
				Label:     "Bone Harvest — put the creature cards on top of your library in any order",
			}.Apply(ctx)
		},
	})
}
