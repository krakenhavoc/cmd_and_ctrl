package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wedgelight Rammer — Artifact — Spacecraft {3}{W}, 3/4 (Edge of
// Eternities, #4461):
//
//	"When this Spacecraft enters, create a 2/2 colorless Robot artifact
//	 creature token.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 9+.)
//	 9+ | Flying, first strike"
//
// Pinnacle Kill-Ship's station lines (Station, SpacecraftAt and
// ThresholdKeywords, CR 702.184a and 721.2a) with Mechan Assembler's
// Robot made by an entry trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03259ab0-caa8-4620-b070-127e2c712252",
		Name:         "Wedgelight Rammer",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wedgelight Rammer — create a 2/2 colorless Robot artifact creature token",
				Do(CreateToken{Template: TokenCard("2/2 colorless Robot artifact"), N: 1})),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(9, 3, 4),
			ThresholdKeywords(9, "flying", "first strike"),
		),
	})
}
