package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Territorial Aetherkite — Creature — Cat Dragon {4}{R}{R}, 6/5:
//
//	"Flying, haste
//	 When this creature enters, you get {E}{E} (two energy counters).
//	 Then you may pay one or more {E}. When you do, this creature deals
//	 that much damage to each other creature."
//
// ADR 0129 §3 (#1995, owner decision 3): "one or more" is the pay_amount
// prompt with a floor of 1, from the energy just gotten and any held
// before. Paying is "doing" (CR 603.12): the reflexive trigger carries
// the amount paid and deals it to each other creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4f63d306-cc14-405c-8424-d8990c670c74",
		Name:            "Territorial Aetherkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Territorial Aetherkite — you get {E}{E}, then may pay one or more {E}",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (GetEnergy{N: 2}).Apply(ctx); err != nil {
						return err
					}
					return PayEnergyAmount{
						Min:      1,
						Question: "Territorial Aetherkite — pay one or more {E}; it deals that much damage to each other creature",
						Unit:     game.PayAmountDamage,
						Then: func(ctx *Context, paid int) error {
							if paid <= 0 {
								return nil
							}
							t := WhenYouDo("Territorial Aetherkite — that much damage to each other creature", territorialAetherkiteDamageBody)
							t.Params = game.EffectParams{Amount: paid}
							return t.Apply(ctx)
						},
					}.Apply(ctx)
				}),
		},
	})
}
