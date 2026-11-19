package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dauntless Unity — Instant {1}{W} (Foundations Jumpstart, #5929):
//
//	"Kicker {1}{W} (You may pay an additional {1}{W} as you cast this
//	 spell.)
//	 Creatures you control get +1/+1 until end of turn. If this spell was
//	 kicked, those creatures get +2/+1 until end of turn instead."
//
// Inspired Charge's team pump: the set of creatures is fixed as the spell
// resolves (CR 611.2c). Kicked, the boost is +2/+1 instead of +1/+1, not
// in addition.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "dc665ddc-4326-4860-94ba-c40ced8b0af6",
		Name:          "Dauntless Unity",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{1}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			power := 1
			if ctx.WasKicked() {
				power = 2
			}
			return BoostUntilEOT{
				Match:     And(Creature(), YouControl()),
				Power:     power,
				Toughness: 1,
				Label:     "Dauntless Unity — creatures you control get a boost",
			}.Apply(ctx)
		},
	})
}
