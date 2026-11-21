package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pinnacle Kill-Ship — Artifact — Spacecraft {7}, 7/7 (Edge of
// Eternities, #4461):
//
//	"When this Spacecraft enters, it deals 10 damage to up to one target
//	 creature.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 7+.)
//	 7+ | Flying"
//
// Galvanizing Sawship's station lines (Station, SpacecraftAt and
// ThresholdKeywords, CR 702.184a and 721.2a) with an entry trigger: the
// Kill-Ship is the source of the 10 damage although it is not yet a
// creature, "up to one" lets it enter with no target, and a target gone
// by resolution takes none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad556aee-3dbf-4c8b-9f3a-31947e26c6f5",
		Name:         "Pinnacle Kill-Ship",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Pinnacle Kill-Ship — 10 damage to up to one target creature", effects.SourceDealsDamageToEachLegalTarget(10)),
				effects.TargetCreature("up to one target creature").WithCount(0, 1)),
				effects.ForTargets(effects.DamageToTarget(0, 10))),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(7, 7, 7),
			effects.ThresholdKeywords(7, "flying"),
		),
	})
}
