package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "92c9aad6-35ec-425d-be7d-393328992820",
		Name:         "Glimmerpost",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Glimmerpost — gain 1 life for each Locus on the battlefield",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
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
					return effects.GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
				}))),
		},
	})
}
