package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychic Drain — Sorcery {X}{U}{B} (Ravnica: City of Guilds, S82 OLD b582):
//
//	"Target player mills X cards and you gain X life."
//
// The mill is the targeted player's (a library with fewer cards mills what it
// has, CR 701.17b), the life is the caster's and is X whatever the mill did.
// With the one target illegal as it resolves the whole spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac5427d7-9af9-4f94-a8c3-da019ec0b5c5",
		Name:         "Psychic Drain",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			p, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			if err := (effects.MillCards{Player: p, N: ctx.X()}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: ctx.X()}.Apply(ctx)
		},
	})
}
