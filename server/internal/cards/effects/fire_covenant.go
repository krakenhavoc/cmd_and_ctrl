package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire Covenant — Instant {1}{B}{R} (EDHREC rank 2708):
//
//	"As an additional cost to cast this spell, pay X life.
//	 Fire Covenant deals X damage divided as you choose among any
//	 number of target creatures."
//
// The Rakdos one-sided wipe: pay life, kill a board at instant speed.
// The life is Toxic Deluge's additional cost (PayXLifeCost), paid with
// the spell on the stack and paid even if it is countered, and the X
// it announces is the amount divided — the same number by definition.
//
// The division is the caster's (#1563, CR 601.2d): announced with the
// targets, every target at least 1 and the shares summing to X, so X
// also caps how many creatures can be chosen. A creature that leaves
// in response takes nothing and its share is lost (CR 608.2b); the
// life is not refunded.
func init() {
	Register(Spec{
		OracleID:       "025939a0-424a-41bf-8fc9-2ef9ea5485f7",
		Name:           "Fire Covenant",
		Completeness:   CompletenessFull,
		XMatters:       true,
		AdditionalCost: PayXLifeCost(),
		Targets:        TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(DivideX()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DealDividedDamage(ctx)
		},
	})
}
