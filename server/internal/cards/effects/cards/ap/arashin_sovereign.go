package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arashin Sovereign — Creature — Dragon {5}{G}{W}, 6/6 (Dragons of Tarkir, #5551):
//
//	"Flying
//	 When this creature dies, you may put it on your choice of the top or bottom of its owner's library."
//
// An optional dies trigger whose placement is the shared top-or-bottom choice
// (Hinder's): you are asked whether to, then which end. The card goes to its
// owner's library, and is skipped if it has left the graveyard (CR 400.7).
//
// No simplification.

func init() {
	const label = "Arashin Sovereign — put it on your choice of the top or bottom of its owner's library"
	effects.Register(effects.Spec{
		OracleID:        "cb58055f-8190-4bc7-ad32-844ca66fd2b0",
		Name:            "Arashin Sovereign",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisDies(label, effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.PutInLibraryInAnyOrder{
					Chooser:   ctx.Controller(),
					Cards:     []uuid.UUID{ctx.Source()},
					From:      game.ZoneGraveyard,
					Placement: game.LibraryPlaceTopOrBottom,
					Label:     label,
				}.Apply(ctx)
			}))), label+"?"),
		},
	})
}
