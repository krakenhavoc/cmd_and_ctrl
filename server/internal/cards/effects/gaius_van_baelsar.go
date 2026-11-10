package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	edict := func(label string, match CardPredicate) func(item *game.StackItem, ctx *Context, _ int) error {
		return func(_ *game.StackItem, ctx *Context, _ int) error {
			return EachPlayerSacrifices{Match: match, Label: label}.Apply(ctx)
		}
	}
	trigger := WhenThisEnters("Gaius van Baelsar — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Each player sacrifices a creature token of their choice.", nil,
			edict("a creature token", And(Creature(), IsTokenPredicate()))),
		ModeDoing("Each player sacrifices a nontoken creature of their choice.", nil,
			edict("a nontoken creature", b04NontokenCreature)),
		ModeDoing("Each player sacrifices an enchantment of their choice.", nil,
			edict("an enchantment", Enchantment())),
	)
	trigger.Purpose.Answers |= game.AnswerRemove
	Register(Spec{
		OracleID:     "86d2a45a-e452-4241-8ab4-28a2bb808ffa",
		Name:         "Gaius van Baelsar",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
