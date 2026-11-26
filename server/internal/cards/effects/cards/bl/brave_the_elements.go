package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brave the Elements — Instant {W} (Commander 2014):
//
//	"Choose a color. White creatures you control gain protection from
//	 the chosen color until end of turn."
//
// The colour is chosen as the spell resolves; the white creatures you
// control are read then (CR 608.2h) and locked (CR 611.2c), so a creature
// that turns white later in the turn is not covered. Bathe in Light's
// tail grants the protection.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d73a51a-1c8e-40cd-8124-aeff406a0884",
		Name:         "Brave the Elements",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl(), effects.OfColor("W"))) {
				ids = append(ids, c.InstanceID)
			}
			if len(ids) == 0 {
				return nil
			}
			then := ctx.ItemTail(effects.RadianceProtectionFromChosenColor)
			then.Carry = ids
			then.Words = []string{"Brave the Elements"}
			effects.ChooseColorTail(game.ColorForProtection, ctx.Game, item.Controller, ctx.Source(), "Brave the Elements — choose a color", then)
			return nil
		},
	})
}
