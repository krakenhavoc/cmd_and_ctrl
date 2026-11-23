package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keldon Marauders — Creature — Human Warrior {1}{R}, 3/3 (Planar Chaos, #5595):
//
//	"Vanishing 2 (This creature enters with two time counters on it. At the
//	 beginning of your upkeep, remove a time counter from it. When the last is
//	 removed, sacrifice it.)
//	 When this creature enters or leaves the battlefield, it deals 1 damage to
//	 target player or planeswalker."
//
// Vanishing is cmdr2B17Vanishing* (Deep Forest Hermit): the entry counters
// are a replacement, the upkeep removal and the last-counter sacrifice are
// two triggers (CR 702.63a).
// The enters-or-leaves trigger is Bogardan Firefiend's targeted damage
// from the source, which a source that has left deals through its last
// known information (CR 113.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e05b850f-24cd-42b9-bc4f-0f05f9aea679",
		Name:         "Keldon Marauders",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Cmdr2B17VanishingEntry(2)},
		Triggered: append(effects.Cmdr2B17VanishingTriggers(),
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEntersOrLeaves("Keldon Marauders — 1 damage to target player or planeswalker",
				effects.SourceDealsDamageToEachLegalTarget(1)), effects.TargetPlayerOrPlaneswalker()), effects.ForTargets(effects.DamageToTarget(0, 1)))),
	})
}
