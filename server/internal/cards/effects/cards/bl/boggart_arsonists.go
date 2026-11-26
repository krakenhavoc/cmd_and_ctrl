package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boggart Arsonists — Creature — Goblin Rogue {2}{R}, 2/1 (Shadowmoor, #5591):
//
//	"Plainswalk
//	 {2}{R}, Sacrifice this creature: Destroy target Scarecrow or Plains."
//
// Plainswalk is the keyword (CR 702.14). "Scarecrow or Plains" is any
// permanent with either subtype: a Scarecrow artifact creature, or any
// land with the Plains type, basic or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "54831951-2cd3-41bb-8326-6cfce48feb34",
		Name:            "Boggart Arsonists",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"plainswalk"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice this creature: Destroy target Scarecrow or Plains.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target Scarecrow or Plains", effects.AnySubtype("Scarecrow", "Plains")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
