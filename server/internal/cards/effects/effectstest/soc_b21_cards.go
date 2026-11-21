package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Borrowed Knowledge ------------------------------------------------------

func SocB21SetHand(p *game.Player, n int) {
	p.Hand.Cards = nil
	for i := 0; i < n; i++ {
		c := game.NewCard("Hand Card", p.ID)
		c.TypeLine = "Sorcery"
		p.Hand.PushTop(c)
	}
}
