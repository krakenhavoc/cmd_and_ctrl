package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// countYouControl counts the permanents the effect's controller controls
// that match, read as the spell resolves (CR 608.2h: a count in the text
// is taken from the game as it is then). Cleave cards whose bracketed
// words are a filter on that count (Winged Portent, Lantern Flare) share
// it.
func CountYouControl(ctx *Context, match CardPredicate) int {
	you := ctx.Controller()
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == you && (match == nil || match(ctx.Game, you, c)) {
			n++
		}
	}
	return n
}

// firstTargetCard is the card the spell's first target slot names, or
// false when it names a player or nothing.
func FirstTargetCard(item *game.StackItem) (game.TargetRef, bool) {
	if item == nil || len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return game.TargetRef{}, false
	}
	return item.Targets[0], true
}
