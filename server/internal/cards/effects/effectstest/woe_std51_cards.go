package effectstest

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std51Me is the active seat; std51Opp the next.
func Std51Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// --- Matzalantli, the Great Door // The Core ---------------------------------------

func Std51Graveyard(p *game.Player, typeLines ...string) []uuid.UUID {
	var ids []uuid.UUID
	for i, tl := range typeLines {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: "Grave " + string(rune('A'+i)), TypeLine: tl, Owner: p.ID, Controller: p.ID,
		})
		ids = append(ids, id)
	}
	return ids
}

func Std51Pay(_ *game.Game, p *game.Player, mana string) {
	for _, sym := range strings.Split(strings.ReplaceAll(mana, "}", ""), "{")[1:] {
		n, err := strconv.Atoi(sym)
		if err != nil {
			p.ManaPool.AddMana(game.ManaToken{Color: sym})
			continue
		}
		for i := 0; i < n; i++ {
			p.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
	}
}
