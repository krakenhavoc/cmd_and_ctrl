package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sarek, Federation Ambassador — Legendary Creature — Vulcan Advisor
// {1}{G}{W}{U}, 0/5 (Star Trek, #4973):
//
//	"Federation — At the beginning of combat on your turn, draw a card
//	 for every five creature types among non-Borg creatures you control.
//	 If you draw three or more cards this way, put two +1/+1 counters on
//	 each creature you control."
//
// "Federation" is a flavour word (CR 207.2d). The count is Hoshi Sato,
// Exolinguist's (trekB7CreatureTypesAmongNonBorgYouControl), read as the
// trigger resolves (CR 608.2h): each creature type once, a changeling
// every type, a Borg nothing. One card per full five types.
//
// "Draw three or more cards this way" counts the cards actually drawn: the
// player's draws this turn are tallied before the draw is filed and again
// after it, so a draw from an empty library is not one of them. The
// counters wait for any prompt the draw asks (a dredge offer, CR 608.2c),
// and go on each creature you control then as one placement of two each
// (b11PutCountersOnEachCreatureYouControl).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72c51539-5f15-4636-8cb8-b688cf53f431",
		Name:         "Sarek, Federation Ambassador",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Sarek, Federation Ambassador — draw a card for every five creature types among non-Borg creatures you control",
				Do(
					Instruction(func(ctx *Context) error {
						n := trekB7CreatureTypesAmongNonBorgYouControl(ctx.Game, ctx.Item) / 5
						if n <= 0 {
							return nil
						}
						ctx.RecordStep(sarekDrawsBefore, nil, len(ctx.Game.DrawnThisTurn[ctx.Controller()]))
						return DrawCards{N: n}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						rec := ctx.StepRecord(sarekDrawsBefore)
						if len(rec.Ns) == 0 {
							return nil
						}
						if len(ctx.Game.DrawnThisTurn[ctx.Controller()])-rec.N() < 3 {
							return nil
						}
						return b11PutCountersOnEachCreatureYouControl(ctx.Game, ctx.Item, 2)
					}),
				)),
		},
	})
}

// sarekDrawsBefore files how many cards the controller had drawn this turn
// before Sarek's draw, so the second sentence can count the cards drawn
// this way. An on-disk step key.
const sarekDrawsBefore = "sarek-federation-ambassador/draws-before"
