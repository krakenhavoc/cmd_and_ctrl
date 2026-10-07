package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Village Watch // Village Reavers — {4}{R} Creature — Human Werewolf 4/3 //
// Creature — Werewolf 5/4 (#2561, ADR 0132):
//
//	Front: "Haste
//	        Daybound"
//	Back:  "Wolves and Werewolves you control have haste.
//	        Nightbound"
//
// The back face's grant is the same lord shape Goblin Chieftain uses,
// with no "other": Village Reavers is itself a Werewolf, so it has haste
// too, which is what the printed text says. "You control" is printed, so
// the grant is scoped to the controller.
//
// No simplification.
func init() {
	const oracle = "b2ecaae4-41ee-4c61-b5ce-db4364b307fc"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Village Watch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste", "daybound"},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Village Reavers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Wolf", "Werewolf"}, YoursOnly: true}, "haste"),
		},
	})
}
