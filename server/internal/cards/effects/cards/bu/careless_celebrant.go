package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Careless Celebrant — Creature — Satyr Shaman {1}{R}, 2/1 (Theros Beyond Death, #5514):
//
//	"When this creature dies, it deals 2 damage to target creature or
//	 planeswalker an opponent controls."
//
// Bogardan Firefiend's dies trigger with this card's narrower clause.
// The Celebrant deals the damage from the graveyard as it last existed
// on the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56afad48-c5e5-4fbe-8889-15bcab234c31",
		Name:         "Careless Celebrant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisDies("Careless Celebrant — 2 damage to target creature or planeswalker an opponent controls",
					effects.SourceDealsDamageToEachLegalTarget(2)),
				effects.TargetPermanent("target creature or planeswalker an opponent controls",
					effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls())),
				effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
