package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bomber Corps — Creature — Human Soldier {1}{R}, 1/2 (Gatecrash, #5565):
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, this creature deals 1 damage to any target."
//
// Firemane Avenger's shape: battalion, then damage to the one target
// still legal, dealt by this creature (by last-known information if it
// has left, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a429c351-d281-4e49-8564-bd96bdf5c41b",
		Name:         "Bomber Corps",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.GtcBattalion("Bomber Corps — 1 damage to any target (battalion)",
				effects.SourceDealsDamageToEachLegalTarget(1)), effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
