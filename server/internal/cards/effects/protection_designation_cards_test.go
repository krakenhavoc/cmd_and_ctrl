package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// protection_designation_cards_test.go — #2181 (Reaver Titan, protection
// from mana value 3 or less) and #2145 (Lord of the Nazgul, protection
// from Ring-bearers). The engine half is in
// game/protection_mana_value_test.go; these assert the cards use it.

const (
	reaverTitanOracle    = "2951671b-ac01-4e65-857f-0b7b7c7478ce"
	lordOfTheNazgulOracl = "2e94bfb2-9f7d-43af-8995-6cd5ffb15b21"
)

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

func countWraiths(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.HasSubtype("Wraith") {
			n++
		}
	}
	return n
}

func TestLordOfTheNazgulGrantsProtectionFromRingBearersToWraiths(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	lord := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lord of the Nazgûl",
		TypeLine: "Legendary Creature — Wraith Noble", OracleID: lordOfTheNazgulOracl,
		Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	oppWraith := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Wraith", TypeLine: "Creature — Wraith",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	if !hasString(protectionsOn(t, g, lord), "Ring-bearers") {
		t.Errorf("the Lord is a Wraith you control: %q", protectionsOn(t, g, lord))
	}
	if len(protectionsOn(t, g, bear)) != 0 || len(protectionsOn(t, g, oppWraith)) != 0 {
		t.Error("only Wraiths YOU control have it")
	}

	// A Ring-bearer opponent's creature can't block the Lord.
	bearer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bearer", TypeLine: "Creature — Hobbit",
		Keywords: []string{"flying"}, Power: 1, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID, RingBearer: true,
	})
	g.ReadSnapshot(func() {
		if r := g.BlockPairRefusalLocked(pdBFCard(t, g, lord), pdBFCard(t, g, bearer)); r.Reason != game.BlockReasonProtection {
			t.Errorf("a Ring-bearer blocking the Lord: reason = %q, want protection", r.Reason)
		}
	})
}

func TestLordOfTheNazgulCastTriggerMakesAWraithAndNineMakeThemNineNine(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lord := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lord of the Nazgûl",
		TypeLine: "Legendary Creature — Wraith Noble", OracleID: lordOfTheNazgulOracl,
		Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	for i := 0; i < 6; i++ { // Lord + 6 = 7; the token makes 8, one short
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Wraith", TypeLine: "Creature — Wraith",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	}
	castCatalogSpell(t, g, "Test Instant", "Instant", "test-lord-instant", nil)
	passPriorityAroundTable(t, g)
	if got := countWraiths(g, me.ID); got != 8 {
		t.Fatalf("Wraiths = %d, want 8 (the trigger made one)", got)
	}
	if got := effectivePower(t, g, lord); got != 4 {
		t.Errorf("eight Wraiths: Lord power = %d, want 4 (no 9/9)", got)
	}

	// One more Wraith, one more cast: the token is the ninth.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wraith", TypeLine: "Creature — Wraith",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Test Instant 2", "Instant", "test-lord-instant", nil)
	passPriorityAroundTable(t, g)
	if got := countWraiths(g, me.ID); got != 10 {
		t.Fatalf("Wraiths = %d, want 10", got)
	}
	if p, tough := effectivePower(t, g, lord), effectiveToughness(t, g, lord); p != 9 || tough != 9 {
		t.Errorf("Lord = %d/%d, want 9/9 with nine or more Wraiths", p, tough)
	}
}
