package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func SemestersEndCard(t *testing.T, g *game.Game, name string) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("%s is not on the battlefield", name)
	return game.Card{}
}

func SemestersEndWalker(g *game.Game, owner uuid.UUID, name, typeLine string, printed, current, p, tough int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: "oracle-" + name,
		StartingLoyalty: printed, Power: p, Toughness: tough,
		Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLoyalty: current},
	})
}

func SemestersEndCast(t *testing.T, g *game.Game, targets ...uuid.UUID) {
	t.Helper()
	refs := make([]game.TargetRef, 0, len(targets))
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	CastCatalogSpell(t, g, "Semester's End", "Instant", GuffaSemestersOracle, refs)
	PassPriorityAroundTable(t, g)
	for _, id := range targets {
		if g.Battlefield.Contains(id) {
			t.Fatalf("target %s was not exiled", id)
		}
	}
}
