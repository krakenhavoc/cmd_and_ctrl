package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rootwater Diver — {U} Creature — Merfolk, 1/1 (Tempest, S82 OLD batch 770,
// #5628):
//
//	"{T}, Sacrifice this creature: Return target artifact card from your
//	 graveyard to your hand."
//
// Tap and the sacrifice are costs, paid on activation (CR 602.2), so the Diver is
// gone while the ability is on the stack. The target is an artifact card in YOUR
// graveyard, re-checked as the ability resolves (CR 608.2b); one that has left
// does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4145c153-1ee3-4583-8a60-8471959656eb",
		Name:         "Rootwater Diver",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Return target artifact card from your graveyard to your hand.",
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true},
			Targets: effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn()),
			Effect:  effects.ReturnTargetCardToHand,
		}},
	})
}
