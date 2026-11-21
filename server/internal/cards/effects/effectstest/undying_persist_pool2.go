package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// undying_persist_pool2_test.go — the second pool PR after #2075 (ADR
// 0113 §4, owner decision 2): the cards that GIVE undying or persist,
// and the rest of the creatures whose text needed a little more. One
// test per card.

const (
	P2UndyingEvil     = "3ceaeb1b-8b25-453d-ac09-e466cfdd9c77"
	P2CauldronHaze    = "56309bbe-86aa-4003-a531-0ef1e318ba15"
	P2CauldronOfSouls = "5811d905-cae5-4f92-a78e-abb43f8864a7"
	P2AntlerSkulkin   = "30572dd1-c982-4872-9f35-a91ed8b238f7"
	P2Rhys            = "9f013b5d-fff0-4c50-8621-158b6dc34834"
	P2DuskLegion      = "f8b12c43-d83a-4e5b-a525-5ee1ae850817"
	P2Endling         = "4da3e510-e495-4077-997f-8efc04f01892"
	P2Mikaeus         = "5d27c63e-d1ef-48af-b51d-01ebc6daeac9"
	P2Rattleblaze     = "313130ab-13db-4d0a-b3f3-b60e060b303c"
	P2Wingrattle      = "034264da-a404-4037-b215-b095bba77490"
	P2Gargoyle        = "f3bd9bfe-847f-4ed1-88aa-6aa6b3fa7199"
	P2Heartmender     = "ab4cf0ba-617f-4ef6-a831-eb3b41b4ab34"
	P2Hancock         = "adf9aed9-dd63-48da-b799-ea94839fc22a"
	P2TwilightShep    = "eff736bc-d311-4bef-9c7e-de4ebe32e46f"
	P2Demonlord       = "049777d0-910d-46d2-9ca1-10277b2fb845"
)

func P2Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}

func P2Activate(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, index int, mana string, targets ...game.TargetRef) error {
	t.Helper()
	FloatForTest(g, p, mana)
	return g.ActivateCatalogAbility(p.ID, id, index, game.ActivateAbilityParams{Targets: targets})
}

func P2Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func P2Keyword(g *game.Game, id uuid.UUID, kw string) bool {
	has := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.LookupCardForEffect(id); ok {
			has = game.HasKeyword(&c, kw)
		}
	})
	return has
}

func P2PT(g *game.Game, id uuid.UUID) (int, int) {
	p, tough := 0, 0
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.LookupCardForEffect(id); ok {
			p, tough = c.PowerForComparison(), c.CurrentToughness()
		}
	})
	return p, tough
}
