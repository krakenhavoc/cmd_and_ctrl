package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exploding Barrel — {3} Artifact (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"{T}: Add one mana of any color. Put a pressure counter on this
//	 artifact.
//	 {8}, {T}, Sacrifice this artifact: It deals 20 damage to target
//	 creature. This ability costs {1} less to activate for each pressure
//	 counter on this artifact. Activate only as a sorcery."
//
// The mana ability is Twitching Doll's: any colour, with the counter as
// its rider through the ordinary placement, so a counter doubler sees it.
// It is still a mana ability (CR 605.1a), resolving at once. The second
// ability is Diary of Dreams' self-discount (CostsLessEach over the
// pressure counters), counted as the cost is determined (CR 601.2f,
// CR 602.2b), before the Barrel is sacrificed to pay it, never below {0}
// generic while the {T} and the sacrifice are still paid. The damage is
// the Barrel's as it last existed (CR 608.2h), and a target that left
// takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	const pressure = "pressure"
	Register(Spec{
		OracleID:     "82800275-1edf-4aa0-af9e-59dea82d7c6a",
		Name:         "Exploding Barrel",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color. Put a pressure counter on this artifact.",
			Rider: func(g *game.Game, _ uuid.UUID, source uuid.UUID) error {
				return g.AddCounterForEffect(source, pressure, 1)
			},
		}},
		Activated: []ActivatedAbility{{
			Label:        "{8}, {T}, Sacrifice this artifact: It deals 20 damage to target creature. This ability costs {1} less to activate for each pressure counter on this artifact.",
			Cost:         Plus(ManaCost("{8}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			CostModifiers: []game.CostModifier{
				CostsLessEach(CountersOnSource(pressure),
					"This ability costs {1} less to activate for each pressure counter on this artifact."),
			},
			Targets: TargetCreature("target creature"),
			Effect:  DealDamageToTheTarget(20),
		}},
	})
}
