package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glimmerpost — Land — Locus (Scars of Mirrodin, #5577):
//
//	"When this land enters, you gain 1 life for each Locus on the
//	 battlefield.
//	 {T}: Add {C}."
//
// Cloudpost's Locus count, read as the trigger resolves and over every
// player's permanents, Glimmerpost itself included while it is still
// there. "Locus" is a word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92c9aad6-35ec-425d-be7d-393328992820",
		Name:         "Glimmerpost",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Glimmerpost — gain 1 life for each Locus on the battlefield",
				Do(Instruction(func(ctx *Context) error {
					locus := ctx.TextSubtype("Locus")
					n := 0
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.HasSubtype(locus) {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
				}))),
		},
	})
}
