package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adéwalé, Breaker of Chains — Legendary Creature — Human Assassin Pirate
// {1}{U}{B}, 4/1 (Assassin's Creed, #5948):
//
//	"When Adéwalé enters, reveal the top six cards of your library. Put an
//	 Assassin, Pirate, or Vehicle card from among them into your hand and
//	 the rest on the bottom of your library in a random order.
//	 Whenever a Vehicle you control deals combat damage to a player, you
//	 may return this card from your graveyard to your hand."
//
// The reveal is Bounty of Skemfar's (everyone sees the six); the pick is
// one card of the three kinds, required when there is one, and the rest go
// to the bottom in a random order. The second ability works only from the
// graveyard (CR 113.6k, Curious Cadaver's InGraveyard): a "may" trigger
// that does nothing if the card has left the graveyard by resolution, and
// "you" is the owner there. Whether the dealer was a Vehicle is read as it
// dealt the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e30d9e08-cf88-4e2d-95e1-7144d8fea5c7",
		Name:         "Adéwalé, Breaker of Chains",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Adéwalé, Breaker of Chains — reveal the top six cards of your library; put an Assassin, Pirate, or Vehicle card from among them into your hand and the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  g.RevealTopOfLibraryForEffect(player, ctx.Source(), 6, "Adéwalé, Breaker of Chains — reveal the top six cards"),
						Match:  effects.Or(effects.HasSubtype("Assassin"), effects.HasSubtype("Pirate"), effects.HasSubtype("Vehicle")),
						Max:    1,
						Label:  "Adéwalé, Breaker of Chains — put an Assassin, Pirate, or Vehicle card into your hand",
						Key:    effects.TakeRevealRestBottomRandom,
					}.Apply(ctx)
				}),
			effects.Optional(effects.InGraveyard(effects.On(game.EventDealDamage, effects.GapB37VehicleYouControlDealtCombatDamageToAPlayer,
				"Adéwalé, Breaker of Chains — return this card from your graveyard to your hand",
				effects.ReturnThisFromYourGraveyardToHand)),
				"Return Adéwalé, Breaker of Chains from your graveyard to your hand?"),
		},
	})
}
