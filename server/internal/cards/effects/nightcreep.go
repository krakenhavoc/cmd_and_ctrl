package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nightcreep — Instant {B}{B}:
//
//	"Until end of turn, all creatures become black and all lands become
//	 Swamps."
//
// Two parts of one effect, each over the set fixed as it begins (CR
// 611.2c): every creature on the battlefield is black until end of turn (a
// layer-5 colour set, CR 613.1e), and every land is a Swamp (ADR 0109 §1,
// #1881: CR 305.7's type set in layer 4). A land loses its other land
// types and the abilities its rules text gives it, keeps its other
// subtypes (CR 205.1a) and taps for {B} (CR 305.6). A creature or land
// that arrives later in the turn is untouched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b4b677e-3b18-454c-b385-b3ffb65f916e",
		Name:         "Nightcreep",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (ScopedEffectFor{
				Match:    Creature(),
				Mods:     []game.Mod{game.SetColorsMod("B")},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Nightcreep — all creatures are black",
			}).Apply(ctx); err != nil {
				return err
			}
			return LandBecomes{
				Match:    Land(),
				Types:    []string{"Swamp"},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Nightcreep — all lands are Swamps",
			}.Apply(ctx)
		},
	})
}
