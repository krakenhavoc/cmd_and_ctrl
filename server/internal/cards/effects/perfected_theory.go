package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Perfected Theory — Instant {U}:
//
//	"Choose one —
//	 • Target creature has base power and toughness 1/1 until end of
//	   turn.
//	 • Target creature has base power and toughness 4/5 until end of
//	   turn."
//
// Both modes are a layer 7b base-P/T set; counters and +N/+N effects
// still apply on top of it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d824a319-fa74-42e4-9cbe-dba1576d6bad",
		Name:         "Perfected Theory",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Target creature has base power and toughness 1/1 until end of turn.",
				TargetCreature("target creature")),
			Mode("Target creature has base power and toughness 4/5 until end of turn.",
				TargetCreature("target creature")),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			sizes := [][2]int{{1, 1}, {4, 5}}
			for i, s := range sizes {
				if !ctx.HasMode(i) {
					continue
				}
				t, ok := ModeTarget(ctx, 0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return untilEndOfTurn(ctx, t.ID, nil,
					"Perfected Theory — base power and toughness set until end of turn",
					game.SetBasePTMods(s[0], s[1])...)
			}
			return nil
		},
	})
}
