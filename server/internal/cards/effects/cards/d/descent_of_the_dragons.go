package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Descent of the Dragons — Sorcery {4}{R}{R} (Dragons of Tarkir):
//
//	"Destroy any number of target creatures. For each creature destroyed this way,
//	 its controller creates a 4/4 red Dragon creature token with flying."
//
// The tokens go to the controller each creature had as it was destroyed (the
// batch's last-known information, CR 603.10), and only for the creatures that
// were really destroyed: an indestructible or regenerated creature stays and
// earns no Dragon.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d32c6f64-c4b0-4451-8986-a0abc1fc2bb3",
		Name:         "Descent of the Dragons",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect(ids, ctx.Tail(descentOfTheDragonsDragons))
		},
	})
}

var descentOfTheDragonsDragons = effects.MassTail("mass/descent-of-the-dragons-dragons",
	func(ctx *effects.Context, swept []game.Card, _ int) error {
		for _, c := range swept {
			tok := ctx.TextToken(effects.TokenCard("4/4 red Dragon with flying"))
			if err := ctx.Game.CreateTokenForEffect(c.Controller, tok, 1); err != nil {
				return err
			}
		}
		return nil
	})
