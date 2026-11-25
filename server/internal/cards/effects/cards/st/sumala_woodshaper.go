package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sumala Woodshaper — Creature — Elf Druid {2}{G}{W}, 2/1 (Guilds of Ravnica,
// #5525):
//
//	"When this creature enters, look at the top four cards of your library. You
//	 may reveal a creature or enchantment card from among them and put it into
//	 your hand. Put the rest on the bottom of your library in a random order."
//
// Harald, King of Skemfar's look-and-take at four cards with a creature or
// enchantment filter: taking one is optional, a revealed card is shown to the
// table, and whatever is not taken goes to the bottom in a random order. A
// double-faced card is judged by its front face as it sits in the library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a44062b-f1db-4122-8570-1241a16cb895",
		Name:         "Sumala Woodshaper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sumala Woodshaper — look at the top four cards, you may take a creature or enchantment card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := item.Controller
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 4),
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.IsCreature() || c.HasCardType("enchantment")
						},
						Max:      1,
						Optional: true,
						Label:    "Sumala Woodshaper — reveal a creature or enchantment card and put it into your hand",
						Key:      effects.TakeRevealRestBottomRandom,
					}.Apply(ctx)
				}),
		},
	})
}
