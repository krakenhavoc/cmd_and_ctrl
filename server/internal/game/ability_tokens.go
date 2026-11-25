package game

// AbilityTokensOf is the ability tokens of `c` in the zone rules HasKeyword
// documents, in ability order: the layered list on the battlefield, the
// card's printed keywords, keyword counters and the catalog's printed
// keywords elsewhere (Cairn Wanderer's "the same is true for … landwalk,
// protection": it takes the very token the card in the graveyard has).
// A nil card has none.
func AbilityTokensOf(c *Card) []string {
	var out []string
	forEachAbilityToken(c, func(tok string) bool {
		out = append(out, tok)
		return true
	})
	return out
}
