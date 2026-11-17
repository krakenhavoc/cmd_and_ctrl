package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flywheel Racer — Artifact — Vehicle {2}, 3/2 (March of the Machine):
//
//	"Vigilance
//	 {T}: Add one mana of any color. Activate only if this permanent is a
//	 creature.
//	 Crew 1"
//
// Veloheart Bike's mana ability with a Condition (the "activate only if" gate is
// checked before the cost is paid, CR 602.5). The Vehicle is a creature only
// once crewed, so an idle Racer taps for nothing; a crewed one that came under
// its controller's control this turn is stopped by summoning sickness
// (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0399b932-2a59-471c-828e-2b16614e0dae",
		Name:            "Flywheel Racer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color. Activate only if this permanent is a creature.",
			Condition: func(g *game.Game, _, source uuid.UUID) bool {
				c, ok := g.LookupCardForEffect(source)
				return ok && c.IsCreature()
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Flywheel Racer"),
		}},
	})
}
