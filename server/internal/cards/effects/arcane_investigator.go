package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "19b229c4-1c21-4c7c-a25e-4e8c1cd4ed4b",
		Name:         "Arcane Investigator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Search the Room — {5}{U}: Roll a d20.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{5}{U}"),
			Effect:  Do(RollDice{Sides: 20, N: 1, Then: arcaneInvestigatorTable}),
		}},
	})
}

// arcaneInvestigatorTable is the results table, after the d20.
var arcaneInvestigatorTable = DiceTail("dice/arcane-investigator-table", func(ctx *Context, rolls []int) error {
	if rolls[0] >= 10 {
		return LookAtTopThenTakeOneRestOnBottom(3, "Arcane Investigator — put one of those cards into your hand")(ctx.Game, ctx.Item)
	}
	return DrawCards{N: 1}.Apply(ctx)
})
