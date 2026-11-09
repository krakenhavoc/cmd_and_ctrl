package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul Search — Sorcery {W}{B} (Murders at Karlov Manor, #4469):
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. Exile that card. If the card's mana value is 1 or less, create
//	 a 1/1 white and black Spirit creature token with flying."
//
// The reveal-and-choose is ChooseFromRevealedHand with Exile (ADR 0116).
// The rider is its keyed continuation, run once the card has moved: the
// mana value is read from the chosen card as it was in the hand (a copy
// read before it moved, so last-known information, CR 608.2h), and a
// hand with no nonland card chose nothing, so no token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58f12f0d-e2c1-48f9-9e27-d1de8db743cb",
		Name:         "Soul Search",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: Nonland(),
				Label:  "nonland card",
				Exile:  true,
				Then:   soulSearchSpirit,
			}.Apply(ctx)
		},
	})
}

var soulSearchSpirit = RevealedPickThen("revealed-pick/soul-search-spirit",
	func(ctx *Context, pick game.RevealedPick) error {
		if len(pick.Chosen) == 0 || pick.Chosen[0].ManaValue() > 1 {
			return nil
		}
		return CreateToken{Template: TokenCard("1/1 white and black Spirit with flying"), N: 1}.Apply(ctx)
	})
