package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frog Butler — {1}{G} Creature — Frog Spirit 1/1 (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Deathtouch
//	 {T}: Add one mana of any color.
//	 {2}: This creature gains reach until end of turn."
//
// Birds of Paradise's any-colour mana ability, and Endling's "gains
// <keyword> until end of turn" activated ability on its own source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5b865275-4114-44d0-9832-ae5d3b4efba4",
		Name:            "Frog Butler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature gains reach until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("reach", "Frog Butler — reach until end of turn"),
		}},
	})
}
