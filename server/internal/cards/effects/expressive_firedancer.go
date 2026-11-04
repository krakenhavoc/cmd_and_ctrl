package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expressive Firedancer — Creature — Human Sorcerer {1}{R}, 2/2 (Secrets of
// Strixhaven, #4456):
//
//	"Opus — Whenever you cast an instant or sorcery spell, this creature
//	 gets +1/+1 until end of turn. If five or more mana was spent to cast
//	 that spell, this creature also gains double strike until end of turn."
//
// Colorstorm Stallion's Opus: a cast trigger (a copy is not cast, CR
// 707.10) that counts the mana spent as the spell is cast and carries it on
// the item, so the "if" holds even when the spell is countered or has left
// the stack before the ability resolves. The +1/+1 happens either way; the
// double strike rides only five or more mana. Both are for the creature that
// triggered, so a Firedancer that left, or came back as a new object, gets
// neither (CR 400.7).
//
// No simplification.
func init() {
	const label = "Expressive Firedancer — Opus: +1/+1 until end of turn; with five or more mana spent, double strike too"
	opus := On(game.EventCast, YouCast(Or(Instant(), Sorcery())), label,
		func(g *game.Game, item *game.StackItem) error {
			if !sourceIsStillThisPermanent(g, item) {
				return nil
			}
			ctx := NewContext(g, item)
			if err := (BoostUntilEOT{
				Target:    item.SourceCardID,
				Power:     1,
				Toughness: 1,
				Label:     "Expressive Firedancer — +1/+1 until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if item.Params.Amount < 5 {
				return nil
			}
			return GrantKeywordUntilEOT{
				Target:   item.SourceCardID,
				Keywords: []string{"double strike"},
				Label:    "Expressive Firedancer — double strike until end of turn",
			}.Apply(ctx)
		})
	opus.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:       "f1812d69-c016-4bc3-b899-ed47f5e893e5",
		Name:           "Expressive Firedancer",
		ReadsManaSpent: game.ReadsYourCasts,
		Completeness:   CompletenessFull,
		Triggered:      []game.TriggeredAbility{opus},
	})
}
