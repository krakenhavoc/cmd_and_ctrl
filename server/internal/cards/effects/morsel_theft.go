package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Morsel Theft — Kindred Sorcery — Rogue {2}{B}{B} (Morningtide, #5592):
//
//	"Prowl {1}{B} (You may cast this for its prowl cost if you dealt
//	 combat damage to a player this turn with a Rogue.)
//	 Target player loses 3 life and you gain 3 life. If this spell's prowl
//	 cost was paid, draw a card."
//
// Soul Feast's drain at 3 (targetPlayerLosesLifeYouGain), then a card
// when the spell was cast for its prowl cost (ctx.PaidAltCost, the cost
// announced with the cast, CR 601.2b). Prowl is the gated alternative cost
// of prowl_keyword.go (CR 702.76a), reading the Kindred card's Rogue type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b58878ce-dd5a-4c43-882c-639de4f4a111",
		Name:             "Morsel Theft",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Prowl("{1}{B}")},
		Targets:          TargetPlayer("target player"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return targetPlayerLosesLifeYouGain(3)(ctx.Item, ctx) }),
			Instruction(morDrawIfProwled),
		),
	})
}
