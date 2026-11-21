package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// banon_the_returners_leader_test.go — #3769: Banon's graveyard cast is
// limited to cards put there THIS TURN from anywhere other than the
// battlefield, once per turn, on your own turn.

const BanonOracle = "1b62e88f-d544-4c01-a5a2-1c185a3c9c1e"

// banonHandCard puts a {0} card in a seat's hand.
func BanonHandCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{0}",
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID,
		KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

func GraveyardCard(p *game.Player, id uuid.UUID) (game.Card, bool) {
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}
