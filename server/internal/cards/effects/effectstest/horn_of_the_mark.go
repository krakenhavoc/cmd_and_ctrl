package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hornTopFive stacks five named cards on top of `p`'s library and
// returns them top-first.
func HornTopFive(p *game.Player, names []struct{ Name, TypeLine string }) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	// PushTop, so the LAST pushed is on top: walk backwards.
	for i := len(names) - 1; i >= 0; i-- {
		id := uuid.New()
		p.Library.PushTop(game.Card{
			InstanceID: id, Name: names[i].Name, TypeLine: names[i].TypeLine,
			Owner: p.ID, Controller: p.ID,
		})
		ids[i] = id
	}
	return ids
}

var HornFive = []struct{ Name, TypeLine string }{
	{"Mountain", "Basic Land — Mountain"},
	{"Rohirrim Rider", "Creature — Human Knight"},
	{"Shock", "Instant"},
	{"Eomer of the Riddermark", "Creature — Human Noble"},
	{"Sol Ring", "Artifact"},
}

func HornRevealed(g *game.Game, id uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards && ev.CardID == id {
			return true
		}
	}
	return false
}
