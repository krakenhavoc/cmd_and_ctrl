package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dissipation Field — Enchantment {2}{U}{U} (EDHREC rank 4316):
//
//	"Whenever a permanent deals damage to you, return it to its
//	 owner's hand."
//
// A four-mana enchantment that makes attacking you a losing
// proposition: the creature connects once, and then goes home. In a
// multiplayer game that is enough to redirect a whole table's combat
// somewhere else, which is the only thing a Dissipation Field is ever
// asked to do.
//
// "A PERMANENT", which is wider and narrower than it first reads:
//
//   - Any permanent type. A Pestilence, a planeswalker's minus, a
//     Vehicle that got through — if the source is on the battlefield,
//     the Field bounces it.
//   - Not a spell. A Lightning Bolt to the face is not a permanent
//     dealing damage, and the Field does nothing. Nor is a source that
//     has already left the battlefield.
//   - "To YOU", not to your creatures or your planeswalkers. The Field
//     protects its controller's life total and nothing else.
//
// It returns the permanent to its OWNER's hand, which is the zone
// router's default — a stolen creature goes back to whoever owns the
// card, not to whoever was attacking with it, and a token bounced this
// way ceases to exist (CR 111.8).
//
// The source is looked up live. Combat damage is dealt before
// state-based actions run, so an attacker that traded with its blocker
// is still on the battlefield when this fires and really does get
// bounced instead of dying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "35426427-4268-4c8f-9fe1-270f2ce43d97",
		Name:         "Dissipation Field",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b41PermanentDealtDamageToYou(ev, source, g)
				return ok
			},
			Key: "Dissipation Field — return it to its owner's hand",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BounceToHand{Target: item.Trigger.Event.Source}.Apply(NewContext(g, item))
			},
		}},
	})
}
