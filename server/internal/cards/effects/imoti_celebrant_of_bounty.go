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
// "Mana value 6 or greater" is the spell's PRINTED mana value, not
// what it cost. CR 202.3c: a cost modifier changes the price and
// never the mana value, so a Goblin Electromancer cannot drop a
// seven-drop out of Imoti's range and a Sphere of Resistance cannot
// push a five-drop into it.
func init() {
	Register(Spec{
		OracleID: "eca32dcd-6845-433e-a631-ed1f0ee78f25",
		Name:     "Imoti, Celebrant of Bounty",
		Triggered: []game.TriggeredAbility{
			Cascade(),
			GrantsCascade("Imoti, Celebrant of Bounty", func(spell game.Card, _ *game.Card, _ *game.Game) bool {
				return spell.ManaValue() >= 6
			}),
		},
	})
}
