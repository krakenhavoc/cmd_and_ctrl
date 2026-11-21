package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elminster's Simulacrum — Instant {4}{U}{U} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"For each opponent, you create a token that's a copy of up to one target
//	 creature that player controls."
//
// One "up to one" clause per opponent, each bound to its opponent
// (PerOpponentClauses, Disorienting Choice's shape). Every pick still legal
// as the spell resolves is copied, the token entering under the caster's
// control (CR 707.2: copiable values as the target last existed, read now,
// CR 608.2b); a pick that left, or changed hands, in response is skipped.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bb32408-81d3-4d0c-8674-1a1c5e3feb3b",
		Name:         "Elminster's Simulacrum",
		Completeness: effects.CompletenessFull,
		Targets:      effects.PerOpponentClauses(game.MaxPlayers-1, "creature", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: t.ID, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
