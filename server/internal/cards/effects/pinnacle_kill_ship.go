package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ad556aee-3dbf-4c8b-9f3a-31947e26c6f5",
		Name:         "Pinnacle Kill-Ship",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Pinnacle Kill-Ship — 10 damage to up to one target creature", sourceDealsDamageToEachLegalTarget(10)),
				TargetCreature("up to one target creature").WithCount(0, 1)),
				ForTargets(DamageToTarget(0, 10))),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(7, 7, 7),
			ThresholdKeywords(7, "flying"),
		),
	})
}
