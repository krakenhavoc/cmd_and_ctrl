package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primordial Gnawer — Creature — Insect Horror {4}{B}, 5/2:
//
//	"When this creature dies, discover 3."
//
// The simplest discover in print (CR 701.57, ADR 0099): a dies trigger
// whose whole effect is the keyword action, so it is the example the
// roadmap's discover row points at. The free cast is a grant taken
// after the trigger resolves, with the timing CR 608.2g allows, and
// passing without casting puts the card into your hand — see
// game/discover.go.
func init() {
	Register(Spec{
		OracleID:     "80ca6596-9763-4310-b30b-7153ade9f567",
		Name:         "Primordial Gnawer",
		Completeness: CompletenessFull,
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Primordial Gnawer — discover 3", DiscoverN(3)),
		},
	})
}
