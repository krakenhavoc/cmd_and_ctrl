package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basilica Shepherd — Creature — Phyrexian Angel {3}{W}{W}, 3/3:
//
//	"Flying
//	 When this creature enters, create two 1/1 colorless Phyrexian Mite
//	 artifact creature tokens with toxic 1 and 'This token can't block.'"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "929c67aa-49ff-4401-85e8-76990fe1a369",
		Name:            "Basilica Shepherd",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 2, CreatureTokens: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Basilica Shepherd — create two Phyrexian Mites", Do(CreateToken{Template: PhyrexianMiteToken(), N: 2})),
		},
	})
}
