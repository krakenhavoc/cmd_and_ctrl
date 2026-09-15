package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nature's Will — Enchantment {2}{G}{G} (EDHREC rank 2550):
//
//	"Whenever one or more creatures you control deal combat damage to
//	 a player, tap all lands that player controls and untap all lands
//	 you control."
//
// The green Bear Umbra. combatDamageToPlayerBy is the condition;
// "one or more" is a dedup, because the engine emits one damage
// event per creature — but per PLAYER, not per source: the stack
// label names the damaged player, and a second event for the same
// player is declined while that label's trigger is pending or on the
// stack (the engine's TriggerInFlightForEffect, keyed by that label). Creatures connecting
// with two players in one combat fire twice, once per player, as
// printed; three creatures hitting one player fire once. First-
// strike and regular damage are two batches and two triggers, as in
// paper. On resolution every untapped land the damaged player
// controls is tapped and every tapped land the controller controls
// is untapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9837287b-d821-4655-8f01-93fc6c7f0ecc",
		Name:         "Nature's Will",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g) &&
					!g.TriggerInFlightForEffect(source.InstanceID, b24NaturesWillLabel(g, ev.Target))
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				victim := ev.Target
				return game.NewTriggeredItem(source, b24NaturesWillLabel(g, victim),
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if err := b24TapAllLandsControlledBy(ctx, victim); err != nil {
							return err
						}
						return b16UntapAllYouControlMatching(ctx, item.Controller, func(c game.Card) bool { return c.IsLand() })
					})
			},
		}},
	})
}
