package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S58 PR 6: Lercron-Xenagos, red-green combat (#2061).

const (
	S58p6AnzragOracle      = "4adcd967-9ff7-4940-b8b9-0c4215bbcb75"
	S58p6BerserkOracle     = "8b67d192-9a05-4a47-82ae-5fc4b7834d88"
	S58p6GwennaOracle      = "7d95e72a-461e-4993-a190-847466a4b17c"
	S58p6LastMarchOracle   = "3999cd64-ff5c-4e2c-8aa6-d14f9f8a2b4c"
	S58p6MageSlayerOracle  = "db444262-c40f-4c80-9d14-265c32f77cf1"
	S58p6RavineOracle      = "8d38194e-b607-4ff4-9c19-0e8636d463bf"
	S58p6SarythOracle      = "4e07a0b3-f340-4ed1-a8a9-c25fe6f37fb3"
	S58p6ShadowOracle      = "4d81db0d-2f56-42ac-92ce-249a34d262d4"
	S58p6SomberwaldOracle  = "f249db15-09fa-4a1c-9461-9ae803c219b6"
	S58p6SkullsporeOracle  = "ba26ff0a-e714-44f2-95cf-1a5a6088edf9"
	S58p6TraverseOracle    = "0d49cf51-af1f-4c17-9cf4-b82bc7c2c72e"
	S58p6UnleashFuryOracle = "4eaa100c-f1ef-4a2b-9370-5dad7f3a95f2"
	S58p6XenagosOracle     = "cb15a8dd-57fe-466f-847e-66476b690a1f"
	S58p6ZopandrelOracle   = "b168e4e6-f572-4d9f-b98f-95b2611354cb"
	S58p6GrowingRiteOracle = "ea9c459a-6047-43aa-968f-a582be4000e8"
	S58p6DamnationOracle   = "d57a8f0b-7989-4db5-8756-6f2690097252"
)

// s58p6View is the battlefield card after a layer recompute.
func S58p6View(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c := FindBattlefieldCardByID(g, id); c != nil {
			out, found = *c, true
		}
	})
	if !found {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return out
}

// s58p6Push seeds a catalog permanent with a mana cost and a body, not
// summoning sick.
func S58p6Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, cost string, p, tough int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

// s58p6CastNow casts a card from p's hand at the current step, with no
// advancing, so the cast window is whatever step the table is in.
func S58p6CastNow(g *game.Game, p *game.Player, name, typeLine, oracle string, targets []game.TargetRef) error {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Targets: targets})
}

func S58p6Activate(t *testing.T, g *game.Game, who uuid.UUID, card uuid.UUID, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(who, card, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
}
