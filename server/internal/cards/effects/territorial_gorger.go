package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Territorial Gorger — Creature — Gremlin {3}{R}, 2/2:
//
//	"Trample
//	 Whenever you get one or more {E} (energy counters), this creature
//	 gets +2/+2 until end of turn."
//
// ADR 0129 §6 (#1995): one trigger per placement of energy, however
// many counters it put on (CR 603.2c). The boost goes to the permanent
// that triggered, not a new object with its name (CR 400.7).
//
// No simplification.
func init() {
	const label = "Territorial Gorger — +2/+2 until end of turn"
	Register(Spec{
		OracleID:        "fcf9c013-7a79-42b1-8aee-6d40ff475fbf",
		Name:            "Territorial Gorger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverYouGetEnergy(label, thisStillHere(func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 2, Label: label}.Apply(NewContext(g, item))
			})),
		},
	})
}
