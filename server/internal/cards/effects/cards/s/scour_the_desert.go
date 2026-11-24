package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scour the Desert — Sorcery {3}{W}{W} (Modern Horizons 2, #5501):
//
//	"Exile target creature card from your graveyard. Create X 1/1 white Bird
//	 creature tokens with flying, where X is the exiled card's toughness."
//
// X is the card's toughness as it sat in the graveyard (a characteristic-
// defining toughness counts, CR 604.3), read before the exile. The tokens are
// made only if the exile happened; a target that left makes the spell do
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86c9e08a-397a-4c24-8f4c-48a648889399",
		Name:         "Scour the Desert",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			tail := ctx.ItemTail(scourTheDesertBirds)
			tail.N = ctx.Game.ToughnessAnywhereForEffect(card)
			return effects.ExileTarget{Target: id, ThenTail: tail}.Apply(ctx)
		},
	})
}

// scourTheDesertBirds is the exile's continuation: if the card was exiled,
// X Birds, X being the toughness read before the exile and carried as the
// tail's N. An on-disk key: never renamed, never reused.
var scourTheDesertBirds = effects.CardTail("then/scour-the-desert-birds", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) || o.Tail.N <= 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("1/1 white Bird with flying"), N: o.Tail.N}.Apply(ctx)
})
