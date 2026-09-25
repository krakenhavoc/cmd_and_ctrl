package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horn of Greed — Artifact {3} (EDHREC rank 2134):
//
//	"Whenever a player plays a land, that player draws a card."
//
// The symmetrical lands-matter draw engine. "Plays a land" is a
// land PLAY, not a land entering: a fetched or ramped land draws
// nothing, a land played from exile through Prosper's grant draws.
// Since #1326 the engine stamps CR 305.4's distinction on the
// settled entry itself (Event.Played), so the trigger reads that
// directly (b20LandPlayed) rather than guessing from the zone the
// land came from. The drawer is the player who played the land — the
// event's Actor, read at resolution off the item's carried trigger
// context (item.Trigger.Event.Actor, #1223) — not the Horn's
// controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b8181d53-1954-4f46-8670-8696440208e8",
		Name:         "Horn of Greed",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				ok := b20LandPlayed(ev, g)
				return ok
			},
			Key: "Horn of Greed — that player draws a card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Trigger.Event.Actor, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
