package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Festive Funeral — Instant {4}{B} (Throne of Eldraine, S82 old sets):
//
//	"Target creature gets -X/-X until end of turn, where X is the number of
//	 cards in your graveyard."
//
// X is read as the spell resolves, with Festive Funeral itself still on the
// stack and so not in the graveyard (CR 608.2). A target that has gone is not
// affected (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Festive Funeral — target creature gets -X/-X until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "85d7c8d4-6f99-4b1d-9e0e-ff41bbeee682",
		Name:         "Festive Funeral",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			x := effects.B31GraveyardSize(ctx.Game, item.Controller)
			return effects.BoostUntilEOT{Target: id, Power: -x, Toughness: -x, Label: label}.Apply(ctx)
		},
	})
}
