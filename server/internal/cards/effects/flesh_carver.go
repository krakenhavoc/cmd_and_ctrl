package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flesh Carver — Creature — Human Wizard {2}{B}, 2/2 (Commander 2014,
// #5553):
//
//	"Intimidate (This creature can't be blocked except by artifact
//	 creatures and/or creatures that share a color with it.)
//	 {1}{B}, Sacrifice another creature: Put two +1/+1 counters on this
//	 creature.
//	 When this creature dies, create an X/X black Horror creature token,
//	 where X is this creature's power."
//
// The sacrifice is a cost of "another" creature (SacrificeAnotherN,
// Hungry Ghoul's outlet), so the Carver can't pay with itself. X is its
// power as it last existed on the battlefield (CR 603.10a, b13LastKnownPower
// through eerieDiesWithPower), counters and pumps included, fixed as the
// ability triggers; a negative power makes a 0/0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8fcaaf1c-c99b-4f28-bf95-7115a11e803d",
		Name:            "Flesh Carver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature: Put two +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, SourceCounters: 2},
			Cost:    Plus(ManaCost("{1}{B}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect:  plusOneCountersOnThis(2),
		}},
		Triggered: []game.TriggeredAbility{
			eerieDiesWithPower("Flesh Carver — create an X/X black Horror, X its power", nil,
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Template: c14BlackHorror(item.Params.Amount), N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
