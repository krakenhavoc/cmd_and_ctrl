package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Misfortune's Gain — Sorcery {3}{W} (Portal Three Kingdoms, S82 OLD b713, #5622):
//
//	"Destroy target creature. Its owner gains 4 life."
//
// The owner is read before the destruction (CR 608.2h), so the life still
// finds them once the creature has gone, and it is gained whether or not
// the creature died (an indestructible one stays and its owner still
// gains). Its OWNER, not its controller: a stolen creature pays out to the
// player whose card it is. A target no longer legal makes the spell do
// nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01b778d7-0965-48ab-a81f-e29f8a233d07",
		Name:         "Misfortune's Gain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			owner := c.Owner
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), owner, 4)
		},
	})
}
