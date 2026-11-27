package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Weed Strangle — Sorcery {3}{B}{B} (Lorwyn, #5593):
//
//	"Destroy target creature. Clash with an opponent. If you win, you gain
//	 life equal to that creature's toughness. (Each clashing player reveals
//	 the top card of their library, then puts that card on their choice of
//	 the top or bottom. A player wins if their card had a greater mana
//	 value.)"
//
// The creature's toughness is read as the spell resolves, before the
// destruction, through the layers (CR 608.2h, so counters and anthems
// count), and carried through the clash. A target that became illegal does
// nothing, clash included (CR 608.2b); a creature that survives (regeneration,
// indestructible) still gives its toughness on a win.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af64ce09-b7d9-4db9-905e-c4356f242ced",
		Name:         "Weed Strangle",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			toughness := 0
			if c, found := ctx.Game.LookupCardForEffect(id); found {
				if info, live := ctx.Game.PermanentForEffect(game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}); live {
					toughness = max(info.Toughness, 0)
				}
			}
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			then := ctx.ItemTail(weedStrangleClashed)
			then.N = toughness
			return effects.ClashWithAnOpponent{Then: then}.Apply(ctx)
		},
	})
}

// weedStrangleClashed is "If you win, you gain life equal to that creature's
// toughness" (the tail's N). An on-disk key.
var weedStrangleClashed = effects.CardTail("then/weed-strangle-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.WonTheClash(o) || o.Tail.N <= 0 {
		return nil
	}
	return effects.GainLife{Player: ctx.Controller(), Amount: o.Tail.N}.Apply(ctx)
})
