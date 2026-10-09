package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shark Typhoon — Enchantment {5}{U}:
//
//	"Whenever you cast a noncreature spell, create an X/X blue Shark
//	 creature token with flying, where X is that spell's mana value.
//	 Cycling {X}{1}{U} ({X}{1}{U}, Discard this card: Draw a card.)
//	 When you cycle this card, create an X/X blue Shark creature token
//	 with flying."
//
// The cast trigger reads the spell's mana value as the trigger resolves
// (CR 202.3e: an {X} counts only while the spell is on the stack), and
// a spell that has already left the stack counts as zero.
//
// The cycling half is one ability here: X is announced with the cycling
// cost and the Shark is made as the cycling ability resolves, together
// with the card it draws. The printed text makes the Shark from a
// separate "when you cycle" trigger, which has no way to read the X the
// cycling ability paid.
func init() {
	cycling := Cycling("{1}{U}")
	cycling.Label = "Cycling {X}{1}{U} ({X}{1}{U}, Discard this card: Draw a card.)"
	cycling.Cost = Plus(ManaCost("{X}{1}{U}"), DiscardThis())
	cycling.Effect = func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return rfReprintBSharkToken(ctx, item.Controller, ctx.X())
	}
	Register(Spec{
		OracleID:     "8c0520fa-276b-4d21-b4a9-dce1fce59f6b",
		Name:         "Shark Typhoon",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The Shark from cycling is made as the cycling ability resolves, not by a separate trigger, so countering the cycling ability stops it too.",
		},
		Activated: []ActivatedAbility{cycling},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Shark Typhoon — create an X/X blue Shark with flying, X the spell's mana value",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return rfReprintBSharkToken(NewContext(g, item), item.Controller, spellManaValueForEffect(g, item.Trigger.Event.CardID))
				}),
		},
	})
}
