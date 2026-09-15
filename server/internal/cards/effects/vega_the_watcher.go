package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vega, the Watcher — Legendary Creature — Bird Spirit {1}{W}{U},
// 2/2 (EDHREC rank 2728):
//
//	"Flying
//	 Whenever you cast a spell from anywhere other than your hand,
//	 draw a card."
//
// The flashback / foretell / impulse-draw commander. Flying rides
// PrintedKeywords. The trigger is a cast by the controller whose
// origin zone was not their hand (b25CastFromNotHand): the cast
// event carries the zone the spell left — the command zone for a
// commander, the graveyard for a flashback, exile for an impulse
// grant or a warped card's later cast. A copy of a spell is not
// cast and fires nothing, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "95e66851-7aff-4559-91b4-6b8a95c6e1f8",
		Name:            "Vega, the Watcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b25CastFromNotHand(ev, source.Controller)
			}, "Vega, the Watcher — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
