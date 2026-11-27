package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auriok Survivors — Creature — Human Soldier {5}{W}, 4/6:
//
//	"When this creature enters, you may return target Equipment card from your
//	 graveyard to the battlefield. If you do, you may attach it to this
//	 creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "103b92c7-8f8e-4924-8d7e-fc0a01486e86",
		Name:         "Auriok Survivors",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Auriok Survivors — you may return target Equipment card from your graveyard to the battlefield",
					effects.AuriokSurvivorsReturn),
				effects.TargetCardInGraveyard("target Equipment card from your graveyard", effects.YouOwn(), effects.HasSubtype("Equipment"))),
				"Auriok Survivors — return target Equipment card from your graveyard to the battlefield?"),
		},
	})
}
