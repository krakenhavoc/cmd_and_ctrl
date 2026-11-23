package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Groundskeeper — Creature — Human Druid {G}, 1/1 (Foundations
// Commander):
//
//	"{1}{G}: Return target basic land card from your graveyard to your
//	 hand."
//
// Carnivorous Cultivator's graveyard-to-hand return with a basic-land
// target clause on an activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2c19fba-9b06-461f-bb5e-e827ba41204d",
		Name:         "Groundskeeper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Return target basic land card from your graveyard to your hand.",
			Cost:    effects.ManaCost("{1}{G}"),
			Targets: effects.TargetCardInGraveyard("target basic land card from your graveyard", effects.CmdrB9BasicLandCard(), effects.YouOwn()),
			Effect:  effects.ReturnTargetGraveyardCardToHand,
		}},
	})
}
