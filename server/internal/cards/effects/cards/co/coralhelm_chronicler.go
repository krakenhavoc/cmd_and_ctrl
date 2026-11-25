package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coralhelm Chronicler — Creature — Merfolk Wizard {2}{U}, 2/2 (Zendikar
// Rising, S82 OLD #5508):
//
//	"Whenever you cast a kicked spell, draw a card, then discard a card.
//	 When this creature enters, look at the top five cards of your library.
//	 You may reveal a card with a kicker ability from among them and put it
//	 into your hand. Put the rest on the bottom of your library in a random
//	 order."
//
// The cast trigger reads the spell's paid kicker as it goes on the stack
// (CR 702.33a, 601.2i). "A card with a kicker ability" is a card whose
// catalog row declares a kicker cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "621fc513-b16a-4c09-93f8-424e688471e2",
		Name:         "Coralhelm Chronicler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB089YouCastAKickedSpell,
				"Coralhelm Chronicler — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) }),
			effects.WhenThisEnters("Coralhelm Chronicler — look at the top five cards; you may reveal a card with a kicker ability and put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 5),
						Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return effects.S82OldB089CardHasKicker(c) },
						Max:      1,
						Optional: true,
						Label:    "Coralhelm Chronicler — you may reveal a card with a kicker ability from among them and put it into your hand",
						Key:      effects.TakeRevealRestBottomRandom,
					}.Apply(ctx)
				}),
		},
	})
}
