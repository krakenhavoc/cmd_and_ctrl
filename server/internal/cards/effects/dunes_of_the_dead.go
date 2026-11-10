package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dunes of the Dead — Land — Desert:
//
//	"{T}: Add {C}.
//	 When this land is put into a graveyard from the battlefield, create
//	 a 2/2 black Zombie creature token."
//
// A leaves-the-battlefield trigger (CR 603.10a) that looks back at the
// land; any move from the battlefield to a graveyard counts, however it
// got there (destroyed, sacrificed, or legend-ruled).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c761f71c-785c-4533-a2b7-2da3667688b8",
		Name:         "Dunes of the Dead",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Dunes of the Dead — create a 2/2 black Zombie creature token",
				Do(CreateToken{Template: BlackZombieToken(), N: 1})),
		},
	})
}
