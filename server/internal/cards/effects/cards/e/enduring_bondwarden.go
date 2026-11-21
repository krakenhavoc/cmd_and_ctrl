package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enduring Bondwarden — Creature — Human Scout {W}, 0/1 (March of the
// Machine):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target
//	 creature. If that's another creature, it gains the following ability
//	 until end of turn.)
//	 When this creature dies, put its counters on target creature you
//	 control."
//
// The dies trigger is Dockworker Drone's (PutItsCountersOnTargetCreature):
// every counter the creature had as it last existed (CR 122.8, 603.10), on
// the target if it is still legal as the ability resolves. Backup grants it
// as a bundle, so a creature that gained it moves its own counters when it
// dies this turn.
//
// No simplification.
const enduringBondwardenGrant = "enduring-bondwarden/dies-counters"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "c17d4818-5825-4bff-8fc5-5fd740cc0c70",
		Name:         "Enduring Bondwarden",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       enduringBondwardenGrant,
			Triggered: []game.TriggeredAbility{enduringBondwardenDies()},
			Text:      "When this creature dies, put its counters on target creature you control.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Backup("Enduring Bondwarden", 1, enduringBondwardenGrant),
			enduringBondwardenDies(),
		},
	})
}

func enduringBondwardenDies() game.TriggeredAbility {
	return effects.Targeting(effects.WhenThisDies("Enduring Bondwarden — put its counters on target creature you control",
		effects.PutItsCountersOnTargetCreature),
		effects.TargetCreature("target creature you control", effects.YouControl()))
}
