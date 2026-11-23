package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temporal Cleansing — Sorcery {3}{U}, convoke:
//
//	"Convoke (Your creatures can help cast this spell. Each creature
//	 you tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 The owner of target nonland permanent puts it into their library
//	 second from the top or on the bottom."
//
// #1298, on ADR 0088's put_in_library (2026-09-23 amendment). A
// two-way choice whose top lane is a DEPTH, and whose chooser is the
// permanent's OWNER rather than the caster — PutIntoLibraryAtDepthOrBottom,
// which raises a `top_or_bottom` prompt with TopDepth 2 addressed to
// the owner. The permanent is public, so the choice hides nothing, and
// the move is the tuck route either way: the CR 614 window opens and a
// commander is offered the command zone (CR 903.9).
//
// Convoke is Spec.TapCost (S22).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f67698a-28db-43cd-aaf1-52371b0b47eb",
		Name:         "Temporal Cleansing",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Not(effects.Land())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.PutIntoLibraryAtDepthOrBottom{
				Card:  item.Targets[0].ID,
				Depth: 2,
				Label: "Temporal Cleansing — put it into your library second from the top or on the bottom",
			}.Apply(ctx)
		},
	})
}
