package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wrath of the Skies — Sorcery {X}{W}{W}:
//
//	"You get X {E} (energy counters), then you may pay any amount of
//	 {E}. Destroy each artifact, creature, and enchantment with mana
//	 value less than or equal to the amount of {E} paid this way."
//
// ADR 0129 §3 (#1995, owner decision 3): X is the announced X; the
// payment is the pay_amount prompt, and the destruction always happens
// — paying nothing still destroys every artifact, creature and
// enchantment with mana value 0, tokens among them. The bot is offered
// the largest mana value an opponent's permanent of those types has,
// within its energy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "21efa6f2-3e7c-4576-b5a3-0b75b2843442",
		Name:         "Wrath of the Skies",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GetEnergy{N: ctx.X()}).Apply(ctx); err != nil {
				return err
			}
			return PayEnergyAmount{
				Question: "Wrath of the Skies — pay any amount of {E}; destroy each artifact, creature and enchantment with mana value up to it",
				Unit:     game.PayAmountOther,
				Goal:     largestOpposingManaValue,
				Then: func(ctx *Context, paid int) error {
					return DestroyAllMatching{Match: And(Or(Artifact(), Creature(), Enchantment()), ManaValueLE(paid))}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
