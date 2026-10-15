package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defiant Survivor — Creature — Human Survivor {2}{G}, 3/2:
//
//	"Survival — At the beginning of your second main phase, if this
//	 creature is tapped, manifest dread. (Look at the top two cards of
//	 your library. Put one onto the battlefield face down as a 2/2
//	 creature and the other into your graveyard. Turn it face up any time
//	 for its mana cost if it's a creature card.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b62c53a-46a2-4c57-9eab-0c6706795b4e",
		Name:         "Defiant Survivor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Survival("Defiant Survivor — manifest dread", Do(ManifestDread{})),
		},
	})
}
