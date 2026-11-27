package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resounding Scream — Sorcery {2}{B} (Shards of Alara):
//
//	"Target player discards a card at random.
//	 Cycling {5}{U}{B}{R} ({5}{U}{B}{R}, Discard this card: Draw a card.)
//	 When you cycle this card, target player discards two cards at random."
//
// The spell is Black Cat's random discard (CR 701.9b) on a target player. The
// cycle trigger is Choking Tethers' graveyard trigger (CR 702.29c), mandatory,
// with its own target chosen as it goes on the stack; a player who is no longer
// legal as it resolves takes nothing (CR 608.2b). An empty hand discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "287d5a5d-e477-4d5b-815a-d83c73f4792d",
		Name:         "Resounding Scream",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return oldB508DiscardAtRandomFromTarget(ctx, 1)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{5}{U}{B}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.OldB508CycleTrigger("Resounding Scream — cycled: target player discards two cards at random",
				effects.TargetPlayer("target player"),
				func(g *game.Game, item *game.StackItem) error {
					return oldB508DiscardAtRandomFromTarget(effects.NewContext(g, item), 2)
				}),
		},
	})
}

// oldB508DiscardAtRandomFromTarget makes the first legal target player discard
// n cards at random.
func oldB508DiscardAtRandomFromTarget(ctx *effects.Context, n int) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return effects.DiscardCards{Player: t.ID, N: n}.Apply(ctx)
		}
	}
	return nil
}
