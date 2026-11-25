package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gideon's Defeat — {W} Instant (Hour of Devastation):
//
//	"Exile target white creature that's attacking or blocking. If it was a
//	 Gideon planeswalker, you gain 5 life."
//
// The creature is read as it last existed before the exile (CR 608.2h).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9d76570-d951-4c09-aa0e-c204c3407aaf",
		Name:         "Gideon's Defeat",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetCreature("target white creature that's attacking or blocking",
			effects.OfColor("W"), effects.AttackingOrBlocking()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			c, ok := effects.OldB238CardTargetLastKnown(ctx)
			if !ok {
				return nil
			}
			wasGideon := c.IsPlaneswalker() && c.HasSubtype("Gideon")
			if err := (effects.ExileChosenTarget{}).Apply(ctx); err != nil {
				return err
			}
			if !wasGideon {
				return nil
			}
			return effects.GainLife{Amount: 5}.Apply(ctx)
		},
	})
}
