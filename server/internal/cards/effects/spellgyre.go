package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spellgyre — Instant {2}{U}{U} (Bloomburrow, #4467):
//
//	"Choose one —
//	 • Counter target spell.
//	 • Surveil 2, then draw two cards. (To surveil 2, look at the top two
//	   cards of your library, then put any number of them into your
//	   graveyard and the rest on top of your library in any order.)"
//
// A modal spell (CR 700.2). The first mode is Amazing Acrobatics'
// counter; a spell gone by resolution is not countered (CR 608.2b). The
// second is Diresight's order: the surveil prompt carries the draw as a
// keyed tail (Water Whip's "draw N"), so both cards are drawn from the
// library as the player arranged it (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7db31623-5dea-4165-8593-fbac568bf5b4",
		Name:         "Spellgyre",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Counter target spell.", TargetSpell("target spell"), CounterTheModesTarget),
			ModeWithPurpose(ModeDoing("Surveil 2, then draw two cards.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return Surveil{Player: item.Controller, N: 2, ThenTail: ctx.TailN(avB34ThenDraw, 2)}.Apply(ctx)
				}), game.Purpose{Draws: 2}),
		),
	})
}
