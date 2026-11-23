package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a3b8a08b-5409-4d5f-9bab-8a7a6376a5b9",
		Name:         "Olivia's Midnight Ambush",
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			shrink := 2
			if effects.ItsNight(ctx.Game) {
				shrink = 13
			}
			for _, t := range ctx.LegalTargets() {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: -shrink, Toughness: -shrink, Label: "Olivia's Midnight Ambush — gets smaller"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
