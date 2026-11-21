package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auraCast casts the named Aura from the active seat's hand onto
// `host` and settles the spell and whatever it triggers.
func AuraCast(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := CastCatalogSpell(t, g, name, AuraTypeLine, oracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	PassPriorityAroundTable(t, g)
	return aura
}

// auraBear puts a vanilla 2/2 that has been in play a while under owner.
func AuraBear(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// auraRestrictions is the effective restriction set of a permanent.
func AuraRestrictions(t *testing.T, g *game.Game, id uuid.UUID) game.Restriction {
	t.Helper()
	var r game.Restriction
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				r, found = c.Effective().Restrictions, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return r
}

// auraSeedLibrary gives a seat enough library to draw from.
func AuraSeedLibrary(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		PushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
}
