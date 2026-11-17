package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3ceca713-38de-4428-a4f1-2146f6e98393",
		Name:         "Rumor Gatherer",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEnteredUnderYourControl, rumorGathererLabel,
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if sncSecondResolutionThisTurn(g, item, rumorGathererLabel) {
						return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
					}
					return Scry{N: 1}.Apply(ctx)
				}),
		},
	})
}
