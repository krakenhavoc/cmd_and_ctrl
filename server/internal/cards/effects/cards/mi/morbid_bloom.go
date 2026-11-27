package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Morbid Bloom — Sorcery {4}{B}{G} (Alara Reborn, #5586):
//
//	"Exile target creature card from a graveyard, then create X 1/1 green
//	 Saproling creature tokens, where X is the exiled card's toughness."
//
// The target is a creature card in any graveyard, chosen as the spell is cast
// and judged again as it resolves (CR 608.2b); a gone target does nothing, no
// tokens included. X is the toughness the card had in the graveyard, read before
// it is exiled (a card in a graveyard has its printed toughness). A card with
// toughness 0 makes no tokens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aacec73f-9c60-4c3f-b323-6261eab342af",
		Name:         "Morbid Bloom",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card in a graveyard", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok {
				return nil
			}
			tail := ctx.ItemTail(morbidBloomSaprolings)
			if c, found := ctx.Game.LookupCardForEffect(t.ID); found {
				tail.N = max(c.CurrentToughness(), 0)
			}
			return effects.ExileTarget{Target: t.ID, ThenTail: tail}.Apply(ctx)
		},
	})
}

// morbidBloomSaprolings is the rest of the spell, once the card has reached
// exile: X Saprolings, X the toughness it had in the graveyard, riding the tail
// as N. An on-disk key: never renamed, never reused.
var morbidBloomSaprolings = effects.CardTail("then/morbid-bloom-saprolings", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: o.Tail.N}.Apply(ctx)
})
