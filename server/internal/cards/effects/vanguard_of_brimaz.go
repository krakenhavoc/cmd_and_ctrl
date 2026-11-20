package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vanguard of Brimaz — Creature — Cat Soldier {W}{W}, 2/2 (Born of the Gods,
// #5559):
//
//	"Vigilance
//	 Heroic — Whenever you cast a spell that targets this creature, create a
//	 1/1 white Cat Soldier creature token with vigilance."
//
// Akroan Crusader's heroic token, a Cat Soldier from the table.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4a4f6e85-df16-49a4-8fc6-149a28c2cb42",
		Name:            "Vanguard of Brimaz",
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			heroic("Vanguard of Brimaz — create a 1/1 white Cat Soldier creature token with vigilance (heroic)",
				Do(CreateToken{Template: TokenCard("1/1 white Cat Soldier with vigilance"), N: 1})),
		},
	})
}
