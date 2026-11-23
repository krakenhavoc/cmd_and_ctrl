package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadbridge Chant — "When this enchantment enters, mill ten cards. At the
// beginning of your upkeep, choose a card at random in your graveyard. If
// it's a creature card, put it onto the battlefield. Otherwise, put it into
// your hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "506667b1-7922-4959-a5b3-0f8abe8c3615",
		Name:         "Deadbridge Chant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Deadbridge Chant — mill ten cards", effects.Do(effects.MillCards{N: 10})),
			effects.AtYourUpkeep("Deadbridge Chant — return a random graveyard card", func(g *game.Game, item *game.StackItem) error {
				ids := effects.GraveyardIDs(g, item.Controller, nil)
				pick := effects.RandomPick(effects.NewContext(g, item), ids, 1)
				if len(pick) == 0 {
					return nil
				}
				return effects.RandomCardToHandOrBattlefield(g, pick[0])
			}),
		},
	})
}
