package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harald, King of Skemfar — Legendary Creature — Elf Warrior {1}{B}{G}, 3/2:
//
//	"Menace
//	 When Harald enters, look at the top five cards of your library. You may
//	 reveal an Elf, Warrior, or Tyvar card from among them and put it into
//	 your hand. Put the rest on the bottom of your library in a random order."
//
// Icon of Ancestry's look-and-take (optional, one card, the rest to the
// bottom in a random order) with a three-way match: a card that has any one
// of the subtypes Elf, Warrior or Tyvar qualifies.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0089b07c-7f47-4a9a-9fb7-cc38ec9d7bcd",
		Name:            "Harald, King of Skemfar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Harald, King of Skemfar — look at the top five cards, you may take an Elf, Warrior or Tyvar card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := item.Controller
					elf, warrior := ctx.TextSubtype("Elf"), ctx.TextSubtype("Warrior")
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 5),
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.HasSubtype(elf) || c.HasSubtype(warrior) || c.HasSubtype("Tyvar")
						},
						Max:      1,
						Optional: true,
						Label:    "Harald, King of Skemfar — reveal an Elf, Warrior, or Tyvar card and put it into your hand",
						Key:      effects.TakeRevealRestBottomRandom,
					}.Apply(ctx)
				}),
		},
	})
}
