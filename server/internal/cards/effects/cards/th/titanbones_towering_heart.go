package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "510bccb1-61cd-49f1-a54a-80352653d1e3",
		Name:            "Titanbones, Towering Heart",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Titanbones, Towering Heart — put two +1/+1 counters on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
			effects.InGraveyard(effects.On(game.EventDiscardCard, effects.Self,
				"Titanbones, Towering Heart — you gain 3 life", effects.Do(effects.GainLife{Amount: 3}))),
		},
	})
}
