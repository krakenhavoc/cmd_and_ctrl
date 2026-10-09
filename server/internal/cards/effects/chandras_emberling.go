package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chandra's Emberling — Creature — Gremlin Elemental {2}{R}, 2/2:
//
//	"Haste
//	 Whenever you cast a noncreature spell, put a +1/+1 counter on
//	 this creature."
//
// The trigger fires on the cast (CR 601.2i), so a countered spell still
// grows it. If the Emberling has left the battlefield by the time the
// ability resolves there is nothing to put the counter on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d3576617-7867-4c3d-aa31-d6742483465b",
		Name:            "Chandra's Emberling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Chandra's Emberling — put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
