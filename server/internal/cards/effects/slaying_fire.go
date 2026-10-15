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
// A payment the engine waived (an insufficient permissive pool or
// Cast anyway) counts zero red mana, so the spell deals its printed 3 —
// weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "1e94e647-9150-4b22-aa9f-d195f64fb20a",
		Name:         "Slaying Fire",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"When “Charge mana costs” is off, or you choose Cast anyway, only mana spent from your mana pool counts, so Slaying Fire deals 4 only when all three red mana came from your mana pool."},
		WantsColors:  []string{"R"},
		Targets:      TargetAny(),
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
