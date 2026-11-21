package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Action News Crew — {1}{W} Creature — Human Citizen 2/2 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"Vigilance
//	 Channel — {6}, Discard this card: Put a +1/+1 counter on each
//	 creature you control. Draw a card."
//
// Channel is an ability word (CR 207.2c): the ability is Boseiju, Who
// Endures's hand-zone activation with DiscardSelf as part of its cost. The
// counters are Abandoned Air Temple's one placement per creature you control
// as it resolves, then the draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a9d68cf5-3055-4ec9-9893-244bfd7046be",
		Name:            "Action News Crew",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {6}, Discard this card: Put a +1/+1 counter on each creature you control. Draw a card.",
			Cost:    game.AbilityCost{Mana: "{6}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.B11PutCounterOnEachCreatureYouControl(g, item); err != nil {
					return err
				}
				return effects.DrawCards{N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
