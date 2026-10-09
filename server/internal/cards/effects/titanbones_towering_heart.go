package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Titanbones, Towering Heart — Legendary Creature — Skeleton Druid
// {3}{G}, 4/3:
//
//	"Reach
//	 Whenever you gain life, put two +1/+1 counters on Titanbones.
//	 When you discard this card, you gain 3 life."
//
// The discard trigger watches from the graveyard (CR 113.6, the cycling
// shape): the card is already there when the discard event fires, and
// any discard counts, a cost or an effect or the cleanup step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "510bccb1-61cd-49f1-a54a-80352653d1e3",
		Name:            "Titanbones, Towering Heart",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Titanbones, Towering Heart — put two +1/+1 counters on it", func(g *game.Game, item *game.StackItem) error {
				if !b09SourceStillOnBattlefield(g, item) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}.Apply(NewContext(g, item))
			}),
			InGraveyard(On(game.EventDiscardCard, Self,
				"Titanbones, Towering Heart — you gain 3 life", Do(GainLife{Amount: 3}))),
		},
	})
}
