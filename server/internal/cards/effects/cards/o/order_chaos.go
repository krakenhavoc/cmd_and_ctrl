package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Order // Chaos — split card (Apocalypse, #5614):
//
//	Order — Instant {3}{W}: "Exile target attacking creature."
//	Chaos — Instant {2}{R}: "Creatures can't block this turn."
//
// Each half is an ADR 0034 face (ADR 0103): Order under the bare oracle
// ID, Chaos under "#1". Only one half can be cast (CR 709.3; the card has
// no fuse).
//
// Order is Not on My Watch: a creature that has been removed from combat
// by resolution is an illegal target (CR 608.2b). Chaos is Falter's
// restriction over every creature: "can't block" changes no
// characteristic, so the set is a rule read live until cleanup
// (CR 611.2c), and a creature that enters after Chaos resolved can't
// block either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc540259-459d-4561-9268-c4f323ea62d2",
		Name:         "Order",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "bc540259-459d-4561-9268-c4f323ea62d2#1",
		Name:         "Chaos",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.RestrictUntilEOT{
				Scope:        game.ScopeCreatures,
				Restrictions: game.CantBlock,
				Label:        "Chaos — creatures can't block",
			}.Apply(ctx)
		},
	})
}
