package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scorned Villager // Moonscarred Werewolf — {1}{G} Creature — Human
// Werewolf 1/1 // Creature — Werewolf 2/2 (#4071, ADR 0132 amendment
// 2026-10-16):
//
//	Front: "{T}: Add {G}.
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Vigilance
//	        {T}: Add {G}{G}.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// Two plain tap mana abilities, one per face; summoning sickness applies
// to both (CR 302.6, enforced by the mana-ability path).
//
// No simplification.
func init() {
	const oracle = "52855f90-19c1-46c9-8eed-88b3c1722bb0"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Scorned Villager",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfNoSpellsLastTurn("Scorned Villager"),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Moonscarred Werewolf",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}{G}",
			Label:    "Add {G}{G}",
		}},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Moonscarred Werewolf"),
		},
	})
}
