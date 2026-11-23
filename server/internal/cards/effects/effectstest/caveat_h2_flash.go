package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3172 cav-h2: Timely Ward's "as though it had flash if it targets a
// commander" (FlashIfTargets).

const TimelyWardOracle = "8e589dca-d31e-4d2f-81b1-18e5f9198795"

func TimelyWardTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, _ := P7Table(t)
	ward := EacHandCard(g, me, "Timely Ward", "Enchantment — Aura", "{2}{W}")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == ward {
				me.Hand.Cards[i].OracleID = TimelyWardOracle
			}
		}
	})
	commander := B12Creature(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
	bear := B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == commander {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	return g, me, ward, commander, bear
}
