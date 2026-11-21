package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lazotep Archway — Land:
//
//	"Lazotep Archway enters the battlefield tapped.
//	 {T}: Add {W} or {B}.
//	 Eternalize {3}{W}{B} ({3}{W}{B}, Exile this card from your graveyard:
//	 Create a token that's a copy of it, except it's a 4/4 black Zombie
//	 creature and loses all other card types. Eternalize only as a sorcery.
//	 And don't forget it still enters tapped!)"
//
// Eternalize with a land's twist: the token is a Creature — Zombie and no
// longer a land, but it is still a copy, so it keeps both abilities and
// enters tapped (gapB31EternalizeCreature, s82_gap_b31_helpers.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ecc8e782-90a0-4a4d-b4cb-1d7383938c14",
		Name:         "Lazotep Archway",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|B}",
			Label:    "Add {W} or {B}",
		}},
		Activated: []ActivatedAbility{gapB31EternalizeCreature("{3}{W}{B}")},
	})
}
