package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steampath Charger — Creature — Lizard Warlock {1}{R}, 2/1 (Bloomburrow,
// #4467):
//
//	"Offspring {2} (You may pay an additional {2} as you cast this spell.
//	 If you do, when this creature enters, create a 1/1 token copy of it.)
//	 When this creature dies, it deals 1 damage to target player."
//
// Offspring is the shared keyword (offspring.go); the token copy has its
// own dies trigger. The dies trigger is Orcish Hellraiser's narrowed to
// a player: the Charger is in the graveyard by then and deals the damage
// as it last existed (CR 603.10a). A player who has left by resolution
// takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "75237b2c-232b-479a-ab3a-e1566c8a9fe7",
		Name:          "Steampath Charger",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Offspring("{2}")},
		Triggered: []game.TriggeredAbility{
			effects.OffspringToken("Steampath Charger"),
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisDies("Steampath Charger — 1 damage to target player", effects.SourceDealsDamageToEachLegalTarget(1)),
				effects.TargetPlayer("target player")), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
