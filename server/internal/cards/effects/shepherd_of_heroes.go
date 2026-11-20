package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shepherd of Heroes — Creature — Angel Cleric {4}{W}, 3/4 (Zendikar
// Rising, #5508):
//
//	"Flying
//	 When this creature enters, you gain 2 life for each creature in your
//	 party. (Your party consists of up to one each of Cleric, Rogue,
//	 Warrior, and Wizard.)"
//
// Flying is printed. The party (CR 700.8, finB9PartySize) is counted as
// the trigger resolves (CR 700.8a), the Shepherd itself being the Cleric
// if nothing else is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6a732900-6529-4d5c-ab94-52bd0048d483",
		Name:            "Shepherd of Heroes",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Shepherd of Heroes — you gain 2 life for each creature in your party",
				Do(Instruction(func(ctx *Context) error {
					return GainLife{Amount: 2 * finB9PartySize(ctx.Game, ctx.Controller())}.Apply(ctx)
				}))),
		},
	})
}
