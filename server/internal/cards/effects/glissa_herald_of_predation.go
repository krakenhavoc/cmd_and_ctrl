package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glissa, Herald of Predation — Legendary Creature — Phyrexian Zombie
// Elf {3}{B}{G}, 3/5 (March of the Machine):
//
//	"At the beginning of combat on your turn, choose one —
//	 • Incubate 2 twice.
//	 • Transform all Incubator tokens you control.
//	 • Phyrexians you control gain first strike and deathtouch until end
//	   of turn."
//
// A modal trigger with no targets, so the mode is chosen as it goes on
// the stack (CR 700.2) and the Incubators it names are read at
// resolution. "Incubate 2 twice" is two separate incubates
// (Incubate.Times), each its own two-counter Incubator.
// "Transform all" reads the set once, as the ability resolves, and turns
// over each Incubator token, front face up, the controller has.
//
// The third bullet gives first strike and deathtouch to the Phyrexian
// creatures the controller has as the ability resolves (CR 611.2c: the
// set is fixed then; a Phyrexian that arrives later this turn does not
// get them). A transformed Incubator is a Phyrexian creature and so is
// in the set.
//
// No simplification.
func init() {
	trigger := AtBeginningOfYourCombat("Glissa, Herald of Predation — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Incubate 2 twice.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return Incubate{N: 2, Times: 2}.Apply(ctx)
			}),
		ModeDoing("Transform all Incubator tokens you control.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return TransformAllIncubators(ctx)
			}),
		ModeDoing("Phyrexians you control gain first strike and deathtouch until end of turn.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return GrantKeywordUntilEOT{
					Match:    And(OfCreatureType("Phyrexian"), YouControl()),
					Keywords: []string{"first strike", "deathtouch"},
					Label:    "Glissa, Herald of Predation — Phyrexians gain first strike and deathtouch",
				}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "9b10e0f7-9414-4769-85d7-c1cd1c74a9a6",
		Name:         "Glissa, Herald of Predation",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
