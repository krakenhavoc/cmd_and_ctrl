package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Poison Dart Frog — Creature — Frog {1}{G}, 1/1 (The Lost Caverns of
// Ixalan, #4470):
//
//	"Reach
//	 {T}: Add one mana of any color.
//	 {2}: This creature gains deathtouch until end of turn."
//
// Birds of Paradise's mana ability (CR 605.1a), reach as the printed
// keyword, and Fangblade Brigand's self-grant with no pump: deathtouch
// until end of turn, which does nothing if the Frog has left by the
// time the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e88a2b71-c90d-4e16-99a3-93d11fe7a968",
		Name:            "Poison Dart Frog",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}"),
			Effect:  effects.ThisCreatureUntilEOT("Poison Dart Frog — deathtouch until end of turn", 0, 0, "deathtouch"),
		}},
	})
}
