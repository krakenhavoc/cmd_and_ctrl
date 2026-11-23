package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr1_g2_cards_test.go — ADR 0108 Delivery PR 1 (#1886, #1887),
// card group 2: each card against the board that tells its wording
// apart.

const (
	P1g2CombustionOracle = "522a04ff-cbfe-47b0-bd29-ef6fc27a6905"
	P1g2BlitzOracle      = "d2db3f9c-26fe-487b-bbb7-8bc6f33456f1"
	P1g2NarsetOracle     = "247075c5-62f8-41a1-91b3-562ff0aabb30"
	P1g2AccursedOracle   = "b5aae42b-3fde-4f10-b85e-882c528badef"
	P1g2SmiteOracle      = "37994591-3494-4314-a7da-49c112b0866f"
	P1g2WithinOracle     = "8a1ccbdb-3d89-42fb-a731-6db4241acf24"
	P1g2BargainOracle    = "bfb3d862-d92d-4a4f-9a22-671b55d954fd"
	P1g2WiltOracle       = "3bed0b60-5944-44bb-9ddd-c82f323e6d20"
	P1g2BouncerOracle    = "36ac8fc6-98ad-499b-9adf-046433e2c122"
	P1g2ZenithOracle     = "82e61db8-4625-488f-8a5f-66ace9bbf34a"
	P1g2LavaOracle       = "e42fb51e-254a-43ac-ad02-9f0fad0f4c8a"
	P1g2WoundOracle      = "a30159ae-f6a6-4e29-bca2-769d3657d310"
	P1g2TorchOracle      = "10d33e95-3e5d-447e-ba4a-acd3c33b4045"
	P1g2SpikefieldOracle = "81036c9f-fe0a-45a7-bcd5-0d344f31055a"
)

// p1g2Cast casts a catalog spell from the active player's hand with the
// whole announcement in `params` (targets, optional costs, cost branch,
// sacrifices, X).
func P1g2Cast(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// p1g2Seats is the active player and the next one.
func P1g2Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func P1g2Indestructible(g *game.Game, owner uuid.UUID, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Darksteel Bear", TypeLine: "Artifact Creature — Bear",
		Power: power, Toughness: toughness, Keywords: []string{"indestructible"}, Owner: owner, Controller: owner,
	})
}

func P1g2Planeswalker(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: loyalty}, Owner: owner, Controller: owner,
	})
}

func P1g2Graveyard(p *game.Player, typeLine string, n int) {
	for i := 0; i < n; i++ {
		p.Graveyard.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Old Card", TypeLine: typeLine, Power: 1, Toughness: 1,
			Owner: p.ID, Controller: p.ID,
		})
	}
}

func P1g2WantDamage(t *testing.T, g *game.Game, id uuid.UUID, want int, what string) {
	t.Helper()
	if got := DamageMarkedOn(g, id); got != want {
		t.Errorf("%s has %d damage, want %d", what, got, want)
	}
}

func P1g2HasKeyword(g *game.Game, id uuid.UUID, kw string) bool {
	var has bool
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			has = game.HasKeyword(&c, kw)
		}
	})
	return has
}
