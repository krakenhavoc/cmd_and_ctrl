package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const rumorGathererLabel = "Rumor Gatherer — alliance: scry 1; the second time this resolves this turn, draw a card instead"

// Rumor Gatherer — Creature — Elf Wizard {1}{W}{W}, 2/1 (Streets of New
// Capenna, #5491):
//
//	"Alliance — Whenever another creature you control enters, scry 1. If
//	 this is the second time this ability has resolved this turn, draw a
//	 card instead."
//
// Captain Michael Burnham's resolution tally (Game.ResolvedThisTurn): it
// counts the resolution in progress, so the second resolution reads exactly
// 2 and draws instead of scrying; the first, third and later scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ceca713-38de-4428-a4f1-2146f6e98393",
		Name:         "Rumor Gatherer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherCreatureEnteredUnderYourControl, rumorGathererLabel,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if effects.SncSecondResolutionThisTurn(g, item, rumorGathererLabel) {
						return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
					}
					return effects.Scry{N: 1}.Apply(ctx)
				}),
		},
	})
}
