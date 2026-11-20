package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dryad's Revival — Sorcery {2}{G} (Midnight Hunt, #5497):
//
//	"Return target card from your graveyard to your hand.
//	 Flashback {4}{G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Regrowth's body (returnTargetFromGraveyardToHand) with the shared
// Flashback. Cast from the graveyard it cannot return itself: it is on the
// stack, not in the graveyard, when it targets.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "8a69a0f0-5f0f-4098-ab6d-be3233af5cd3",
		Name:             "Dryad's Revival",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCardInGraveyard("target card in your graveyard", effects.YouOwn()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{4}{G}")},
		OnResolve:        effects.ReturnTargetFromGraveyardToHand,
	})
}
