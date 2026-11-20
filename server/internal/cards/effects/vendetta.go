package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "50a16634-ebf7-477f-ad6f-23332599f838",
		Name:         "Vendetta",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		OnResolve:    vendettaResolve,
	})
}

// vendettaResolve is "Destroy target … creature. It can't be regenerated.
// You lose life equal to that creature's toughness.", shared with Devour in
// Shadow, which prints it for any creature.
func vendettaResolve(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		toughness := c.CurrentToughness()
		if err := (DestroyTarget{Target: t.ID, CantBeRegenerated: true}).Apply(ctx); err != nil {
			return err
		}
		if toughness <= 0 {
			return nil
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -toughness)
	}
	return nil
}
