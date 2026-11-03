package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slithering Cryptid — {2}{G/U} Creature — Fish Mutant 2/3 (Teenage
// Mutant Ninja Turtles, #4457):
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
		OracleID:     "f570bac8-9987-4963-af02-476d18abc847",
		Name:         "Slithering Cryptid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Slithering Cryptid — create a Mutagen token",
				Do(CreateToken{Template: MutagenToken(), N: 1})),
		},
	})
}
