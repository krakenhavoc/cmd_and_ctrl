package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frantic Salvage — Instant {3}{W} (Mirrodin Besieged, S82 OLD b468, #5576):
//
//	"Put any number of target artifact cards from your graveyard on top of your
//	 library.
//	 Draw a card."
//
// Any number of targets, none included (CR 115.1); a spell that chose some and
// has none still legal does nothing at all (CR 608.2b), draw included. The
// cards still there are put on top in an order their owner chooses (CR 401.4;
// the put_in_library prompt), and the draw waits for that answer, so it may
// take one of them. With no target chosen the spell only draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d2e7099-dcd5-428f-a732-2a82b8d1dd1a",
		Name:         "Frantic Salvage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerValue},
		Targets: effects.TargetCardInGraveyard("any number of target artifact cards from your graveyard", effects.Artifact(), effects.YouOwn()).
			WithCount(0, 0),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(item.Targets) > 0 && len(ids) == 0 {
				return nil
			}
			if len(ids) == 0 {
				return effects.DrawCards{N: 1}.Apply(ctx)
			}
			return effects.PutInLibraryInAnyOrder{
				Chooser:   ctx.Controller(),
				Cards:     ids,
				From:      game.ZoneGraveyard,
				Placement: game.LibraryPlaceTop,
				Label:     "Frantic Salvage — put the artifact cards on top of your library",
				ThenTail:  ctx.Tail(franticSalvageDraw),
			}.Apply(ctx)
		},
	})
}

// franticSalvageDraw is "Draw a card" once the pile is placed. An on-disk key:
// never renamed, never reused.
var franticSalvageDraw = effects.CardTail("frantic-salvage/draw", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.DrawCards{N: 1}.Apply(ctx)
})
