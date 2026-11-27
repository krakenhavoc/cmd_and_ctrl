package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nylea's Disciple — Creature — Centaur Archer {2}{G}{G}, 3/3 (Theros, S82 OLD b391, #5561):
//
//	"When this creature enters, you gain life equal to your devotion to
//	 green."
//
// Karametra's Acolyte's devotion count (CR 700.5), read as the trigger
// resolves: the Disciple's own {G}{G} is on the battlefield by then, so it
// counts itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aaefdb6f-306e-4fdc-9e49-af7a63e70f06",
		Name:         "Nylea's Disciple",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Nylea's Disciple — you gain life equal to your devotion to green",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.GainLife{Amount: effects.DevotionTo(ctx.Game, ctx.Controller(), ctx.TextColor("G"))}.Apply(ctx)
				}))),
		},
	})
}
