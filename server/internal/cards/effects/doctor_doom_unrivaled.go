package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doctor Doom, Unrivaled — Legendary Creature — Human Sorcerer Villain
// {2}{B}{B}, 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Lifelink
//	 {T}: You draw a card and lose 1 life. Then if your library has no
//	 cards in it, you win the game. (You win even if you have 0 life or
//	 didn't draw a card.)"
//
// One ability resolving in printed order: draw, lose 1, then the library
// is read as it now stands. The win is the Thassa's Oracle one (CR
// 104.2b): checked after the draw and the life loss, whether or not a card
// was drawn, and a "can't win" effect such as Platinum Angel stops it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb3dd50b-ce78-4727-b174-06f99fc5ec74",
		Name:            "Doctor Doom, Unrivaled",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: You draw a card and lose 1 life. Then if your library has no cards in it, you win the game.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  Do(DrawCards{N: 1}, GainLife{Amount: -1}, Instruction(msc5WinIfLibraryEmpty)),
		}},
	})
}

// msc5WinIfLibraryEmpty is "then if your library has no cards in it, you
// win the game".
func msc5WinIfLibraryEmpty(ctx *Context) error {
	me := ctx.Controller()
	p := ctx.Game.PlayerByIDForEffect(me)
	if p == nil || p.Library.Size() > 0 {
		return nil
	}
	return WinTheGame{Player: me}.Apply(ctx)
}
