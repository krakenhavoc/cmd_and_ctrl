package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hellkite Igniter — Creature — Dragon {5}{R}{R}, 5/5 (Magic Commander
// Collection, #5908):
//
//	"Flying, haste
//	 {1}{R}: This creature gets +X/+0 until end of turn, where X is the
//	 number of artifacts you control."
//
// X is counted as the ability resolves, after layers (Aerid Konstrari's
// read), and the boost applies only while the Igniter is still the same
// permanent on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "12aba1d4-477f-4d36-96bb-c3c7bbd51b75",
		Name:            "Hellkite Igniter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +X/+0 until end of turn, where X is the number of artifacts you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Effect: Do(Instruction(func(ctx *Context) error {
				item := ctx.Item
				if !sourceIsStillThisPermanent(ctx.Game, item) {
					return nil
				}
				return BoostUntilEOT{
					Target: item.SourceCardID,
					Power:  rfCreatureAArtifactsYouControl(ctx.Game, item.Controller),
					Label:  "Hellkite Igniter — +X/+0",
				}.Apply(ctx)
			})),
		}},
	})
}
