package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_triggers_test.go — the catalog half of #925: the two
// cards whose printed abilities watch from a graveyard. The engine
// half (the zone index, the harvest, the CR 108.4 owner rule, the
// undo across the prompt) is in game/trigger_zones_test.go.

const (
	NarcomoebaOracle = "dc65bb62-4ba2-4ec1-b3b9-9c51e64cbfc8"
	BloodghastOracle = "e97f9c2b-b41e-4f36-9245-77c0ac125647"
)

// putOnTopOfLibrary places a catalog card on top of a seat's library
// so a mill can put it into that seat's graveyard from there — the
// exact zone pair Narcomoeba's trigger names.
func PutOnTopOfLibrary(g *game.Game, p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracle,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

func InZone(z *game.Zone, id uuid.UUID) bool {
	if z == nil {
		return false
	}
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// onBattlefieldByName reports whether a card with that name is on the
// battlefield. Name rather than instance ID because a return from a
// graveyard to the battlefield is a new object (CR 400.7).
func OnBattlefieldByName(g *game.Game, name string) bool {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}
