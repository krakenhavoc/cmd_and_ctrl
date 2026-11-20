package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Another Chance — Instant {2}{B} (The Lost Caverns of Ixalan, #4470):
//
//	"You may mill two cards. Then return up to two creature cards from
//	 your graveyard to your hand."
//
// The mill is a yes/no asked as the spell resolves, and only when the
// library holds two cards (a player can't choose to mill more than their
// library holds, per the ruling). Whatever the answer, the second sentence
// follows: the offer is every creature card in the graveyard as it is now,
// the just-milled ones included, with a floor of zero (Grapple with the
// Past's shape, #2941).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df9cdc11-cb92-46ce-a83d-eb26e3cd004a",
		Name:         "Another Chance",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Library != nil && p.Library.Size() >= 2 {
				return effects.MayChoice{
					Question: "Another Chance — mill two cards?",
					Then:     anotherChanceMill,
				}.Apply(ctx)
			}
			return anotherChanceReturn(ctx)
		},
	})
}

// anotherChanceMill is the answer to the mill question: mill on a yes, then
// the return either way.
var anotherChanceMill = effects.MayAnswer("confirm/another-chance-mill", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if r.Accepted {
		if err := (effects.MillCards{N: 2}).Apply(ctx); err != nil {
			return err
		}
	}
	return anotherChanceReturn(ctx)
})

func anotherChanceReturn(ctx *effects.Context) error {
	return effects.ReturnChosenFromGraveyard{
		Player:   ctx.Controller(),
		Question: "Another Chance — return up to two creature cards from your graveyard to your hand",
		Match:    func(c game.Card) bool { return c.IsCreature() },
		Max:      2,
		ToHand:   true,
	}.Apply(ctx)
}
