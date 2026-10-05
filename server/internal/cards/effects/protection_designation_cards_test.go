package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// protection_designation_cards_test.go — #2181 (Reaver Titan, protection
// from mana value 3 or less). The engine half is in
// game/protection_mana_value_test.go; these assert the card uses it.

const reaverTitanOracle = "2951671b-ac01-4e65-857f-0b7b7c7478ce"

func pdBFCard(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("card %v is not on the battlefield", id)
	return nil
}

func TestReaverTitanIsProtectedFromCheapSources(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	titan := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Reaver Titan",
		TypeLine: "Artifact Creature — Vehicle", OracleID: reaverTitanOracle,
		ManaCost: "{7}", Power: 10, Toughness: 10, Owner: me.ID, Controller: me.ID,
	})
	if got := protectionsOn(t, g, titan); !hasString(got, "mana value 3 or less") {
		t.Fatalf("Reaver Titan's protections = %q", got)
	}
	cheap := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cheap Blocker", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	dear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dear Blocker", TypeLine: "Creature — Bear",
		ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	g.ReadSnapshot(func() {
		if r := g.BlockPairRefusalLocked(pdBFCard(t, g, titan), pdBFCard(t, g, cheap)); r.Reason != game.BlockReasonProtection {
			t.Errorf("a mana value 2 creature blocking Reaver Titan: reason = %q, want protection", r.Reason)
		}
		if r := g.BlockPairRefusalLocked(pdBFCard(t, g, titan), pdBFCard(t, g, dear)); !r.Legal() {
			t.Errorf("a mana value 4 creature may block: %q", r.Reason)
		}
	})

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(cheap, titan, 3); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(dear, titan, 4); err != nil {
			t.Fatal(err)
		}
	})
	if got := pdBFCard(t, g, titan).DamageMarked; got != 4 {
		t.Errorf("Reaver Titan has %d damage, want 4 (the cheap creature's prevented)", got)
	}
}

func TestReaverTitanAttackDealsFiveToEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	titan := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Reaver Titan",
		TypeLine: "Artifact Creature — Vehicle", OracleID: reaverTitanOracle,
		ManaCost: "{7}", Power: 10, Toughness: 10, Owner: me.ID, Controller: me.ID,
	})
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = lifeOf(g, p.ID)
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: titan, Actor: me.ID, Target: g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID})
	})
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := before[p.ID]
		if p.ID != me.ID {
			want -= 5
		}
		if got := lifeOf(g, p.ID); got != want {
			t.Errorf("%s life = %d, want %d", p.Name, got, want)
		}
	}
}
