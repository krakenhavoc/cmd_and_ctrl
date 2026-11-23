package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salvage Scout — Creature — Human Scout {W}, 1/1 (Scars of Mirrodin, #5577):
//
//	"{W}, Sacrifice this creature: Return target artifact card from your
//	 graveyard to your hand."
//
// Argivian Archaeologist's return with a sacrifice cost. The Scout is
// sacrificed as the ability is activated, after its target is chosen, so
// it can't return itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "502e991c-f4de-4095-adac-cb0c25e44a7c",
		Name:         "Salvage Scout",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, Sacrifice this creature: Return target artifact card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn()),
			Effect:  effects.ReturnTargetedCardToHand,
		}},
	})
}
