package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Campus Crier — Creature — Human Advisor {1}{W}, 3/1 (Reality
// Fracture):
//
//	"{1}, Exile this card from your graveyard: Empower Jace 2."
//
// ADR 0139 proof card: the keyword action from an ability that
// functions from the graveyard (CR 113.6), so the source has gone by
// the time it resolves and the action reads only its controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c1a9d60f-a469-48f9-8f1c-3ca5ae86bd64",
		Name:         "Campus Crier",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}, Exile this card from your graveyard: Empower Jace 2.",
			Cost:  Plus(ManaCost("{1}"), ExileThis()),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return EmpowerJace{N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
