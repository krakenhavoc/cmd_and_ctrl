package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// undying_persist_pool1_test.go — the first pool PR after #2075 (ADR
// 0113 §4, owner decision 2): creatures whose undying or persist the
// engine already handles, with the rest of their text. One test per
// card; the keyword itself is held by undying_persist_test.go.

const (
	PoolKitchenFinks       = "5470dcfa-4eff-43da-abf7-19922841f719"
	PoolGeralfsMessenger   = "740f9740-aaa7-4061-86f5-85be79742541"
	PoolGeralfsMindcrush   = "e8fa9455-92ac-46e0-bb34-4175c3c66fee"
	PoolEvernightShade     = "9af1b6c4-295f-41d2-b3a0-0863798330d4"
	PoolRestlessApparitn   = "fbd6dda7-ccee-4640-805d-78a9d60f45ef"
	PoolGlenElendra        = "5f3f68b5-8c6a-4181-bb8a-9c73967d198a"
	PoolKithkinSpellduster = "de8beb92-234b-4684-9185-c7077c8a5133"
	PoolAerieOuphes        = "70cdd24a-8dc5-47b5-8729-9ebf486f4821"
	PoolGrazingKelpie      = "a2db1035-008e-4de3-b5e5-2a0de83082b2"
	PoolLesserMasticore    = "a6a257bd-ee19-4246-b997-5288e6fa41e1"
	PoolSightlessGhoul     = "8791adf7-16b5-4692-9eaa-0d0c58867d9d"
	PoolStormboundGeist    = "b5027d79-83e9-437a-baae-03cc3a13a608"
	PoolHowlgeist          = "5a00eb94-6c0f-4f41-8c16-fc4e2f7cb9aa"
	PoolThunderblust       = "30d961f7-aae2-4e5a-8f83-9dc17c7dee47"
	PoolWoodfallPrimus     = "2f70f1bb-29fa-4abb-afc2-653acd0a08b9"
	PoolPyreheartWolf      = "b97ccb69-e76c-4962-91ca-c7fd857140e8"
)

// poolPush puts a catalog creature on the battlefield with no entry.
func PoolPush(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	return B11Push(g, owner, name, "Creature — Test", oracle, power, toughness)
}

func PoolActivate(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana string, targets ...game.TargetRef) {
	t.Helper()
	FloatForTest(g, p, mana)
	if err := g.ActivateCatalogAbility(p.ID, id, 0, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("activate: %v", err)
	}
}

func CardRef(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func PowerOf(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	p := 0
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("no card %s", id)
		}
		p = c.PowerForComparison()
	})
	return p
}
