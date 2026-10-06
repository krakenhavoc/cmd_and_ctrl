package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winds of Qal Sisma — Instant {1}{G}:
//
//	"Prevent all combat damage that would be dealt this turn.
//	 Ferocious — If you control a creature with power 4 or greater,
//	 instead prevent all combat damage that would be dealt this turn by
//	 creatures your opponents control."
//
// Ferocious is checked once, as the spell resolves; the ruling: the
// narrower shield applies "even if you no longer control a creature with
// power 4 or greater as that damage would be dealt". The narrower shield
// is #2026's controller test, read as each creature would deal combat
// damage (CR 609.7b). Without ferocious it is Fog.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "e7871b4d-a408-4377-beee-6b1d3c7dd57d",
		Name:         "Winds of Qal Sisma",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if b24ControlsCreatureWithPowerAtLeast(ctx.Game, ctx.Controller(), 4) {
				return combatShieldAgainstCreatures(game.DamageSourceFilter{Controller: game.SourceControllerOpponents}).Apply(ctx)
			}
			return PreventAllCombatDamageThisTurn{Label: "Winds of Qal Sisma: prevent combat damage"}.Apply(ctx)
		},
	})
}
