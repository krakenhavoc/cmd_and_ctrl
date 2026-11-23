package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gearshift Ace — {1}{W} Creature — Dwarf Pilot 2/1:
//
//	"First strike
//	 Whenever this creature crews a Vehicle, that Vehicle gains first
//	 strike until end of turn."
//
// Canyon Vaulter's shape with only the crew event: the trigger resolves
// before the crew ability, so the Vehicle gains first strike while it is
// still an artifact and keeps it once it becomes a creature. A Vehicle
// that left in response is a new object and gains nothing (CR 400.7).
//
// No simplification.
func init() {
	const label = "Gearshift Ace — that Vehicle gains first strike until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "dde61c4c-1815-45b3-9e68-3097c28c363c",
		Name:            "Gearshift Ace",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisCrewsOrSaddles(label, true, false, func(g *game.Game, item *game.StackItem) error {
				id, ok := effects.CrewedOrSaddledObject(g, item)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
