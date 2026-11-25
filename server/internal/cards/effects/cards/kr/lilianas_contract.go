package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana's Contract — Enchantment {3}{B}{B} (Secret Lair Drop):
//
//	"When this enchantment enters, you draw four cards and you lose 4 life.
//	 At the beginning of your upkeep, if you control four or more Demons with
//	 different names, you win the game."
//
// The win is an intervening "if" (CR 603.4): checked as the upkeep begins and
// again as the trigger resolves. Demons are read through the layered view (a
// Demon by subtype, so a changeling counts), and "different names" counts the
// distinct names among them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6858702a-e563-4dcb-a50d-65406cea7c86",
		Name:         "Liliana's Contract",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Liliana's Contract — you draw four cards and you lose 4 life",
				effects.Do(effects.DrawCards{N: 4}, effects.GainLife{Amount: -4})),
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && demonsWithDifferentNames(g, source.Controller) >= 4
			}, "Liliana's Contract — you win the game",
				effects.Do(effects.If{
					Cond: effects.CondBoard{Name: "you control four or more Demons with different names", Test: func(ctx *effects.Context) bool {
						return demonsWithDifferentNames(ctx.Game, ctx.Controller()) >= 4
					}},
					Then: []effects.Applier{effects.WinTheGame{}},
				})),
		},
	})
}

// demonsWithDifferentNames counts the distinct names among the Demon creatures
// `controller` controls.
func demonsWithDifferentNames(g *game.Game, controller uuid.UUID) int {
	names := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.HasSubtypeWord("Demon") {
			names[c.Name] = true
		}
	}
	return len(names)
}
