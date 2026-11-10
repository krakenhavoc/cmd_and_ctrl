package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2_b05_helpers.go — shared bodies for the Merfolk batch (#5368).

// cmdr2B05MerfolkFilter is "Merfolk creatures", every Merfolk or only
// the source controller's, and with or without the source itself.
func cmdr2B05MerfolkFilter(others, yours bool) TribeFilter {
	return TribeFilter{Tribes: []string{"Merfolk"}, Others: others, YoursOnly: yours}
}

// cmdr2B05ControlsOtherMerfolk reports whether the source's controller
// controls at least n Merfolk creatures other than the source. It reads
// the layered view on every recompute (CR 611.3a), changelings included.
func cmdr2B05ControlsOtherMerfolk(g *game.Game, source *game.Card, n int) bool {
	f := cmdr2B05MerfolkFilter(true, true)
	count := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		c := c
		if f.Matches(&c, g, source) {
			count++
		}
	}
	return count >= n
}

// cmdr2B05SelfKeywordWhileOtherMerfolk is "<this> has <keyword> as long
// as you control at least n other Merfolk".
func cmdr2B05SelfKeywordWhileOtherMerfolk(n int, keyword string) game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfCreature(target, g, source) && cmdr2B05ControlsOtherMerfolk(g, source, n)
	}, keyword)
	ab.Label = "Has " + keyword + " as long as you control at least " + numberWord(n) + " other Merfolk."
	return ab
}
