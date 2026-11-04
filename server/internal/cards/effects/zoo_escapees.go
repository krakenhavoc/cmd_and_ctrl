package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zoo Escapees — {1}{G} Creature — Boar Rhino 2/2 (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"When this creature leaves the battlefield, create a Mutagen token.
//	 (It's an artifact with "{1}, {T}, Sacrifice this token: Put a +1/+1
//	 counter on target creature. Activate only as a sorcery.")"
//
// A leaves-the-battlefield trigger (CR 603.6c): to any zone, its
// controller as it left making the token (CR 603.3a). The Mutagen token is the catalog's CR 111.10v template (MutagenToken, #4676).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "107e9ada-4e2e-4032-addd-274bca956621",
		Name:         "Zoo Escapees",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("Zoo Escapees — create a Mutagen token",
				Do(CreateToken{Template: MutagenToken(), N: 1})),
		},
	})
}
