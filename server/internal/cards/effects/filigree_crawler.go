package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Filigree Crawler — Artifact Creature — Insect {4}, 2/2 (Aether Revolt,
// #5536):
//
//	"When this creature dies, create a 1/1 colorless Thopter artifact
//	 creature token with flying."
//
// Thopter Mechanic's death trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f74095aa-232f-432e-a899-86e9463ab5be",
		Name:         "Filigree Crawler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Filigree Crawler — create a 1/1 colorless Thopter artifact creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}
