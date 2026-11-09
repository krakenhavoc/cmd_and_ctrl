package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tellah, Great Sage — Legendary Creature — Human Wizard {3}{U}{R}, 3/3
// (Final Fantasy, #4462):
//
//	"Whenever you cast a noncreature spell, create a 1/1 colorless Hero
//	 creature token. If four or more mana was spent to cast that spell,
//	 draw two cards. If eight or more mana was spent to cast that spell,
//	 sacrifice Tellah and it deals that much damage to each opponent."
//
// Expressive Firedancer's cast trigger: the mana spent is counted as the
// spell is cast and carried on the item, so the ifs hold when the spell is
// countered or gone before the ability resolves. The Hero is made
// regardless. The sacrifice is for the object that triggered (CR 400.7), but
// "it deals that much damage" is the sentence's own and still happens when
// Tellah has already left, from its last-known information (CR 608.2h).
//
// No simplification.
func init() {
	const label = "Tellah, Great Sage — create a 1/1 Hero; with four or more mana spent, draw two; with eight or more, sacrifice Tellah and it deals that much damage to each opponent"
	trig := On(game.EventCast, YouCast(Noncreature()), label,
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if err := (CreateToken{Template: TokenCard(jobSelectHeroToken), N: 1}).Apply(ctx); err != nil {
				return err
			}
			spent := item.Params.Amount
			if spent >= 4 {
				if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			if spent >= 8 {
				if sourceIsStillThisPermanent(g, item) {
					if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
						return err
					}
				}
				return DamageEachOpponent{N: spent}.Apply(ctx)
			}
			return nil
		})
	trig.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:       "1a080b26-c766-4f7c-9198-4974d2133f75",
		Name:           "Tellah, Great Sage",
		ReadsManaSpent: game.ReadsYourCasts,
		Completeness:   CompletenessFull,
		Triggered:      []game.TriggeredAbility{trig},
	})
}
