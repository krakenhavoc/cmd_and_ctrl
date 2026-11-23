package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mob Rule — Sorcery {4}{R}{R}:
//
//	"Choose one —
//	 • Gain control of all creatures with power 4 or greater until end of
//	   turn. Untap those creatures. They gain haste until end of turn.
//	 • Gain control of all creatures with power 3 or less until end of turn.
//	   Untap those creatures. They gain haste until end of turn."
//
// Two modes with no target. The creatures are the ones on the battlefield as
// the spell resolves, judged on their layered power, yours included: taking
// a creature you already control just untaps it and gives it haste.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acca2128-6ebb-48a5-a6a2-1bc5e4aa86ce",
		Name:         "Mob Rule",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Gain control of all creatures with power 4 or greater until end of turn. Untap those creatures. They gain haste until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.RprB28StealAllMatching(ctx, "Mob Rule", effects.RprB28CreatureWithPower(func(p int) bool { return p >= 4 }))
				}),
			effects.ModeDoing("Gain control of all creatures with power 3 or less until end of turn. Untap those creatures. They gain haste until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.RprB28StealAllMatching(ctx, "Mob Rule", effects.RprB28CreatureWithPower(func(p int) bool { return p <= 3 }))
				}),
		),
	})
}
