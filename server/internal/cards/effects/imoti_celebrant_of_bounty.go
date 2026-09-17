package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imoti, Celebrant of Bounty — Legendary Creature — Snake Druid
// {3}{G}{U}, 3/3:
//
//	"Cascade
//	 Spells you cast with mana value 6 or greater have cascade."
//
// Both shapes of the keyword on one card: Imoti HAS cascade (a
// FromStack trigger on the spell) and GIVES it to big spells (an
// ordinary battlefield trigger). Declaring them side by side is the
// clearest illustration of why they are different mechanisms.
//
// "Mana value 6 or greater" is the spell's mana value on the stack,
// not what it cost. CR 202.3c: a cost modifier changes the price and
// never the mana value, so a Goblin Electromancer cannot drop a
// seven-drop out of Imoti's range and a Sphere of Resistance cannot
// push a five-drop into it. The one thing that does count is X
// (CR 202.3e): a {X}{G} spell cast with X=5 is mana value 6 and gets
// cascade, which then exiles until a card with mana value less than 6.
func init() {
	Register(Spec{
		OracleID: "eca32dcd-6845-433e-a631-ed1f0ee78f25",
		Name:     "Imoti, Celebrant of Bounty",
		Triggered: []game.TriggeredAbility{
			Cascade(),
			GrantsCascade("Imoti, Celebrant of Bounty", func(spell game.Card, _ *game.Card, g *game.Game) bool {
				mv, ok := g.ManaValueForEffect(spell)
				return ok && mv >= 6
			}),
		},
	})
}
