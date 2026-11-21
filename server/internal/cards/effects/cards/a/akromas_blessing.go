package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "a3fc6888-0a4b-4e1d-b709-042b8b6b1a04",
		Name:         "Akroma's Blessing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() && c.Controller == item.Controller {
					ids = append(ids, c.InstanceID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			then := ctx.ItemTail(effects.RadianceProtectionFromChosenColor)
			then.Carry = ids
			then.Words = []string{"Akroma's Blessing"}
			effects.ChooseColorTail(game.ColorForProtection, ctx.Game, item.Controller, ctx.Source(), "Akroma's Blessing — choose a color", then)
			return nil
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{W}")},
	})
}
