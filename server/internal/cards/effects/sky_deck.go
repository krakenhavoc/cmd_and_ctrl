package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sky Deck — Plane — Chicago (Unknown Event, #4870):
//
//	"Whenever you planeswalk here, at the beginning of your upkeep, or
//	 whenever chaos ensues, put a glass counter on up to one target nonland
//	 permanent.
//	 When you planeswalk away from here, destroy all permanents with glass
//	 counters on them."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first ability is ONE triggered ability with three conditions (OnAny,
// as Antarctic Research Base's two are), so its label, target and effect are
// shared. The glass counter is an ordinary named counter (storage only). The
// second ability looks back at the plane that has just left (CR 901.11b) and
// destroys every permanent bearing one, whoever controls it, all at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8baa42cf-32bc-4d1f-b4bd-02a4fce376c4",
		Name:         "Sky Deck",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(OnAny([]game.EventKind{game.EventPlaneswalkTo, game.EventBeginUpkeep, game.EventRollDie},
				skyDeckTriggers,
				"Sky Deck — put a glass counter on up to one target nonland permanent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return AddCounter{Target: id, Kind: planesB13GlassCounter, N: 1}.Apply(ctx)
				}),
				TargetPermanent("up to one target nonland permanent", Nonland()).WithCount(0, 1)),
			WhenYouPlaneswalkAwayFromThis("Sky Deck — destroy all permanents with glass counters on them",
				func(g *game.Game, item *game.StackItem) error {
					return DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.Counters[planesB13GlassCounter] > 0
					}}.Apply(NewContext(g, item))
				}),
		},
	})
}

const planesB13GlassCounter = "glass"

// skyDeckTriggers is "whenever you planeswalk here, at the beginning of your
// upkeep, or whenever chaos ensues".
func skyDeckTriggers(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventRollDie {
		return chaosEnsued(ev, source, lki, g)
	}
	return youPlaneswalkedToThisOrYourUpkeep(ev, source, lki, g)
}
