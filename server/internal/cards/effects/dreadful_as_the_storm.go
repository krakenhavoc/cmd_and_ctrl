package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreadful as the Storm — Instant {2}{U}:
//
//	"Target creature has base power and toughness 5/5 until end of
//	 turn. The Ring tempts you."
//
// The base power and toughness are a layer 7b effect (CR 613.4b), so
// counters and pumps still apply on top. Any creature can be the
// target, an opponent's included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd9f4f20-cdef-4a20-b4cb-4f25c137a786",
		Name:         "Dreadful as the Storm",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := untilEndOfTurn(ctx, t.ID, nil, "Dreadful as the Storm — base 5/5 until end of turn",
					game.SetBasePTMods(5, 5)...); err != nil {
					return err
				}
			}
			return TheRingTemptsYou{}.Apply(ctx)
		},
	})
}
