package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadbridge Chant — "When this enchantment enters, mill ten cards. At the
// beginning of your upkeep, choose a card at random in your graveyard. If
// it's a creature card, put it onto the battlefield. Otherwise, put it into
// your hand."
func init() {
	Register(Spec{
		OracleID:     "506667b1-7922-4959-a5b3-0f8abe8c3615",
		Name:         "Deadbridge Chant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Deadbridge Chant — mill ten cards", Do(MillCards{N: 10})),
			AtYourUpkeep("Deadbridge Chant — return a random graveyard card", func(g *game.Game, item *game.StackItem) error {
				ids := graveyardIDs(g, item.Controller, nil)
				pick := randomPick(NewContext(g, item), ids, 1)
				if len(pick) == 0 {
					return nil
				}
				return randomCardToHandOrBattlefield(g, pick[0])
			}),
		},
	})
}
