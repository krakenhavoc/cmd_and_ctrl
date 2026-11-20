package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Divine Smite — Instant {1}{W}:
//
//	"Target creature or planeswalker an opponent controls phases out. If
//	 that permanent is black, exile it instead. (If it phases out, treat it
//	 and anything attached to it as though they don't exist until its
//	 controller's next turn.)"
//
// One target, one outcome: read as the spell resolves (CR 608.2b, the
// target must still be legal), the permanent is exiled if it is black then
// (through the layers), and phases out otherwise. Phasing out is not
// leaving the battlefield (CR 702.26d), so the phase-out is no removal; the
// exile is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b06b242-caed-4c8f-b5ab-30e86061286e",
		Name:         "Divine Smite",
		Completeness: effects.CompletenessFull,
		Targets: effects.Harmful(effects.TargetPermanent("target creature or planeswalker an opponent controls",
			effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()))),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, id := range effects.LegalTargetCards(item, ctx.Game) {
				c, ok := ctx.Game.LookupCardForEffect(id)
				if ok && c.HasColor(ctx.TextColor("B")) {
					if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
					continue
				}
				if err := (effects.PhaseOut{Targets: []uuid.UUID{id}}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
