package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Herald of Hadar — Creature — Human Warlock {4}{B}, 4/4 (Adventures in
// the Forgotten Realms, #5499):
//
//	"Circle of Death — {5}{B}: Roll a d20.
//	 1—9 | Each opponent loses 2 life.
//	 10—19 | Each opponent loses 2 life and you gain 2 life.
//	 20 | Each opponent loses 2 life and you gain 2 life. Create two
//	 Treasure tokens."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "25ab42a5-ec1a-4510-b957-1d4f14fa10b0",
		Name:         "Herald of Hadar",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Circle of Death — {5}{B}: Roll a d20.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{5}{B}"),
			Effect:  Do(RollDice{Sides: 20, N: 1, Then: heraldOfHadarTable}),
		}},
	})
}

// heraldOfHadarTable is the results table, after the d20.
var heraldOfHadarTable = DiceTail("dice/herald-of-hadar-table", func(ctx *Context, rolls []int) error {
	if err := (EachOpponentLosesLife{N: 2}).Apply(ctx); err != nil {
		return err
	}
	if rolls[0] >= 10 {
		if err := (GainLife{Amount: 2}).Apply(ctx); err != nil {
			return err
		}
	}
	if rolls[0] == 20 {
		return treasureTokens(ctx.Game, ctx.Controller(), 2)
	}
	return nil
})
