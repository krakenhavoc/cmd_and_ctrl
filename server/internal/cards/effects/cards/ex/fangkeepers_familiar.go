package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fangkeeper's Familiar — Creature — Snake {1}{B}{G}{U}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flash
//	 When this creature enters, choose one —
//	 • You gain 3 life and surveil 3.
//	 • Destroy target enchantment.
//	 • Counter target creature spell."
//
// Apothecary Stomper's modal enters trigger: the mode, and its target,
// are chosen as the trigger goes on the stack and the target is re-checked
// at resolution (CR 608.2b). The Familiar was cast with flash, so a
// creature spell is a fine target while it resolves from the stack on top
// of it. The first mode gains the life, then surveils (the prompt is the
// last step, so the surveil happens after the life gain).
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Fangkeeper's Familiar — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("You gain 3 life and surveil 3.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			if err := (effects.GainLife{Player: item.Controller, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			return effects.Surveil{Player: item.Controller, N: 3}.Apply(ctx)
		}),
		effects.ModeDoing("Destroy target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment()), effects.DestroyTheModesTarget),
		effects.ModeDoing("Counter target creature spell.", effects.TargetSpell("target creature spell", effects.Creature()), effects.CounterTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:        "06343800-b555-4a13-ad4d-19e37d200138",
		Name:            "Fangkeeper's Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
