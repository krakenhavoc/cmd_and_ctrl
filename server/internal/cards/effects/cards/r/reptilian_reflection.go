package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reptilian Reflection — Enchantment {2}{R} (Jumpstart 2022):
//
//	"Whenever you cycle a card, you may have this enchantment become a 5/4
//	 Dinosaur creature with trample and haste in addition to its other types
//	 until end of turn."
//
// WheneverYouCycle with the "you may" asked as the trigger resolves. The
// change is Lumbering Falls' layer 4/6/7b scoped effect for the turn: the
// Enchantment stays an Enchantment and gains Creature and Dinosaur, base 5/4,
// trample and haste. A Reflection that left, or left and came back, is a new
// object and changes nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73f6619c-9d92-434e-9f06-52e0ff15cb5e",
		Name:         "Reptilian Reflection",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.WheneverYouCycle("Reptilian Reflection — you may have it become a 5/4 Dinosaur creature with trample and haste until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						if !effects.SourceIsStillThisPermanent(g, item) {
							return nil
						}
						ctx := effects.NewContext(g, item)
						return effects.ScopedEffectFor{
							Target: item.SourceCardID,
							Mods: append([]game.Mod{
								game.AddTypesMod("Creature"),
								game.AddSubtypesMod(ctx.TextSubtypes("Dinosaur")...),
								game.AddKeywordsMod("trample", "haste"),
							}, game.SetBasePTMods(5, 4)...),
							Duration: effects.DurationUntilEndOfTurn(ctx),
							Label:    "Reptilian Reflection — a 5/4 Dinosaur creature with trample and haste until end of turn",
						}.Apply(ctx)
					}),
				"Reptilian Reflection — become a 5/4 Dinosaur creature with trample and haste until end of turn?"),
		},
	})
}
