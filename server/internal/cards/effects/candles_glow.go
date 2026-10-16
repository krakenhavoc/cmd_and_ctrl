package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Candles' Glow — Instant — Arcane {1}{W}:
//
//	"Prevent the next 3 damage that would be dealt to any target this
//	turn. You gain life equal to the damage prevented this way.
//	Splice onto Arcane {1}{W}"
//
// Mending Hands' shield with ADR 0108 §8's follow-up (CR 615.5): the life
// is gained as the damage is prevented, and none under "damage can't be
// prevented" (CR 615.12). Splice is ADR 0202 (#1989).
func init() {
	Register(Spec{
		OracleID:     "7509b934-6fbc-42fd-9040-cdf16bbfff42",
		Name:         "Candles' Glow",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{1}{W}"),
		Targets:      TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok {
				return nil
			}
			return PreventNextDamage{
				Target: t.ID,
				Amount: 3,
				Then:   preventedGainLifeBody,
				Label:  "Candles' Glow — prevent the next 3 damage and gain that much life",
			}.Apply(ctx)
		},
	})
}
