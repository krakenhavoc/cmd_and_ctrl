package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// divided_cards_test.go — #1563's new catalog cards: one per owner of a
// divided clause the engine now honours — a trigger with a bounded
// count (Inferno Titan), a trigger with "any number" (Bogardan
// Hellkite), a spell whose amount is an X paid in life (Fire
// Covenant) and an activated ability (Mogg Mob).

const (
	infernoTitanOracle     = "0ce47c8b-1e1f-463f-94f0-35ca00be89e6"
	bogardanHellkiteOracle = "699bc130-2f1f-4bc8-a25f-9329e40efbb1"
	fireCovenantOracle     = "025939a0-424a-41bf-8fc9-2ef9ea5485f7"
	moggMobOracle          = "fb7171f8-60a0-4309-93ff-113e1f62113c"
)

func TestInfernoTitanDividesThreeOnEntryAndOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	castCatalogSpell(t, g, "Inferno Titan", "Creature — Giant", infernoTitanOracle, nil)
	passPriorityAroundTable(t, g)
	// Enters: two targets, 2 to the second — a split, not an even one.
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 2}, a, b)
	passPriorityAroundTable(t, g)
	if d1, d2 := e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked; d1 != 1 || d2 != 2 {
		t.Fatalf("enters: %d/%d, want 1/2", d1, d2)
	}

	// Attacks: the same trigger, this time with a player among the
	// three targets. A second Titan that has been around since before
	// the turn does the attacking.
	attacker := pushVanillaCreature(g, me.ID, "Inferno Titan", 6, 6)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].OracleID = infernoTitanOracle
			}
		}
	})
	life := opp.Life
	declareAttack(t, g, opp.ID, attacker)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the attack trigger asks for targets")
	}
	refs := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: a}}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{opp.ID: 2, a: 1}); err != nil {
		t.Fatalf("attack trigger: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the player's share: life %d → %d, want -2", life, opp.Life)
	}
	if d := e2Card(t, g, a).DamageMarked; d != 2 {
		t.Errorf("the creature's share: %d total marked, want 1+1", d)
	}
}

func TestInfernoTitanFirebreathing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	titan := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Inferno Titan", TypeLine: "Creature — Giant",
		OracleID: infernoTitanOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, titan, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, titan); p != 7 {
		t.Errorf("{R}: +1/+0 — power %d, want 7", p)
	}
}

// "Any number of targets": the clause is unbounded, the five points
// are not — a sixth target is refused because it would get 0.
func TestBogardanHellkiteDividesFiveAmongAnyNumber(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	for _, n := range []string{"A", "B", "C", "D", "E", "F"} {
		ids = append(ids, b12Creature(g, opp.ID, n, "Creature — Wall", 0, 30))
	}
	castCatalogSpell(t, g, "Bogardan Hellkite", "Creature — Dragon", bogardanHellkiteOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for targets")
	}
	six := map[uuid.UUID]int{}
	for _, id := range ids {
		six[id] = 1
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, cardRefs(ids...), six); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("six targets for five points: %v, want ErrInvalidParam", err)
	}
	life := opp.Life
	refs := append(cardRefs(ids[0]), game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{ids[0]: 1, opp.ID: 4}); err != nil {
		t.Fatalf("a legal division: %v", err)
	}
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, ids[0]).DamageMarked; d != 1 || opp.Life != life-4 {
		t.Errorf("1 to the creature, 4 to the player: %d marked, life %d → %d", d, life, opp.Life)
	}
}

func TestFireCovenantDividesTheLifePaid(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Ox", 3, 3)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	b12PlayFromHand(t, g, "Fire Covenant", "Instant", fireCovenantOracle,
		game.CastSpellParams{XValue: 5, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 3}})
	if me.Life != life-5 {
		t.Errorf("X=5 is paid in life at cast: %d → %d", life, me.Life)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("2 to the 2/2 and 3 to the 3/3 kill both")
	}
}

func TestMoggMobDividesAsItIsActivated(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	mob := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mogg Mob", TypeLine: "Creature — Goblin",
		OracleID: moggMobOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	refs := append(cardRefs(a), game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if err := g.ActivateCatalogAbility(me.ID, mob, 0, game.ActivateAbilityParams{
		Targets: refs, Distribution: map[uuid.UUID]int{a: 2, opp.ID: 2},
	}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a division of 4: %v, want ErrInvalidParam", err)
	}
	if !g.Battlefield.Contains(mob) {
		t.Fatal("a refused activation paid its sacrifice")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, mob, 0, game.ActivateAbilityParams{
		Targets: refs, Distribution: map[uuid.UUID]int{a: 2, opp.ID: 1},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, a).DamageMarked; d != 2 || opp.Life != life-1 {
		t.Errorf("2 to the creature, 1 to the player: %d marked, life %d → %d", d, life, opp.Life)
	}
}
