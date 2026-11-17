package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akroma's Blessing — Instant {2}{W} (Onslaught):
//
//	"Choose a color. Creatures you control gain protection from the
//	 chosen color until end of turn.
//	 Cycling {W}"
//
// The creatures are the caster's as the spell resolves (CR 611.2c), so
// one that enters later is not protected. The colour is chosen then too,
// as the colour they are protected FROM (ColorForProtection, #780), and
// the protection is Bathe in Light's tail over that set. A caster with no
// creature is asked nothing. Cycling is the keyword's own hand ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3fc6888-0a4b-4e1d-b709-042b8b6b1a04",
		Name:         "Akroma's Blessing",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() && c.Controller == item.Controller {
					ids = append(ids, c.InstanceID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			then := ctx.ItemTail(radianceProtectionFromChosenColor)
			then.Carry = ids
			then.Words = []string{"Akroma's Blessing"}
			ChooseColorTail(game.ColorForProtection, ctx.Game, item.Controller, ctx.Source(), "Akroma's Blessing — choose a color", then)
			return nil
		},
		Activated: []ActivatedAbility{Cycling("{W}")},
	})
}
