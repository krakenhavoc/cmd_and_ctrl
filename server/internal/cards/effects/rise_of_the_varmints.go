package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rise of the Varmints — Sorcery {3}{G} (Outlaws of Thunder Junction,
// #4468):
//
//	"Create X 2/1 green Varmint creature tokens, where X is the number of
//	 creature cards in your graveyard.
//	 Plot {2}{G} (You may pay {2}{G} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// X is counted as the spell resolves (b11CreatureCardsInGraveyard: cards
// with the creature type in your graveyard, tokens never among them), and
// that many 2/1 green Varmints are created at once. With no creature card
// there, nothing is created. Plot is the ordinary special action (CR
// 702.170).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "d981977d-acfa-432a-b1d7-1808a89a5998",
		Name:           "Rise of the Varmints",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{2}{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := b11CreatureCardsInGraveyard(ctx.Game, item.Controller)
			if n <= 0 {
				return nil
			}
			return CreateToken{Template: TokenCard("2/1 green Varmint"), N: n}.Apply(ctx)
		},
	})
}
