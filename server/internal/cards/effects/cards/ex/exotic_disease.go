package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exotic Disease — Sorcery {4}{B} (Planeshift, S82 OLD b642, #5615):
//
//	"Domain — Target player loses X life and you gain X life, where X is the
//	 number of basic land types among lands you control."
//
// X is the caster's DomainCount, read once as the spell resolves (CR 608.2c). The
// gain is X whether or not the loss happened (CR 608.2c): a target that is
// illegal makes the whole spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9480e142-289a-439d-b6a5-2666fb0e5be3",
		Name:         "Exotic Disease",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			p := effects.TargetedPlayer(ctx)
			if p == uuid.Nil {
				return nil
			}
			x := effects.DomainCount(ctx.Game, ctx.Controller())
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), p, -x); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: x}.Apply(ctx)
		},
	})
}
