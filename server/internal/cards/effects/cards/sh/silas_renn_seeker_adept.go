package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silas Renn, Seeker Adept — Legendary Artifact Creature — Human {1}{U}{B},
// 2/2:
//
//	"Deathtouch
//	 Whenever Silas Renn deals combat damage to a player, choose target
//	 artifact card in your graveyard. You may cast that card this turn.
//	 Partner (You can have two commanders if both have partner.)"
//
// Emry's cast permission (a plain CR 611.2c grant, you still pay its
// costs) behind a combat-damage trigger. Partner is a deck-construction
// rule read off the oracle text (CR 702.124).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0dc78b1e-581e-4f28-bf35-e2082553d6a9",
		Name:            "Silas Renn, Seeker Adept",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisDealsCombatDamageToAPlayer("Silas Renn, Seeker Adept — you may cast a target artifact card from your graveyard this turn",
					effects.EmryGrantCastFromGraveyard),
				effects.TargetCardInGraveyard("target artifact card in your graveyard", effects.YouOwn(), effects.Artifact()),
			),
		},
	})
}
