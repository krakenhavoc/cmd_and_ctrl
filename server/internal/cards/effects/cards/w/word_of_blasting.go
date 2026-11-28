package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Word of Blasting — Instant {1}{R} (Mercadian Masques, S82 OLD, #5619):
//
//	"Destroy target Wall. It can't be regenerated. Word of Blasting deals damage
//	 equal to that Wall's mana value to the Wall's controller."
//
// The Wall's controller and mana value are read as the spell resolves, before it
// is destroyed (CR 608.2h's last-known information, were it to go first), so the
// damage lands on whoever controlled it even when it was indestructible and
// stayed. A Wall token has mana value 0 and so deals no damage. Regeneration is
// ignored (CR 701.19c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b391f24-992f-4cc0-ac4d-31456983fc03",
		Name:         "Word of Blasting",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target Wall", effects.HasSubtype("Wall")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			controller := c.Controller
			mv, _ := ctx.Game.ManaValueForEffect(c)
			if err := (effects.DestroyTarget{Target: id, CantBeRegenerated: true}).Apply(ctx); err != nil {
				return err
			}
			return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: mv}.Apply(ctx)
		},
	})
}
