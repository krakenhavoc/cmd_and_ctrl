package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Omen of Fire — Instant {3}{R}{R} (Alliances, S82 OLD batch 817):
//
//	"Return all Islands to their owners' hands. Each player sacrifices a
//	 Plains or a white permanent of their choice for each white permanent
//	 they control."
//
// The Islands go first, so a white Island that was bounced is not counted;
// the sacrifice then counts and asks once the bounce has finished.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5935e9b3-b9b9-413c-ba17-103f5792131a",
		Name:         "Omen of Fire",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{Match: effects.HasSubtype("Island"), ThenTail: ctx.ItemTail(omenOfFireSacrifice)}.Apply(ctx)
		},
	})
}

var omenOfFireSacrifice = effects.CardTail("omen-of-fire/sacrifice", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.OldB817EachPlayerSacrificesForEachWhitePermanent(ctx.Game, ctx.Item)
})
