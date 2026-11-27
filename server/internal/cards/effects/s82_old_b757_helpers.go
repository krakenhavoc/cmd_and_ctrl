package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S82OldB757CardsNamedInGraveyards is "the number of cards named <name> in
// all graveyards": every seat's graveyard, counted now.
func S82OldB757CardsNamedInGraveyards(g *game.Game, name string) int {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if game.CardNameMatches(c, name) {
				n++
			}
		}
	}
	return n
}

// S82OldB757TapFirstClauseTarget is "tap target <permanent>" as the first
// step of a two-clause spell: it taps the creature clause 0 picked, and does
// nothing when that pick is gone (CR 608.2b).
func S82OldB757TapFirstClauseTarget(ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return TapTarget{Target: t.ID}.Apply(ctx)
}
