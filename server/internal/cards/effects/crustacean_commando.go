package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crustacean Commando — {1}{U} Creature — Crab Mutant Soldier 0/3
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"When this creature enters, create a Mutagen token. (It's an artifact
//	 with "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target
//	 creature. Activate only as a sorcery.")"
//
// The Mutagen token is the catalog's CR 111.10v template (MutagenToken, #4676).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b24a87af-407f-4c58-80b4-caab9c65a233",
		Name:         "Crustacean Commando",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Crustacean Commando — create a Mutagen token",
				Do(CreateToken{Template: MutagenToken(), N: 1})),
		},
	})
}
