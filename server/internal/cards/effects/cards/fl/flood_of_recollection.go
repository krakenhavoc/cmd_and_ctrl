package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flood of Recollection — Sorcery {U}{U} (Commander Legends):
//
//	"Return target instant or sorcery card from your graveyard to your hand.
//	 Exile Flood of Recollection."
//
// A graveyard return with a target re-checked on resolution (CR 608.2b), and
// the spell exiling itself as the last part of its resolution (Temporal
// Mastery's shape). If the target is illegal the spell does not resolve and
// goes to the graveyard instead, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02393093-e970-40a0-848b-9cd758450cbb",
		Name:         "Flood of Recollection",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.ReturnTargetCardToHand(ctx.Game, item); err != nil {
				return err
			}
			return effects.ExileTarget{Target: item.SourceCardID}.Apply(ctx)
		},
	})
}
