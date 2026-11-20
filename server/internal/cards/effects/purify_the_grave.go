package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "1dafc03b-6c45-4f28-8a1f-b4f1044adc97",
		Name:             "Purify the Grave",
		Completeness:     CompletenessFull,
		Targets:          TargetCardInGraveyard("target card from a graveyard"),
		Purpose:          ForTargets(ExilesFromGraveyard(0, "")),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{W}")},
		OnResolve:        exileTargetCardsOnResolve,
	})
}
