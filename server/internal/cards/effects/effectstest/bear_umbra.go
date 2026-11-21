package effectstest

import (
	"testing"

	"github.com/google/uuid"

	cardsb "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/b"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bear_umbra_test.go pins Bear Umbra and the umbra armor seam (CR
// 702.89a): a destruction of the enchanted creature is replaced by
// "remove all damage from it and destroy this Aura"; sacrifice and
// zero toughness are not destructions and are not replaced.

// umbraBear puts a 2/2 (or the given base toughness) under me's control
// with a Bear Umbra on it, and returns both IDs.
func UmbraBear(t *testing.T, g *game.Game, owner uuid.UUID, toughness int) (bear, umbra uuid.UUID) {
	t.Helper()
	bear = PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: TestCreatureTypeLine,
		Power: 2, Toughness: toughness, Owner: owner, Controller: owner,
	})
	umbra = PushAuraOnLand(t, g, owner, "Bear Umbra", cardsb.BearUmbraOracleID, bear)
	return bear, umbra
}

func UmbraInGraveyard(p *game.Player, id uuid.UUID) bool { return p.Graveyard.Contains(id) }
