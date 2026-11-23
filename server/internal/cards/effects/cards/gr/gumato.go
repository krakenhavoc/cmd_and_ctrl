package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gumato — Creature — Ape Beast {4}{G}, 4/5 (Star Trek Commander, #4974):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)
//	 {T}: Add {G}{G}{G}."
//
// A fixed-colour mana ability (CR 605.1a): it resolves at once, and the
// {T} in its cost keeps it off while Gumato is summoning sick (CR 302.6).
// Vigilance means it can attack and still tap for mana after combat.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "67114cbb-cbf4-4250-a5e9-e1c4bc5841f6",
		Name:            "Gumato",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}{G}{G}",
			Label:    "Add {G}{G}{G}",
		}},
	})
}
