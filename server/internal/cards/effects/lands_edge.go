package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Land's Edge — World Enchantment {1}{R}{R}:
//
//	"Discard a card: If the discarded card was a land card, this
//	 enchantment deals 2 damage to target player or planeswalker. Any
//	 player may activate this ability."
//
// ADR 0109 §8 (#1862). The discard is the ACTIVATOR's (CR 602.1a): any
// player may activate the row (ADR 0106 §1), and the card comes from
// their own hand. The effect reads the card back off the payment
// record (Context.DiscardedCard, CR 400.7j), because by resolution
// nothing on the board says which card paid. The target is chosen as
// the ability is activated, whatever card is discarded, and a nonland
// discard resolves doing nothing — that is the printed card.
//
// The world rule (CR 704.5k) is the engine's: a second world
// enchantment entering puts the older one into its owner's graveyard.
//
// No purpose for the bot: whether the ability helps depends on holding
// a land card, which ActivationPurpose does not describe.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82d58f2d-66a4-4154-9826-01ca8e8d32d0",
		Name:         "Land's Edge",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "Discard a card: If the discarded card was a land card, this enchantment deals 2 damage to target player or planeswalker. Any player may activate this ability.",
			Cost:      DiscardACard(),
			AnyPlayer: true,
			Targets:   targetPlayerOrPlaneswalker(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				card, ok := ctx.DiscardedCard()
				if !ok || !card.IsLand() {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
