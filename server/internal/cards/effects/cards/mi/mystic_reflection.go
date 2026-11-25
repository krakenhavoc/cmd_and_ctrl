package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mystic Reflection — Instant {1}{U} (Kaldheim):
//
//	"Choose target nonlegendary creature. The next time one or more
//	 creatures or planeswalkers enter this turn, they enter as copies of
//	 the chosen creature.
//	 Foretell {U}"
//
// The #6160 one-shot entry replacement
// (game.ModNextEntersAsCopies, CR 614.1c): the target is chosen as the
// spell resolves, and the next entry of one or more creatures or
// planeswalkers this turn — under anyone's control, cast, put or created
// — enters as copies of it (CR 707.2), every permanent of that one
// simultaneous entry and nothing after it. The copiable values are read as
// they enter, from the creature's last-known information if it has left
// by then (CR 608.2h). A planeswalker that enters this way is a copy of
// the creature and has no loyalty. Foretell is the shared special action
// (CR 702.143).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "491dc110-b459-412e-8a0b-45e5744ee8d1",
		Name:           "Mystic Reflection",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{U}")},
		Targets:        effects.TargetCreature("target nonlegendary creature", effects.Not(effects.Legendary())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			ctx.Game.NextCreaturesOrPlaneswalkersEnterAsCopiesThisTurnForEffect(ctx.Source(), ctx.Controller(), targets[0].ID,
				"Mystic Reflection — the next creatures or planeswalkers to enter this turn enter as copies")
			return nil
		},
	})
}
