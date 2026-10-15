package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Canyon Vaulter — {1}{W} Creature — Kor Pilot 3/1:
//
//	"Whenever this creature saddles a Mount or crews a Vehicle during your
//	 main phase, that Mount or Vehicle gains flying until end of turn."
//
// One trigger watching both events (WhenThisCrewsOrSaddles, ADR 0186 §2),
// narrowed to your own main phase when the event happens
// (DuringYourMainPhase). The trigger is put on the stack as the cost is
// paid, so it resolves BEFORE the crew ability: the Vehicle gains flying
// while it is still an artifact, and keeps it once it becomes a creature.
// "That Mount or Vehicle" is the object the Vaulter paid for; a Vehicle
// that left in response is a new object and gains nothing (CR 400.7).
// Two Vaulters tapped for one crew are two events and two triggers.
//
// No simplification.
func init() {
	const label = "Canyon Vaulter — that Mount or Vehicle gains flying until end of turn"
	Register(Spec{
		OracleID:     "829e121f-d2fa-40ef-afc7-9b2072a949d7",
		Name:         "Canyon Vaulter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			DuringYourMainPhase(WhenThisCrewsOrSaddles(label, true, true, func(g *game.Game, item *game.StackItem) error {
				id, ok := CrewedOrSaddledObject(g, item)
				if !ok {
					return nil
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: label}.Apply(NewContext(g, item))
			})),
		},
	})
}
