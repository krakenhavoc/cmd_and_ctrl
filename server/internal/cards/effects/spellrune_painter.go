package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spellrune Painter // Spellrune Howler — {2}{R} Creature — Human Shaman
// Werewolf 2/3 // Creature — Werewolf 3/4 (#2586, ADR 0132):
//
//	Front: "Whenever you cast an instant or sorcery spell, this creature
//	        gets +1/+1 until end of turn.
//	        Daybound"
//	Back:  "Whenever you cast an instant or sorcery spell, this creature
//	        gets +2/+2 until end of turn.
//	        Nightbound"
//
// The cast trigger and the self-pump are existing vocabulary; each face
// has its own amount.
//
// No simplification.
func init() {
	const oracle = "4ff8c359-430f-434e-ac1e-822abdc28360"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Spellrune Painter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()), "Spellrune Painter — gets +1/+1 until end of turn",
				thisGetsUntilEndOfTurn(1, 1, "Spellrune Painter — instant or sorcery pump")),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Spellrune Howler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()), "Spellrune Howler — gets +2/+2 until end of turn",
				thisGetsUntilEndOfTurn(2, 2, "Spellrune Howler — instant or sorcery pump")),
		},
	})
}
