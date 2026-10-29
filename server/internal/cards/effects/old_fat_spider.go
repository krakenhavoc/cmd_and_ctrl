package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Old Fat Spider — Creature — Spider {4}{G}{G} (The Hobbit), 6/7:
//
//	"Reach
//	 This creature can't be blocked by creatures with power 2 or less.
//	 Whenever this creature becomes the target of a spell or ability an
//	 opponent controls, draw a card."
//
// #4454, S76. Legolas Greenleaf's block rule (CantBeBlockedBy on this
// creature) and Thunderbreak Regent's opponent-targeting trigger: the
// becomes-target event names the targeting player in Actor, so a spell or
// ability its own controller aims at it draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8790f842-e842-4bed-adf4-5b3cc5fd68a9",
		Name:            "Old Fat Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBecomesTarget},
			Key:     "Old Fat Spider — draw a card (an opponent targeted it)",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesTarget && ev.CardID == source.InstanceID &&
					ev.Actor != uuid.Nil && ev.Actor != source.Controller
			},
			Effect: Do(DrawCards{N: 1}),
		}},
	})
}
