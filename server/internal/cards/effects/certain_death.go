package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Certain Death — Sorcery {5}{B} (Eldritch Moon, #5541):
//
//	"Destroy target creature. Its controller loses 2 life and you gain 2
//	 life."
//
// The creature's controller is read before the destruction, so the life
// loss still finds them once it has gone, and it is lost whether or not
// the creature died (an indestructible one stays and its controller
// still loses). A target that is no longer legal makes the spell do
// nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d4b8a31-6500-4e3d-bcaf-f6f184d7743e",
		Name:         "Certain Death",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve:    certainDeathResolve,
	})
}

// certainDeathResolve destroys the target, then its controller loses 2
// life and the caster gains 2, in printed order (CR 608.2c).
func certainDeathResolve(_ *game.StackItem, ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	controller := c.Controller
	if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), controller, -2); err != nil {
		return err
	}
	return GainLife{Amount: 2}.Apply(ctx)
}
