package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monumental Corruption — Sorcery {3}{B}{B} (Phyrexia: All Will Be One
// Commander):
//
//	"Target player draws X cards and loses X life, where X is the number
//	 of artifacts you control."
//
// X is the number of artifacts the CASTER controls, counted once as the spell
// resolves (CR 608.2h) and used for both halves; the targeted player may be
// the caster or anyone else. A player who is no longer a legal target neither
// draws nor loses life (CR 608.2b). With X = 0 nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b49924b-cf83-46d9-b665-9c718a40e16c",
		Name:         "Monumental Corruption",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			x := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsArtifact() })
			if x <= 0 {
				return nil
			}
			if err := (effects.DrawCards{Player: t.ID, N: x}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -x)
		},
	})
}
