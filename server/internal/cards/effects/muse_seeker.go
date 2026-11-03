package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Muse Seeker — Creature — Elf Wizard {1}{U}, 1/2 (Secrets of Strixhaven,
// #4456):
//
//	"Opus — Whenever you cast an instant or sorcery spell, draw a card. Then
//	 discard a card unless five or more mana was spent to cast that spell."
//
// Expressive Firedancer's Opus: a cast trigger (a copy is not cast, CR
// 707.10) that counts the mana spent as the spell is cast and carries it on
// the item, so the "unless" holds even when the spell has left the stack
// before the ability resolves. The draw always happens; the discard is the
// player's choice from the hand after the draw, so the drawn card is a legal
// discard, and it is skipped when five or more mana was spent. The trigger
// does not depend on Muse Seeker staying on the battlefield.
//
// No simplification.
func init() {
	const label = "Muse Seeker — Opus: draw a card, then discard a card unless five or more mana was spent"
	opus := On(game.EventCast, YouCast(Or(Instant(), Sorcery())), label,
		func(g *game.Game, item *game.StackItem) error {
			if item.Params.Amount >= 5 {
				return DrawCards{N: 1}.Apply(NewContext(g, item))
			}
			return b16DrawThenDiscard(g, item, 1)
		})
	opus.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:       "31f3d37f-e613-427f-b34c-cc87611ea078",
		Name:           "Muse Seeker",
		ReadsManaSpent: game.ReadsYourCasts,
		Completeness:   CompletenessFull,
		Triggered:      []game.TriggeredAbility{opus},
	})
}
