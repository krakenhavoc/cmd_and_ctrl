package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Polluted Bonds — Enchantment {3}{B}{B} (EDHREC rank 3649):
//
//	"Whenever a land an opponent controls enters, that player loses 2
//	 life and you gain 2 life."
//
// The landfall drain, pointed at the table. Fires once per land
// entering under any opponent's control, however it entered —
// played, fetched, reanimated, a Dryad Arbor cast — with "that
// player" captured when the trigger fires. Life loss, not damage,
// so no prevention shield sees it; the 2 gained is 2 whether or not
// the opponent is still seated when it resolves. Post-layer types,
// so an animated land still counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13b33a41-c22b-4e7a-8323-abadf7f80961",
		Name:         "Polluted Bonds",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b34LandAnOpponentControlsEntered(ev, source, g)
				return ok
			},
			Key: "Polluted Bonds — that player loses 2 life, you gain 2 life",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				land, _ := b34LandAnOpponentControlsEntered(ev, source, g)
				item := game.NewTriggeredItem(source, "Polluted Bonds — that player loses 2 life, you gain 2 life")
				item.Params.Player = land.Controller
				return item
			},
			Effect: b34ThatPlayerLosesTwoYouGainTwo,
		}},
	})
}
