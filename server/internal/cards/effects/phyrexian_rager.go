package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Rager — Creature — Phyrexian Horror {2}{B}, 2/2:
//
//	"When this creature enters, you draw a card and you lose 1 life."
//
// Dusk Legion Zealot's shape at a bigger body (#4092, Captain
// N'ghathrod's list). "When this creature enters" is a TRIGGER that uses
// the stack, so it can be responded to. Both halves are one resolution,
// so at 1 life you still draw the card and only then, as state-based
// actions run (CR 704.3), lose the game.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e409c9be-0c9a-43c3-adb4-8c47afc1d551",
		Name:         "Phyrexian Rager",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Phyrexian Rager — draw a card, lose 1 life", drawACardAndLoseALife),
		},
	})
}

// drawACardAndLoseALife is "you draw a card and you lose 1 life" as an
// enters-trigger body: the draw, then the loss, for the ability's
// controller. Shared by Phyrexian Rager and Dusk Legion Zealot.
func drawACardAndLoseALife(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: item.Controller, Amount: -1}.Apply(ctx)
}
