package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infinite Diversity — Sorcery {7}{R}{R} (Star Trek, #4973):
//
//	"Federation — Exile target instant or sorcery card from your
//	 graveyard. Copy that card X times, where X is the number of creature
//	 types among non-Borg creatures you control. You may cast the copies
//	 without paying their mana costs. Exile Infinite Diversity."
//
// Federation is an ability word with no rules meaning (CR 207.2c). The
// exile goes through the CR 614 window, and only a card that really
// reached exile is copied (Mizzix's Mastery's shape, #3258). X is counted
// as the copies are made: each creature type once, however many of your
// non-Borg creatures share it, read through the layers. A changeling is
// every creature type, Borg included, so it adds nothing. The copies (CR
// 707.12) sit in exile with a free-cast offer each and are cast while the
// spell is still resolving (CR 608.2g); one not cast ceases to exist.
// The spell then exiles itself. With X = 0 the card is exiled and nothing
// is copied. A target that left in response makes the spell do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7d8ed66-d66e-42fd-9c3e-8c2d8194e009",
		Name:         "Infinite Diversity",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target instant or sorcery card from your graveyard",
			Or(Instant(), Sorcery()), YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return ctx.Game.ExileCardKeyedForEffect(t.ID, NewContext(ctx.Game, item).ItemTail(trekB16InfiniteDiversityTail))
		},
	})
}
