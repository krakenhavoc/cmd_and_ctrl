package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunting Velociraptor — Creature — Dinosaur {2}{R}, 3/2 (Jurassic World
// Collection, #5948):
//
//	"First strike
//	 Dinosaur spells you cast have prowl {2}{R}. (You may cast a spell for
//	 its prowl cost if you dealt combat damage to a player this turn with a
//	 creature with any of its creature types.)"
//
// A granted alternative cost (ADR 0118 §3) carrying prowl's own gate
// (prowlConditionMet, CR 702.76a): the spell's creature types are read in the
// zone it is cast from, and a creature of the caster's that dealt combat
// damage this turn must share one. The offer reaches Dinosaur spells you cast
// from any zone, so a prowl-style cast out of the graveyard by another
// permission is as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bca33fb6-cd37-4259-8c0b-7a4675754a00",
		Name:            "Hunting Velociraptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{{
			Offer: game.AlternativeCost{
				Key:            GrantedAltCostProwl,
				Label:          "Prowl {2}{R}",
				ManaCost:       "{2}{R}",
				SpellCondition: prowlConditionMet,
			},
			Spells:     game.PermissionFilter{CreatureType: "Dinosaur"},
			SpellsText: "Dinosaur spells",
		}},
	})
}
