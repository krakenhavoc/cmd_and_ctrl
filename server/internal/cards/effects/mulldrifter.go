package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mulldrifter — 2/2 Elemental for {4}{U}:
//
//   "Flying. When Mulldrifter enters the battlefield, draw two cards.
//   Evoke {2}{U} (You may cast this spell for its evoke cost. If you
//   do, it's sacrificed when it enters the battlefield.)"
//
// S19 sub-PR 3 ships the ETB-draw half via the Triggered slot —
// mandatory, no target, no opponent prompt. Evoke is an alt-cast
// path (CR 702.74) that lands with S29's alt-cast arc; Mulldrifter
// just costs {4}{U} from hand for now.
//
// Sub-PR 3 sandbox pattern: Build runs the effect inline (DrawN via
// effect_api) and returns nil so no StackItem queues. The
// EventDrawCard breadcrumb supplies the "trigger fired" visibility;
// CR-correct stack-resolution timing is deferred to a future
// engine-completeness sweep.
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
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				ctx := NewContext(g, nil)
				_ = DrawCards{Player: source.Controller, N: 2}.Apply(ctx)
				return nil
			},
		}},
	})
}
