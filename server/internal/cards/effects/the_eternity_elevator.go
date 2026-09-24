package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Eternity Elevator — Legendary Artifact — Spacecraft {5}:
//
//	"{T}: Add {C}{C}{C}.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a
//	 sorcery.)
//	 20+ | {T}: Add X mana of any one color, where X is the number of
//	 charge counters on The Eternity Elevator."
//
// Station is Galvanizing Sawship's shape: the printed reminder text
// IS the Station() activated ability (station.go, ADR 0071 decision
// 2), and the "20+" line is an ordinary mana-ability Condition gate
// reading the same charge counters, with ProducedOneColor supplying
// the dynamic amount. Two {T} abilities on one permanent share the
// tap symbol exactly as they do on Prismatic Lens and Liquimetal
// Torque; the 20+ ability is simply unavailable below the threshold.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11323af4-b8b8-4ca9-932f-377c7fd77dea",
		Name:         "The Eternity Elevator",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}{C}{C}",
				Label:    "Add {C}{C}{C}",
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				Condition:    eternityElevatorAt20,
				ProducedFunc: ProducedOneColor(chargeCountersOn),
				Label:        "20+ — Add X mana of any one color, where X is the number of charge counters on The Eternity Elevator",
			},
		},
		Activated: []ActivatedAbility{Station()},
	})
}

// eternityElevatorAt20 is the "20+" gate; chargeCountersOn
// (astral_cornucopia.go) supplies the amount once it's open.
func eternityElevatorAt20(g *game.Game, _, source uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(source)
	return ok && c.Counters[game.CounterCharge] >= 20
}
