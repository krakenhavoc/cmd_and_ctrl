package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "df9cdc11-cb92-46ce-a83d-eb26e3cd004a",
		Name:         "Another Chance",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Library != nil && p.Library.Size() >= 2 {
				return MayChoice{
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
var anotherChanceMill = MayAnswer("confirm/another-chance-mill", func(ctx *Context, r game.ConfirmAnswer) error {
	if r.Accepted {
		if err := (MillCards{N: 2}).Apply(ctx); err != nil {
			return err
		}
	}
	return anotherChanceReturn(ctx)
})

func anotherChanceReturn(ctx *Context) error {
	return ReturnChosenFromGraveyard{
		Player:   ctx.Controller(),
		Question: "Another Chance — return up to two creature cards from your graveyard to your hand",
		Match:    func(c game.Card) bool { return c.IsCreature() },
		Max:      2,
		ToHand:   true,
	}.Apply(ctx)
}
