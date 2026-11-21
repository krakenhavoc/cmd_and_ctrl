package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auroral Procession — Instant {G}{U} (EDHREC rank 3583):
//
//	"Return target card from your graveyard to your hand."
//
// Regrowth at instant speed, any card type. A target that left the
// graveyard in response fizzles the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "511d3de9-076b-4811-bddf-00623f98993f",
		Name:         "Auroral Procession",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B34ReturnChosenGraveyardCardToHand(ctx)
		},
	})
}
