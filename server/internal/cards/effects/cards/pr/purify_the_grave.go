package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Purify the Grave — Instant {W} (Innistrad, #5572):
//
//	"Exile target card from a graveyard.
//	 Flashback {W} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Coffin Purge's body (a card in any graveyard, exiled) with its
// Flashback at {W}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "1dafc03b-6c45-4f28-8a1f-b4f1044adc97",
		Name:             "Purify the Grave",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCardInGraveyard("target card from a graveyard"),
		Purpose:          effects.ForTargets(effects.ExilesFromGraveyard(0, "")),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{W}")},
		OnResolve:        effects.ExileTargetCardsOnResolve,
	})
}
