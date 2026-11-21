package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ecc8e782-90a0-4a4d-b4cb-1d7383938c14",
		Name:         "Lazotep Archway",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|B}",
			Label:    "Add {W} or {B}",
		}},
		Activated: []effects.ActivatedAbility{effects.GapB31EternalizeCreature("{3}{W}{B}")},
	})
}
