package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func Deck3655GraveyardCreature(p *game.Player, name string, power int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear",
		Power: power, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}
