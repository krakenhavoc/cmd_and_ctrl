package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// High Market — Land (EDHREC rank 430):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice a creature: You gain 1 life."
//
// A sacrifice outlet on a land — the one every aristocrats deck runs
// because it costs no card slot and answers a theft or a Chaos Warp
// at instant speed. The life is incidental; the outlet is the card.
//
// The sacrifice is a COST (Goblin Bombardment's clause with a tap
// added), paid at announce, so the dies-triggers it causes go on the
// stack above the ability and resolve first. The source itself is
// not a creature, so "a creature" needs no "another".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86fb3749-37d6-48a6-8524-71e996850307",
		Name:         "High Market",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{T}, Sacrifice a creature: You gain 1 life.",
			Cost:  Plus(TapCost(), SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
