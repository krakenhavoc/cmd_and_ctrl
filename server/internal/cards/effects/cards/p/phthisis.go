package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phthisis — {3}{B}{B}{B}{B} Sorcery:
//
//	"Destroy target creature. Its controller loses life equal to its
//	 power plus its toughness.
//	 Suspend 5—{1}{B}"
//
// The life loss is read off the creature BEFORE it is destroyed: the
// effect uses the creature's last-known power and toughness (CR 608.2h)
// and the player who controlled it, and it happens whether or not the
// destruction did anything (an indestructible creature still costs its
// controller the life). A negative total loses nothing. A target that
// is gone when the spell resolves is illegal and the spell does nothing
// (CR 608.2b). The suspend special action, the time counters and the
// free cast on the last one are the engine's (Rift Bolt, Shivan Meteor).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "530f088d-bfa1-4410-b1aa-ade8abdb2b7a",
		Name:         "Phthisis",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		SpecialActions: []game.SpecialAction{
			effects.Suspend(5, "{1}{B}"),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			loss := card.CurrentPower() + card.CurrentToughness()
			if loss < 0 {
				loss = 0
			}
			// The life loss runs from the destroy's continuation, as Azog's
			// amass does: it does not ask whether anything was destroyed.
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{t.ID},
				effects.WithN(effects.WithCarry(ctx.Tail(phthisisLoseLife), card.Controller), loss))
		},
	})
}

// phthisisLoseLife is the second sentence: the creature's controller
// (carried) loses life equal to its power plus its toughness (N), read
// before the destroy.
var phthisisLoseLife = effects.CardTail("phthisis/controller-loses-life", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), effects.TailCarry0(o), -o.Tail.N)
})
