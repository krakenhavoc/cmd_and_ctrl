package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcane Investigator — Creature — Elf Wizard {1}{U}, 2/1 (Adventures in
// the Forgotten Realms, #5499):
//
//	"Search the Room — {5}{U}: Roll a d20.
//	 1—9 | Draw a card.
//	 10—20 | Look at the top three cards of your library. Put one of them
//	 into your hand and the rest on the bottom of your library in any order."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19b229c4-1c21-4c7c-a25e-4e8c1cd4ed4b",
		Name:         "Arcane Investigator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Search the Room — {5}{U}: Roll a d20.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{5}{U}"),
			Effect:  effects.Do(effects.RollDice{Sides: 20, N: 1, Then: arcaneInvestigatorTable}),
		}},
	})
}

// arcaneInvestigatorTable is the results table, after the d20.
var arcaneInvestigatorTable = effects.DiceTail("dice/arcane-investigator-table", func(ctx *effects.Context, rolls []int) error {
	if rolls[0] >= 10 {
		return effects.LookAtTopThenTakeOneRestOnBottom(3, "Arcane Investigator — put one of those cards into your hand")(ctx.Game, ctx.Item)
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
})
