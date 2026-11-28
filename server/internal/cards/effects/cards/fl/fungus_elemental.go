package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fungus Elemental — Creature — Fungus Elemental {3}{G}, 3/3 (Weatherlight, S82 older sets batch 780, #5629):
//
//	"{G}, Sacrifice a Forest: Put a +2/+2 counter on this creature. Activate only if this creature entered this turn."
//
// The cost is {G} and a Forest of the activator's choice (any land with the subtype, nonbasic included), and the ability
// may be activated only on the turn the creature entered (game.EnteredThisTurn, CR 400.7), as many times as the lands
// allow.
//
// No simplification.
func init() {
	const label = "{G}, Sacrifice a Forest: Put a +2/+2 counter on this creature. Activate only if this creature entered this turn."
	effects.Register(effects.Spec{
		OracleID:     "3d3c5dd5-ad4e-48ca-92b1-91e1f4ad5c3b",
		Name:         "Fungus Elemental",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Cost:      effects.Plus(effects.ManaCost("{G}"), effects.SacrificeN(1, "a Forest", effects.HasSubtype("Forest"))),
			Condition: effects.S82OldB780ThisEnteredThisTurn(),
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.AddCounter{Target: item.SourceCardID, Kind: "+2/+2", N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
