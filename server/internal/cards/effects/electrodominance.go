package effects

// Electrodominance — Instant {X}{R}{R} (Tarkir: Dragonstorm Commander):
//
//	"Electrodominance deals X damage to any target. You may cast a spell
//	 with mana value X or less from your hand without paying its mana
//	 cost."
//
// Blaze's damage, then Kari Zev's Expertise's free cast with the ceiling
// at the announced X: a choose_cards pick over the hand (floor zero, "you
// may"), then the pass-closed free cast opened inside the resolution
// (offerFreeHandCast), ignoring timing (CR 608.2g), with an {X} spell cast
// at X = 0. The damage is dealt first and the target does not die until the
// resolution ends, so a spell cast now sees it still on the battlefield. A
// target that is illegal on resolution makes the whole spell do nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a59d86e-405d-4c70-b2a5-c64f5cecfd9e",
		Name:         "Electrodominance",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return cmdrB19DealXToTheTarget(ctx) }),
			Instruction(func(ctx *Context) error {
				offerFreeHandCast(ctx, ctx.Controller(), electrodominanceFreeCast,
					"a spell with mana value X or less", spellWithManaValueAtMost(ctx.X()))
				return nil
			}),
		),
	})
}

// electrodominanceFreeCast is the offer's keyed continuation (#4246).
var electrodominanceFreeCast = registerFreeHandCastOffer("choose-cards/electrodominance-cast", "Electrodominance")
