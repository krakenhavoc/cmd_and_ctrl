package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Olivia's Midnight Ambush — {1}{B} Instant (#2561, ADR 0132):
//
//	"Target creature gets -2/-2 until end of turn. If it's night, that
//	 creature gets -13/-13 until end of turn instead."
//
// "Instead" is a single read of the designation as the spell resolves:
// one shrink, of whichever size, never both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3b8a08b-5409-4d5f-9bab-8a7a6376a5b9",
		Name:         "Olivia's Midnight Ambush",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			shrink := 2
			if ItsNight(ctx.Game) {
				shrink = 13
			}
			for _, t := range ctx.LegalTargets() {
				if err := (BoostUntilEOT{Target: t.ID, Power: -shrink, Toughness: -shrink, Label: "Olivia's Midnight Ambush — gets smaller"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
