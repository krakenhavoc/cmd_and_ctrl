package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "5b865275-4114-44d0-9832-ae5d3b4efba4",
		Name:            "Frog Butler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}: This creature gains reach until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{2}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("reach", "Frog Butler — reach until end of turn"),
		}},
	})
}
