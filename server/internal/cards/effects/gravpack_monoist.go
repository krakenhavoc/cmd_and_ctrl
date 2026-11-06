package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gravpack Monoist — Creature — Human Scout {2}{B}, 2/1 (Edge of
// Eternities, #4461):
//
//	"Flying
//	 When this creature dies, create a tapped 2/2 colorless Robot
//	 artifact creature token."
//
// Beamsaw Prospector's dies trigger making Melded Moxite's Robot through
// CreateTokenAdvanced: the token enters tapped, under the control of the
// Monoist's controller as it last existed (CR 603.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9c009b3f-6580-4ab8-84e8-d4cf7a456a02",
		Name:            "Gravpack Monoist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Gravpack Monoist — create a tapped 2/2 colorless Robot artifact creature token",
				Do(CreateTokenAdvanced{Spec: Token(TokenCard("2/2 colorless Robot artifact")).EntersTapped(), N: 1})),
		},
	})
}
