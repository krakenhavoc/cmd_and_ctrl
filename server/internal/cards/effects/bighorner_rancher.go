package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bighorner Rancher — Creature — Human Ranger {4}{G}, 2/5 (Fallout,
// #5469):
//
//	"Vigilance
//	 {T}: Add an amount of {G} equal to the greatest power among
//	 creatures you control.
//	 Sacrifice this creature: You gain life equal to the greatest
//	 toughness among other creatures you control."
//
// The mana ability is Topiary Lecturer's amount of {G} over Orcish
// Siegemaster's "greatest power among creatures you control"
// (b42GreatestPowerControlledBy), the Rancher itself included, read on
// the layered board as the ability resolves. The sacrifice is Arbor
// Adherent's "greatest toughness among other creatures you control"
// (arborAdherentX), read as the ability resolves; the Rancher was
// sacrificed as the cost, so it is never among them, and with no other
// creature the life gained is 0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "826e6a6f-3b66-4225-87ee-04e345e047b3",
		Name:            "Bighorner Rancher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			Label:          "Add an amount of {G} equal to the greatest power among creatures you control",
			ProducedAmount: bighornerRancherGreen,
		}},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: You gain life equal to the greatest toughness among other creatures you control.",
			Cost:    SacrificeThis(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := arborAdherentX(g, item.Controller, item.SourceCardID)
				if n <= 0 {
					return nil
				}
				return GainLife{Player: item.Controller, Amount: n}.Apply(NewContext(g, item))
			},
		}},
	})
}

// bighornerRancherGreen is the Rancher's {G} for each point of the
// greatest power among creatures its controller controls.
func bighornerRancherGreen(g *game.Game, controller, _ uuid.UUID) game.ManaAmount {
	g.RecomputeLayersIfStaleLocked()
	return game.ManaAmount{Options: []string{"G"}, N: b42GreatestPowerControlledBy(g, controller)}
}
