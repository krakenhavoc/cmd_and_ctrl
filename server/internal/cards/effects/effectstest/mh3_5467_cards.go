package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mh3b1HandNames lists the names of the cards in p's hand.
func Mh3b1HandNames(p *game.Player) map[string]bool {
	out := map[string]bool{}
	for _, c := range p.Hand.Cards {
		out[c.Name] = true
	}
	return out
}
