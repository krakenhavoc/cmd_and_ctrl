package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reckless Velocitaur — {3}{R} Creature — Minotaur Pilot 3/3:
//
//	"Whenever this creature saddles a Mount or crews a Vehicle during your
//	 main phase, that Mount or Vehicle gets +2/+0 and gains trample until
//	 end of turn."
//
// Canyon Vaulter's shape (crew_triggers.go) with a pump and a different
// keyword. The +2/+0 and the trample are two continuous effects, both
// applied to the object the Velocitaur paid for as the trigger resolves,
// which is before the crew ability does; the power is a layer-7c
// modifier, so it sits on the Vehicle's printed power once it is a
// creature.
//
// No simplification.
func init() {
	const label = "Reckless Velocitaur — that Mount or Vehicle gets +2/+0 and gains trample until end of turn"
	Register(Spec{
		OracleID:     "367420cb-e77b-4d0c-9a83-0deee28f6d36",
		Name:         "Reckless Velocitaur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			DuringYourMainPhase(WhenThisCrewsOrSaddles(label, true, true, func(g *game.Game, item *game.StackItem) error {
				id, ok := CrewedOrSaddledObject(g, item)
				if !ok {
					return nil
				}
				ctx := NewContext(g, item)
				if err := (BoostUntilEOT{Target: id, Power: 2, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: label}.Apply(ctx)
			})),
		},
	})
}
