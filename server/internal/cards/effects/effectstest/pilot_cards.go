package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func PushPilotCard(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		OracleID:   oracle,
		TypeLine:   typeLine,
		Power:      power,
		Toughness:  toughness,
		Owner:      owner,
		Controller: owner,
	})
}

// pushTestVehicle is a crew-3 Vehicle carried through the timestamping
// path, so layer-7 effects see it.
func PushTestVehicle(g *game.Game, owner uuid.UUID, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Cost Power Vehicle",
		OracleID:   CostPowerVehicleOracle,
		TypeLine:   "Artifact — Vehicle",
		Power:      power,
		Toughness:  toughness,
		Owner:      owner,
		Controller: owner,
	})
}
