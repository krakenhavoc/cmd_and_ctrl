package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Outrageous Robbery — Instant {X}{B}{B} (Edea steal-and-sac deck,
// #1565):
//
//	"Target opponent exiles the top X cards of their library face down.
//	 You may look at and play those cards for as long as they remain
//	 exiled. If you cast a spell this way, you may spend mana as though
//	 it were mana of any type to cast it."
//
// Gonti, Night Minister's exile-and-play grant with a count of X and a
// target: the top X cards of the target opponent's library, playable
// by the caster (lands included — "play") for as long as each stays in
// exile. X is the announced X (CR 601.2b). An X of zero exiles nothing.
// If the target opponent has left the game by resolution, the spell has
// no legal target and does nothing.
//
// The cards are exiled FACE DOWN (game.FaceDownPermitted, #1573): only
// the caster may look at them. Mana of any TYPE can be spent to cast
// them (CastPermission.AnyType), {C} included.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b194438-6946-45dd-8d77-c9de8c115d09",
		Name:         "Outrageous Robbery",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				return effects.GontiExileTopForPlay(ctx, t.ID, item.Controller, ctx.X())
			}
			return nil
		},
	})
}
