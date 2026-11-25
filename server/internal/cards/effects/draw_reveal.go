package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// draw_reveal.go — #6238 (Booby Trap): the trigger half of "The chosen
// player reveals each card they draw. When the chosen player draws a card
// with the chosen name, ...". The reveal itself is a static
// (Spec.RevealsChosenPlayersDraws, game/draw_reveal.go).

// ChosenPlayerDrawsChosenName is the When of "When the chosen player draws
// a card with the chosen name": an EventDrawCard whose drawer is the
// source's ChosenPlayer and whose card bears the source's ChosenName
// (compared as every chosen name is, game.CardNameMatches: every face,
// case-insensitive). Nothing chosen, nothing triggers.
func ChosenPlayerDrawsChosenName() When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if source == nil || source.ChosenName == "" || ev.Actor != source.ChosenPlayer {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && game.CardNameMatches(c, source.ChosenName)
	}
}
