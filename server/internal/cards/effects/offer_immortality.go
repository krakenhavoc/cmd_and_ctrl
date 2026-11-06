package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Offer Immortality — Instant {1}{B} (Foundations, #4465):
//
//	"Target creature gains deathtouch and indestructible until end of
//	 turn."
//
// The still-legal target gains both keywords until end of turn
// (jmpB7FirstLegalTargetGains, Duelist's Heritage's body).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be4499e6-616a-4c11-b37e-9201d2492682",
		Name:         "Offer Immortality",
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch", "indestructible")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return jmpB7FirstLegalTargetGains("Offer Immortality — deathtouch and indestructible",
				"deathtouch", "indestructible")(ctx.Game, item)
		},
	})
}
