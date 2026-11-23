package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Baldur's Gate — Legendary Land — Gate (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Add X mana of any one color, where X is the number of other
//	 Gates you control."
//
// A mana ability (CR 605.1a) with an amount counted as it is activated: Mona
// Lisa's "N mana of any one color" (one colour pick, N mana of it), with N the
// Gates the controller has besides this one. With no other Gate it adds
// nothing, but still costs {2} and taps.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da307ea2-4df7-4d6b-be0f-9dc6ac93db61",
		Name:         "Baldur's Gate",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:  effects.ManaAbilityCost{Tap: true, Mana: "{2}"},
				Label: "{2}, {T}: Add X mana of any one color, where X is the number of other Gates you control",
				ProducedAmount: func(g *game.Game, controller, source uuid.UUID) game.ManaAmount {
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.InstanceID != source && c.Controller == controller && effects.CardHasSubtype(c, "Gate") {
							n++
						}
					}
					return game.ManaAmount{Options: game.AllColors, N: n, OneColor: true}
				},
			},
		},
	})
}
