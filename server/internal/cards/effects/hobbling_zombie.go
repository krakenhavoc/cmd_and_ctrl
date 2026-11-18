package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hobbling Zombie — Creature — Zombie {2}{B}, 2/2 (Midnight Hunt):
//
//	"Deathtouch
//	 When this creature dies, create a 2/2 black Zombie creature token
//	 with decayed."
//
// Deathtouch is the printed keyword. The death trigger is a
// leaves-the-battlefield ability, so it looks back (CR 603.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9d1c3ec1-00a4-4890-93a3-e3cf137ca7e4",
		Name:            "Hobbling Zombie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Hobbling Zombie — create a 2/2 black Zombie creature token with decayed",
				Do(CreateToken{Template: TokenCard("2/2 black Zombie with decayed"), N: 1})),
		},
	})
}
