package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2S33YourCommanderAttacked is "whenever your commander attacks": a
// commander its controller owns is declared as an attacker, by that
// controller. An attack on a planeswalker or battle counts, since the
// clause names no player (CR 508.1m, CR 508.3).
func cmdr2S33YourCommanderAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCommander && c.Owner == source.Controller
}
