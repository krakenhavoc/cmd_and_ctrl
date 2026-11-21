package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Argivian Find — Instant {W} (Weatherlight, #5629):
//
//	"Return target artifact or enchantment card from your graveyard to your
//	 hand."
//
// Elven Cache's return, narrowed to an artifact or enchantment card. A
// target that left the graveyard in response does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b6c74a7-d7d3-4c05-84b1-bcb580bb476f",
		Name:         "Argivian Find",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target artifact or enchantment card from your graveyard", effects.Or(effects.Artifact(), effects.Enchantment()), effects.YouOwn()),
		OnResolve:    effects.ReturnTargetFromGraveyardToHand,
	})
}
