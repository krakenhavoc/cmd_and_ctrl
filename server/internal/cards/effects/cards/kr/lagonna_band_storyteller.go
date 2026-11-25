package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lagonna-Band Storyteller — Creature — Centaur Advisor {3}{W}, 3/4:
//
//	"When this creature enters, you may put target enchantment card from your
//	 graveyard on top of your library. If you do, you gain life equal to its
//	 mana value."
//
// The mana value is read as the card sat in the graveyard; the life is gained
// only when the card did reach the top of the library.
func init() {
	const label = "Lagonna-Band Storyteller — put target enchantment card from your graveyard on top of your library; gain life equal to its mana value"
	effects.Register(effects.Spec{
		OracleID:     "0d208c1e-c273-437f-9aea-136400e450e1",
		Name:         "Lagonna-Band Storyteller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						card, ok := g.LookupCardForEffect(id)
						if !ok {
							return nil
						}
						value := card.ManaValue()
						if err := g.TuckToLibraryForEffect(id, false); err != nil {
							return err
						}
						if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
							return nil
						}
						return effects.GainLife{Amount: value}.Apply(ctx)
					}),
					"Lagonna-Band Storyteller — put target enchantment card from your graveyard on top of your library?"),
				effects.TargetCardInGraveyard("target enchantment card from your graveyard", effects.Enchantment(), effects.YouOwn()),
			),
		},
	})
}
