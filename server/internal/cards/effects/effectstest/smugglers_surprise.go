package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const SmugglersSurpriseOracle = "605f719d-01f8-406c-9d4c-3f4992c6a69f"

// ssCard pushes a card with the given P/T onto the top of a zone.
func SsCard(z *game.Zone, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	id := uuid.New()
	z.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
	return id
}
