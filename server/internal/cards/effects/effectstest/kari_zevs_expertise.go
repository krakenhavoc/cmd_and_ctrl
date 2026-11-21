package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const KariZevsExpertiseOracle = "e0de0121-6185-45c5-9d88-a415b07347e0"

// pushHandCardWithManaCost seeds a card into a player's hand with a
// real mana value, for the free-cast pick's mana-value ceiling.
func PushHandCardWithManaCost(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}
