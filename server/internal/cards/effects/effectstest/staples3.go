package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	PitilessPlundererOracle    = "a784481f-eccb-4112-bb38-04a659319660"
	BastionOfRemembranceOracle = "c7f33cea-2ec8-4081-9208-a5b1d86721b3"
	RavenousChupacabraOracle   = "7b459306-149b-4f43-abc1-2dd70c748c0e"
	AuraShardsOracle           = "8d03d050-391c-4311-8c42-4ee632d40fdc"
	VampiricRitesOracle        = "660de988-b6fb-4f36-8006-42af3e7f908d"
	AmbitionsCostOracle        = "84de4fec-2f38-4293-93d3-b3882c5aac14"
)

// seedPermanentWithOracle puts a catalog permanent on the battlefield
// so its triggered abilities are live without casting it.
func SeedPermanentWithOracle(g *game.Game, owner uuid.UUID, name, typeLine, oracleID string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Power:      1,
		Toughness:  4,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// knownToTable marks every seat a knower of a battlefield card seeded
// straight into the zone, which is what the engine does when a card
// really enters (CR 400.2: the battlefield is public). The seed
// helpers skip it, and a view test needs it: since #95 the wire
// strips mana abilities and the rest of the catalog surface from a
// card its viewer does not know.
func KnownToTable(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID != id {
				continue
			}
			known := make(map[uuid.UUID]bool, len(g.Seats))
			for _, p := range g.Seats {
				known[p.ID] = true
			}
			g.Battlefield.Cards[i].KnownBy = known
		}
	})
}
