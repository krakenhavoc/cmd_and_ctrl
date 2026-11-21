package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gaius van Baelsar — Legendary Creature — Human Soldier {2}{B}{B}, 3/2
// (Final Fantasy, #4462):
//
//	"When Gaius van Baelsar enters, choose one —
//	 • Each player sacrifices a creature token of their choice.
//	 • Each player sacrifices a nontoken creature of their choice.
//	 • Each player sacrifices an enchantment of their choice."
//
// A modal enters trigger (Apothecary Stomper's shape), each mode Accursed
// Marauder's edict over a different set: every player, the controller
// included, is asked in APNAP order, each offered only their own matching
// permanents, and the choices leave together. Gaius is a nontoken creature
// already on the battlefield, so he may be fed to his own second mode.
//
// No simplification.
func init() {
	edict := func(label string, match effects.CardPredicate) func(item *game.StackItem, ctx *effects.Context, _ int) error {
		return func(_ *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.EachPlayerSacrifices{Match: match, Label: label}.Apply(ctx)
		}
	}
	trigger := effects.WhenThisEnters("Gaius van Baelsar — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Each player sacrifices a creature token of their choice.", nil,
			edict("a creature token", effects.And(effects.Creature(), effects.IsTokenPredicate()))),
		effects.ModeDoing("Each player sacrifices a nontoken creature of their choice.", nil,
			edict("a nontoken creature", effects.B04NontokenCreature)),
		effects.ModeDoing("Each player sacrifices an enchantment of their choice.", nil,
			edict("an enchantment", effects.Enchantment())),
	)
	trigger.Purpose.Answers |= game.AnswerRemove
	effects.Register(effects.Spec{
		OracleID:     "86d2a45a-e452-4241-8ab4-28a2bb808ffa",
		Name:         "Gaius van Baelsar",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
