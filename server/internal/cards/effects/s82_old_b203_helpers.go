package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S82OldB203BasicLandOrJiang is the card filter of Journey for the Elixir
// (#5528): "a basic land card and a card named Jiang Yanggu".
func S82OldB203BasicLandOrJiang(c game.Card) bool {
	return s82OldB203IsBasicLand(c) || game.CardNameMatches(c, "Jiang Yanggu")
}

func s82OldB203IsBasicLand(c game.Card) bool {
	return c.IsLand() && c.HasSupertype("Basic")
}

// SearchRuleBasicLandAndJiangYanggu is Journey for the Elixir's set rule:
// of the cards found, at most one is a basic land card and at most one is
// named Jiang Yanggu. The key is an on-disk identity.
var SearchRuleBasicLandAndJiangYanggu = game.RegisterSearchRule("search/basic-land-and-jiang-yanggu",
	func(picked []game.Card, _ int) bool {
		lands, named := 0, 0
		for _, c := range picked {
			switch {
			case s82OldB203IsBasicLand(c):
				lands++
			case game.CardNameMatches(c, "Jiang Yanggu"):
				named++
			default:
				return false
			}
		}
		return lands <= 1 && named <= 1
	})
