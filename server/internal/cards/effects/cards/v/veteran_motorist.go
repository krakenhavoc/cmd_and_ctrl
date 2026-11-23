package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veteran Motorist — {R}{W} Creature — Dwarf Pilot 3/1:
//
//	"When this creature enters, scry 2.
//	 Whenever this creature crews a Vehicle, that Vehicle gets +1/+1
//	 until end of turn."
//
// The crew trigger is WhenThisCrewsOrSaddles with only the crew event
// (ADR 0186 section 2). It goes on the stack as the creature is tapped for
// the cost, so it resolves BEFORE the crew ability (CR 208.3a's own
// example is this card): the Vehicle is not yet a creature, and the pump
// is on the permanent, ready for when it becomes one. "That Vehicle" is
// the object the Motorist paid for; one that left in response is a new
// object and gets nothing (CR 400.7).
//
// No simplification.
func init() {
	const label = "Veteran Motorist — that Vehicle gets +1/+1 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "3e6f51ae-7239-4fcc-919f-aa2437e8ef15",
		Name:         "Veteran Motorist",
		Purpose:      game.Purpose{Scry: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Veteran Motorist — scry 2", effects.Do(effects.Scry{N: 2})),
			effects.WhenThisCrewsOrSaddles(label, true, false, func(g *game.Game, item *game.StackItem) error {
				id, ok := effects.CrewedOrSaddledObject(g, item)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
