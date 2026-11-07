package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horrid Vigor — Instant {1}{G} (Duskmourn: House of Horror, #4466):
//
//	"Target creature gains deathtouch and indestructible until end of
//	 turn."
//
// Give In to Violence's body with no pump: both keywords land on the
// still-legal target (CR 608.2b) until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bef3f0a9-4d5c-4469-81df-adbeedeed768",
		Name:         "Horrid Vigor",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch", "indestructible")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(0, 0, "Horrid Vigor — deathtouch and indestructible", "deathtouch", "indestructible")(ctx.Game, item)
		},
	})
}
