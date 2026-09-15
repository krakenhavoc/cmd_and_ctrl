package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horn of Greed — Artifact {3} (EDHREC rank 2134):
//
//	"Whenever a player plays a land, that player draws a card."
//
// The symmetrical lands-matter draw engine. "Plays a land" is a
// land PLAY, not a land entering: a fetched or ramped land draws
// nothing, a land played from exile through Prosper's grant draws.
// The engine emits no land-play event, so the trigger reads the
// origin of the battlefield entry and, for the exile / graveyard
// origins that a returning land shares with a played one, the
// engine's own land-drop tally (b20LandPlayed). The drawer is the
// player who played the land — the event's Actor — not the Horn's
// controller.
//
// Sandbox simplification, declared: when a land was RETURNED to the
// battlefield from exile or a graveyard by an effect earlier in the
// same turn, the tally can no longer tell a later land played from
// exile or a graveyard apart from another return, and the trigger
// stays quiet for it. A land played from hand is always seen.
// Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "b8181d53-1954-4f46-8670-8696440208e8",
		Name:         "Horn of Greed",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A land played from exile or from a graveyard isn't counted if another land already came back to the battlefield from exile or a graveyard earlier that turn."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				ok := b20LandPlayed(ev, g)
				return ok
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				drawer := ev.Actor
				return game.NewTriggeredItem(source, "Horn of Greed — that player draws a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: drawer, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
