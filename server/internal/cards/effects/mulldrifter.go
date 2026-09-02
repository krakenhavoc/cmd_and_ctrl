package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mulldrifter — 2/2 Elemental for {4}{U}:
//
//	"Flying. When Mulldrifter enters the battlefield, draw two cards.
//	Evoke {2}{U} (You may cast this spell for its evoke cost. If you
//	do, it's sacrificed when it enters the battlefield.)"
//
// S19 sub-PR 3 ships the ETB-draw half via the Triggered slot —
// mandatory, no target, no opponent prompt. Evoke is an alt-cast
// path (CR 702.74) that lands with S29's alt-cast arc; Mulldrifter
// just costs {4}{U} from hand for now.
//
// Build returns a real stack item; the draw happens when the
// trigger resolves, so opponents get their response window.
func init() {
	Register(Spec{
		OracleID:        "24d0f5e7-0d9e-4b76-900e-a7274e80312d",
		Name:            "Mulldrifter",
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Mulldrifter — draw two cards",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
