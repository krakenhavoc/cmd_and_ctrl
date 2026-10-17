package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kruin Outlaw // Terror of Kruin Pass — {1}{R}{R} Creature — Human Rogue
// Werewolf 2/2 // Creature — Werewolf 3/3 (#4071, ADR 0132 amendment
// 2026-10-16):
//
//	Front: "First strike
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Double strike
//	        Werewolves you control have menace.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The menace grant is Village Reavers' tribal keyword grant, scoped to
// your Werewolves and including the Terror itself.
//
// No simplification.
func init() {
	const oracle = "cf518f59-f083-42a0-9953-fac2757bf804"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Kruin Outlaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfNoSpellsLastTurn("Kruin Outlaw"),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Terror of Kruin Pass",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Werewolf"}, YoursOnly: true}, "menace"),
		},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Terror of Kruin Pass"),
		},
	})
}
