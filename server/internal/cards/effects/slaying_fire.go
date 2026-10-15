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
// what adamant wants. What adamant wants is the other declaration,
// WantsColors (#3853): concentrate the payment in one colour. An
// auto-paid cast then taps Mountains for the generic half before a
// Sol Ring, so three Mountains make it four damage. The card still
// reads what was actually spent, so two Mountains and a Sol Ring paid
// by hand make it three: that is the printed behaviour.
//
// A waived payment (strict mana off with too little in the pool, or Cast
// anyway) no longer leaves this card guessing: it declares ReadsManaSpent,
// so the caster is asked what they paid outside the pool, or whether they
// cast it without paying its mana cost, and the record it reads is exact
// (#3855, ADR 0068 amendment 2026-10-15).
func init() {
	Register(Spec{
		OracleID:       "1e94e647-9150-4b22-aa9f-d195f64fb20a",
		Name:           "Slaying Fire",
		ReadsManaSpent: game.ReadsOwnPayment,
		Completeness:   CompletenessFull,
		WantsColors:    []string{"R"},
		Targets:        TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			damage := 3
			if AdamantSpent(ctx, "R", 3) {
				damage = 4
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
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
			})
		},
	})
}
