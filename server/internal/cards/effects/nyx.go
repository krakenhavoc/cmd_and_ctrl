package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nyx — Plane — Theros (Commander Masters, #4870):
//
//	"Nontoken creatures are enchantments in addition to their other types.
//	 Constellation — Whenever an enchantment you control enters, you gain
//	 1 life.
//	 Whenever chaos ensues, choose a color. Add an amount of mana of that
//	 color equal to your devotion to that color."
//
// A plane (ADR 0216 §6). The static is a layer-4 type addition over every
// nontoken creature, whoever controls it, applying from the moment it enters
// (the ruling), so a creature entering under the planar controller is an
// enchantment as it enters and constellation triggers for it. The chaos
// ability (CR 311.7) is not a mana ability (the ruling): it asks the planar
// controller to choose a color as it resolves, and the devotion (CR 700.5,
// the mana symbols of that color among the mana costs of permanents they
// control) is counted once the color is chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6552e3dd-ce35-4267-a25e-f5d4dceb8311",
		Name:         "Nyx",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Nontoken creatures are enchantments in addition to their other types.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature() && !target.IsToken()
			},
			Apply: addEnchantmentType,
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsEnchantment()
			}, "Nyx — constellation: you gain 1 life", Do(GainLife{Amount: 1})),
			WheneverChaosEnsues("Nyx — choose a color; add an amount of mana of that color equal to your devotion to that color",
				func(g *game.Game, item *game.StackItem) error {
					ChooseColorTail(game.ColorForMana, g, item.Controller, item.SourceCardID, "Nyx — choose a color",
						NewContext(g, item).ItemTail(nyxColorChosen))
					return nil
				}),
		},
	})
}

// nyxColorChosen adds the mana once the color (the tail's Word) is chosen,
// counting the planar controller's devotion to it then. An on-disk key:
// never renamed, never reused.
var nyxColorChosen = CardTail("then/nyx-devotion-mana", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.Word == "" {
		return nil
	}
	n := devotionTo(ctx.Game, ctx.Controller(), o.Tail.Word)
	if n <= 0 {
		return nil
	}
	return AddMana{Amount: ManaN(n, o.Tail.Word)}.Apply(ctx)
})
