package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Takeno's Cavalry — Creature — Human Samurai Archer {3}{W}, 1/1 (Betrayers
// of Kamigawa, #5602):
//
//	"Bushido 1
//	 {T}: This creature deals 1 damage to target attacking or blocking
//	 Spirit."
//
// Bushido 1 is the shared keyword trigger (CR 702.45a). The ability is Centaur
// Archer's, aimed at a Spirit creature that is attacking or blocking, read
// through the layers at announce and again as it resolves (CR 608.2b), so it
// is only legal in combat.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5577514-d2af-4904-9f94-b472228c6e3d",
		Name:         "Takeno's Cavalry",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Bushido(1, "Takeno's Cavalry")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target attacking or blocking Spirit.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target attacking or blocking Spirit", effects.Subtype("Spirit"), effects.AttackingOrBlocking()),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
