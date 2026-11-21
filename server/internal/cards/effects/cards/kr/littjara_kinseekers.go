package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Littjara Kinseekers — Creature — Shapeshifter {3}{U}, 2/4 (Jumpstart 2022, #5948):
//
//	"Changeling (This card is every creature type.)
//	 When this creature enters, if you control three or more creatures that
//	 share a creature type, put a +1/+1 counter on this creature, then
//	 scry 1."
//
// The intervening "if" (CR 603.4) is read at trigger time and again as the
// ability resolves. Creature types are read through the layers by
// game.CreatureTypesOf, so a changeling (the Kinseekers included) shares
// every type with every creature that has one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "67ff2880-1998-48e9-88c7-2aa012697401",
		Name:            "Littjara Kinseekers",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Scry: 1},
		PrintedKeywords: []string{game.KeywordChangeling},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control three or more creatures that share a creature type",
				Test: func(ctx *effects.Context) bool {
					return threeShareAType(ctx.Game, ctx.Controller())
				},
			}, effects.WhenThisEnters("Littjara Kinseekers — put a +1/+1 counter on this creature, then scry 1",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
					}),
					effects.Scry{N: 1},
				))),
		},
	})
}

// threeShareAType reports whether `player` controls three or more
// creatures that have one creature type in common.
func threeShareAType(g *game.Game, player uuid.UUID) bool {
	counts := map[string]int{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || !c.IsCreature() {
			continue
		}
		cc := c
		for _, t := range game.CreatureTypesOf(&cc) {
			counts[t]++
			if counts[t] >= 3 {
				return true
			}
		}
	}
	return false
}
