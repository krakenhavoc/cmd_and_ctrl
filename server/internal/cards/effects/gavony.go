package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gavony — Plane — Innistrad (#4870):
//
//	"All creatures have vigilance.
//	 Whenever chaos ensues, creatures you control gain indestructible
//	 until end of turn."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The static is a layer-6 grant to every creature. The chaos ability
// (CR 311.7) locks the set of creatures the planar controller controls as it
// resolves (CR 611.2c, the rulings), so a creature they gain later in the
// turn is not indestructible and one they lose keeps it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2387edac-765b-4502-bd3d-88be8ba7d0a4",
		Name:         "Gavony",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{b16GrantKeywords(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsCreature()
		}, "vigilance")},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Gavony — creatures you control gain indestructible until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"indestructible"},
						Label: "Gavony — indestructible until end of turn"}.Apply(NewContext(g, item))
				}),
		},
	})
}
