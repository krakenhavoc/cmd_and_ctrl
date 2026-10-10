package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Midnight Clock — Artifact {2}{U}:
//
//	"{T}: Add {U}.
//	 {2}{U}: Put an hour counter on this artifact.
//	 At the beginning of each upkeep, put an hour counter on this
//	 artifact.
//	 When the twelfth hour counter is put on this artifact, shuffle
//	 your hand and graveyard into your library, then draw seven
//	 cards. Exile this artifact."
//
// The counter kind is a free-form string (game/counter_types.go —
// unknown names round-trip without validation); "hour" needs no
// constant of its own for one card. Both counter-placing abilities
// share one closure since AddCounter needs the source's own instance
// ID, which is only known at resolution time (item.SourceCardID).
//
// The payoff is a trigger on the placement that CROSSES twelve
// (b12CounterTotalBefore reads the total before the event, so a
// thirteenth counter, or a placement that starts above twelve, fires
// nothing). It resolves as Echo of Eons does for one player: the hand
// and graveyard leave together as one tuck batch, the library is
// shuffled, seven cards are drawn, and the artifact is exiled if it is
// still the same permanent on the battlefield. The trigger does not
// need its source to resolve, so the shuffle and draw happen even if
// the Clock was removed in response.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c68faebc-b2cd-461b-b93e-e1fcd4816810",
		Name:         "Midnight Clock",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Put an hour counter on this artifact",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{2}{U}"),
			Effect:  midnightClockPutHourCounter,
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(_ game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return true
			}, "Midnight Clock — put an hour counter", midnightClockPutHourCounter),
			On(game.EventCounterPlaced, midnightClockTwelfthHour,
				"Midnight Clock — shuffle your hand and graveyard into your library, draw seven, exile this artifact",
				midnightClockPayoff),
		},
	})
}

// midnightClockPutHourCounter is shared by the activated ability and
// the upkeep trigger — both are "put an hour counter on this
// artifact" and nothing else. Package-level so neither stack item
// captures a *Card or *Game.
func midnightClockPutHourCounter(g *game.Game, item *game.StackItem) error {
	return AddCounter{Target: item.SourceCardID, Kind: "hour", N: 1}.Apply(NewContext(g, item))
}

// midnightClockTwelfthHour is "when the twelfth hour counter is put on
// this artifact": an hour-counter placement on the source whose new
// total reaches twelve from below.
func midnightClockTwelfthHour(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCounterPlaced || ev.Target != source.InstanceID || ev.Label != "hour" || ev.Amount < 12 {
		return false
	}
	before, _ := b12CounterTotalBefore(ev, g)
	return before < 12
}

// midnightClockPayoff shuffles the controller's hand and graveyard into
// their library, draws seven, then exiles the Clock. Only values cross
// the continuation, never the item or the game it started with.
func midnightClockPayoff(g *game.Game, item *game.StackItem) error {
	return midnightClockShuffleAway(g, item.Controller, item.SourceCardID)
}

func midnightClockShuffleAway(g *game.Game, player, clock uuid.UUID) error {
	ids := append(allHandCardIDs(g, player), echoOfEonsGraveyardCardIDs(g, player)...)
	return g.TuckCardsToLibraryThenForEffect(ids, game.TuckOptions{}, func(g *game.Game, _ []uuid.UUID) error {
		if err := g.ShuffleLibraryForEffect(player); err != nil {
			return err
		}
		if err := g.DrawNForEffect(player, 7); err != nil {
			return err
		}
		if g.Battlefield.Contains(clock) {
			return g.ExileCardForEffect(clock)
		}
		return nil
	})
}
