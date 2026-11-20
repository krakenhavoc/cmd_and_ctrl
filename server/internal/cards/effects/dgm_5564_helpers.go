package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// dgm_5564_helpers.go — the shared shape of Dragon's Maze's Maze cycle
// (#5564): Maze Abomination, Maze Behemoth, Maze Glider, Maze Rusher and
// Maze Sentinel each print "Multicolored creatures you control have
// <keyword>."

// dgmMulticoloredCreaturesHave is that sentence as a layer-6 grant
// (CR 613.1f). Colours are settled in layer 5, before layer 6, so the
// predicate reads the colours of the characteristic being built: a
// creature made two colours by an effect gets the keyword, and one made
// monocolored loses it (CR 105.2b). The source counts when it is itself
// multicolored.
func dgmMulticoloredCreaturesHave(keyword string) game.StaticAbility {
	return Labelled(KeywordGrant(dgmMulticoloredCreatureYouControl, keyword),
		"Multicolored creatures you control have "+keyword+".")
}

// dgmMulticoloredCreatureYouControl is "multicolored creatures you
// control".
func dgmMulticoloredCreatureYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller && len(target.EffectiveColors()) >= 2
}
