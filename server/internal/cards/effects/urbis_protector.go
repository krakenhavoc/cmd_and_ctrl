package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urbis Protector — Creature — Human Cleric {4}{W}{W}, 1/1 (Ravnica: Clue
// Edition, S82):
//
//	"When this creature enters, create a 4/4 white Angel creature token
//	 with flying."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "564dd5e7-6c1e-4959-a594-60e08552aec9",
		Name:         "Urbis Protector",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Urbis Protector — create a 4/4 white Angel creature token with flying",
				Do(CreateToken{Template: TokenCard("4/4 white Angel with flying"), N: 1})),
		},
	})
}
