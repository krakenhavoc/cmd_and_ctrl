package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Virtus the Veiled — Legendary Creature — Azra Assassin {2}{B}, 1/1
// (Battlebond, #5529):
//
//	"Partner with Gorm the Great (When this creature enters, target player
//	 may put Gorm into their hand from their library, then shuffle.)
//	 Deathtouch
//	 Whenever Virtus deals combat damage to a player, that player loses
//	 half their life, rounded up."
//
// Partner with is the shared CR 702.124j pair (PartnerWith). Deathtouch
// rides PrintedKeywords. The damage trigger is Unstoppable Slasher's: the
// shared WheneverThisDealsCombatDamageToAPlayer with the halving read as it
// resolves (21 loses 11), aimed at the player the damage event names.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6da4f16c-f187-479e-b19b-78a137b1abe5",
		Name:            "Virtus the Veiled",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Virtus the Veiled", "Gorm the Great"),
			effects.WheneverThisDealsCombatDamageToAPlayer("Virtus the Veiled — that player loses half their life, rounded up",
				effects.TriggeringTargetLosesHalfTheirLife),
		},
	})
}
