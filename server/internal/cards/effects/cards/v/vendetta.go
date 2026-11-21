package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vendetta — Instant {B} (Rise of the Eldrazi, #5581):
//
//	"Destroy target nonblack creature. It can't be regenerated. You lose
//	 life equal to that creature's toughness."
//
// The toughness is read as the spell resolves, before the destruction, which
// is the creature as it last existed if it is destroyed (CR 608.2h). The life
// is lost even if the creature isn't destroyed (indestructible).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50a16634-ebf7-477f-ad6f-23332599f838",
		Name:         "Vendetta",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				toughness := c.CurrentToughness()
				if err := (effects.DestroyTarget{Target: t.ID, CantBeRegenerated: true}).Apply(ctx); err != nil {
					return err
				}
				if toughness <= 0 {
					return nil
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -toughness)
			}
			return nil
		},
	})
}
