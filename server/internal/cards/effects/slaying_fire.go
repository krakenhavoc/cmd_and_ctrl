package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slaying Fire — Instant {2}{R} (EDHREC rank 9127):
//
//	"Slaying Fire deals 3 damage to any target.
//	 Adamant — If at least three red mana was spent to cast this
//	 spell, it deals 4 damage instead."
//
// The first adamant card in the catalog (#761). Adamant is an ability
// word (CR 207.2c), so the counting lives here: AdamantSpent reads
// how many RED mana paid, off the record on the stack item.
//
// It deliberately does NOT set Spec.WantsDistinctColors. That switch
// spreads a payment across COLOURS, which is the exact opposite of
// what adamant wants, and concentrating a payment into one colour is
// a different strategy the solver does not have. So adamant reads
// what the player actually spent: three Mountains make it four
// damage, two Mountains and a Sol Ring make it three. That is the
// printed behaviour, and the reason the ADR left the concentrating
// strategy out rather than inventing one nobody asked for.
//
// A payment the engine waived (permissive mode, a strict-mode
// override) counts zero red mana, so the spell deals its printed 3 —
// weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "1e94e647-9150-4b22-aa9f-d195f64fb20a",
		Name:         "Slaying Fire",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so Slaying Fire always deals 3 rather than 4."},
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			damage := 3
			if AdamantSpent(ctx, "R", 3) {
				damage = 4
			}
			for _, ref := range ctx.LegalTargets() {
				if err := (DealDamage{
					Source: item.SourceCardID,
					Target: ref.ID,
					Amount: damage,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
