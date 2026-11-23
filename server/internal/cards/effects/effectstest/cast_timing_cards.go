package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ctHandCard seeds one card into a seat's hand.
func CtHandCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// ctStoredTimings counts the cast-timing statements stored on a
// player. Since #1195 folded onto #1197's registry a granted timing
// statement is a game.PlayerStatic carrying a rule and no Keyword, so
// it shares one slice, one sweep and one duration with "you gain
// protection from everything until your next turn".
func CtStoredTimings(p *game.Player) int {
	n := 0
	for _, st := range p.Statics {
		if st.Timing.Timing != game.TimingNormal {
			n++
		}
	}
	return n
}
