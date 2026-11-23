package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Speedway Fanatic — {1}{R} Creature — Human Pilot 2/1:
//
//	"Haste
//	 Whenever this creature crews a Vehicle, that Vehicle gains haste
//	 until end of turn."
//
// Canyon Vaulter's shape with only the crew event. The haste is on the
// Vehicle's object, so a Vehicle that came in this turn can attack once it
// is crewed; one that left in response is a new object and gains nothing
// (CR 400.7).
//
// No simplification.
func init() {
	const label = "Speedway Fanatic — that Vehicle gains haste until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "5e7e1b23-ac4b-4f95-9388-feb099d6779f",
		Name:            "Speedway Fanatic",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisCrewsOrSaddles(label, true, false, func(g *game.Game, item *game.StackItem) error {
				id, ok := effects.CrewedOrSaddledObject(g, item)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
