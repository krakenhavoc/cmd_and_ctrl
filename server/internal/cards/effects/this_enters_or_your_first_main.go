package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// this_enters_or_your_first_main.go — a trigger condition two cards share
// (Võx, Child of the Collective and Crack in Time).

// ThisEntersOrYourFirstMain is this permanent entering, or the beginning of its
// controller's precombat main phase.
func ThisEntersOrYourFirstMain(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventETB:
		return Self(ev, source, lki, g)
	case game.EventBeginPrecombatMain:
		return ByYou(ev, source, lki, g)
	}
	return false
}
