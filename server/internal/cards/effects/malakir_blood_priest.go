package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malakir Blood-Priest — Creature — Vampire Cleric {1}{B}, 2/1 (Zendikar
// Rising, #5508):
//
//	"When this creature enters, each opponent loses X life and you gain X
//	 life, where X is the number of creatures in your party. (Your party
//	 consists of up to one each of Cleric, Rogue, Warrior, and Wizard.)"
//
// X is the party (CR 700.8, finB9PartySize) as the trigger resolves (CR
// 700.8a), the Blood-Priest itself being the Cleric if nothing else is.
// Each opponent loses X and you gain X, not X per opponent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a3cb6fe-24ef-4c09-8ff8-f303fc472a8b",
		Name:         "Malakir Blood-Priest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Malakir Blood-Priest — each opponent loses X life and you gain X life, X the creatures in your party",
				Do(Instruction(func(ctx *Context) error {
					x := finB9PartySize(ctx.Game, ctx.Controller())
					if err := (EachOpponentLosesLife{N: x}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Amount: x}.Apply(ctx)
				}))),
		},
	})
}
