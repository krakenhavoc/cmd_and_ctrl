package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Village Messenger // Moonrise Intruder — {R} Creature — Human Werewolf
// 1/1 // Creature — Werewolf 2/2 (Innistrad Remastered, S82 GAP, #5948):
//
//	Front: "Haste
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Menace
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The original-Innistrad werewolf pair of triggers
// (innistrad_werewolves.go) around a printed keyword on each face.
//
// No simplification.
func init() {
	const oracle = "f6634a84-a9ce-4bbd-bc2c-3f9179b7f16a"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Village Messenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfNoSpellsLastTurn("Village Messenger"),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Moonrise Intruder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Moonrise Intruder"),
		},
	})
}
