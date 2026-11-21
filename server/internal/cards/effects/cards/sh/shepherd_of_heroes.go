package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "6a732900-6529-4d5c-ab94-52bd0048d483",
		Name:            "Shepherd of Heroes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shepherd of Heroes — you gain 2 life for each creature in your party",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.GainLife{Amount: 2 * effects.FinB9PartySize(ctx.Game, ctx.Controller())}.Apply(ctx)
				}))),
		},
	})
}
