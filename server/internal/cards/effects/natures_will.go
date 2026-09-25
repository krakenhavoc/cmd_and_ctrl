package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nature's Will — Enchantment {2}{G}{G} (EDHREC rank 2550):
//
//	"Whenever one or more creatures you control deal combat damage to
//	 a player, tap all lands that player controls and untap all lands
//	 you control."
//
// The green Bear Umbra. combatDamageToPlayerBy is the condition;
// "one or more" is the engine's once-per-batch guard with its PLAYER
// dimension (OncePerBatchPerPlayer, CR 603.2c / #784), so a second
// damage event naming the same player in the same damage step is
// declined and a creature connecting with a second player is its own
// trigger. Creatures connecting with two players in one combat fire
// twice, once per player, as printed; three creatures hitting one
// player fire once. First-strike and regular damage are two damage
// steps and two batches (CR 510.4), so two triggers, as in paper.
// The stack label still names the damaged player, which is what
// tells two simultaneous triggers apart on the stack. On resolution
// every untapped land the damaged player controls is tapped and
// every tapped land the controller controls is untapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9837287b-d821-4655-8f01-93fc6c7f0ecc",
		Name:         "Nature's Will",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			OncePerBatch: true,
			BatchKey:     PerPlayer,
			Key:          b24NaturesWillKey,
			Watches:      []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, b24NaturesWillLabel(g, ev.Target))
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				victim := item.Trigger.Event.Target
				if err := b24TapAllLandsControlledBy(ctx, victim); err != nil {
					return err
				}
				return b16UntapAllYouControlMatching(ctx, item.Controller, func(c game.Card) bool { return c.IsLand() })
			},
		}},
	})
}
