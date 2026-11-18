package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "8a69a0f0-5f0f-4098-ab6d-be3233af5cd3",
		Name:             "Dryad's Revival",
		Completeness:     CompletenessFull,
		Targets:          TargetCardInGraveyard("target card in your graveyard", YouOwn()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{G}")},
		OnResolve:        returnTargetFromGraveyardToHand,
	})
}
