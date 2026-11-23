package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushHandCard puts a card into a player's hand and returns its ID.
func PushChromeMoxHandCard(p *game.Player, name, typeLine string, colors []string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Colors:     colors,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}
