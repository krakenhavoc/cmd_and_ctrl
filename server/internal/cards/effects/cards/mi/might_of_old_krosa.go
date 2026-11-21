package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Might of Old Krosa — Instant {G} (S82 GAP, #5948):
//
//	"Target creature gets +2/+2 until end of turn. If you cast this spell
//	 during your main phase, that creature gets +4/+4 until end of turn
//	 instead."
//
// A spell cast in a main phase resolves in that same phase (no step ends while
// the stack is not empty, CR 500.2), so the phase is read as it resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b50a6db0-16ce-4e98-afbd-ec37e1653e7c",
		Name:         "Might of Old Krosa",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			n := 2
			if effects.B14InYourMainPhase(ctx.Game, item.Controller) {
				n = 4
			}
			return effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: n, Toughness: n,
				Label: "Might of Old Krosa — gets bigger until end of turn"}.Apply(ctx)
		},
	})
}
