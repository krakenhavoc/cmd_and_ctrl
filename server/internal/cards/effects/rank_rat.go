package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rank Rat — Creature — Zombie Rat {1}{B}, 1/1:
//
//	"When this creature enters, each opponent discards a card."
//
// Each opponent chooses their own discard (CR 701.9a), so this opens one
// discard prompt per opponent rather than discarding at random. An
// opponent with an empty hand is skipped by the prompt itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "64e7ef06-ba30-4b19-9d44-ca77944cc930",
		Name:         "Rank Rat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rank Rat — each opponent discards a card", func(g *game.Game, item *game.StackItem) error {
				for _, opp := range NewContext(g, item).Opponents() {
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: opp, Source: item.SourceCardID, N: 1})
				}
				return nil
			}),
		},
	})
}
