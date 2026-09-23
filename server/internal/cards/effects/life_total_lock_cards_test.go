package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_total_lock_cards_test.go — #1200's three cards, each proved
// through the engine rather than by reading its own Spec back.
//
// The assertions are about what a PLAYER at the table would notice:
// the drain that drains nobody, the life gain that does not arrive,
// the shockland that enters tapped because you could not pay for it,
// and the board that comes back at the same moment the shield lifts.
// A test that only checked `spec.PlayerLifeTotalLocked` would pass
// with every consumer backed out.

const (
	platinumEmperionOracle = "bcd2f70c-36c2-44b2-9d4b-1000e9bb62b6"
	teferisReproachOracle  = "9381b4a5-a8e0-412d-bc7f-ae15afa0f135"
)

func seatLifeLocked(g *game.Game, p *game.Player) bool {
	var out bool
	g.ReadSnapshot(func() { out = g.PlayerLifeTotalCantChangeLocked(p) })
	return out
}

// --- Platinum Emperion ----------------------------------------------

// TestPlatinumEmperionLocksItsControllersLifeTotal is the card's whole
// printed text, and all three sentences of its parenthetical: no gain,
// no loss, no payment.
func TestPlatinumEmperionLocksItsControllersLifeTotal(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Platinum Emperion", platinumEmperionOracle,
		"Artifact Creature — Golem")

	if !seatLifeLocked(g, me) {
		t.Fatal("the Emperion's controller is not locked")
	}
	if seatLifeLocked(g, opp) {
		t.Error("the Emperion locked a seat that does not control it")
	}

	startMe, startOpp := me.Life, opp.Life
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 6); err != nil {
			t.Fatalf("gain at the Emperion's controller: %v", err)
		}
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -6); err != nil {
			t.Fatalf("loss at the Emperion's controller: %v", err)
		}
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -6); err != nil {
			t.Fatalf("loss at the other seat: %v", err)
		}
	})
	if me.Life != startMe {
		t.Errorf("the Emperion's controller is at %d, want %d", me.Life, startMe)
	}
	if opp.Life != startOpp-6 {
		t.Errorf("the other seat is at %d, want %d — the Emperion is locking the table", opp.Life, startOpp-6)
	}

	// "You can't pay any amount of life except 0."
	var err error
	g.WithWriteLock(func() { err = g.PayLifeForEffect(uuid.Nil, me.ID, 1) })
	if err == nil {
		t.Error("the Emperion's controller paid 1 life")
	}
	g.WithWriteLock(func() { err = g.PayLifeForEffect(uuid.Nil, me.ID, 0) })
	if err != nil {
		t.Errorf("paying 0 life under an Emperion: %v", err)
	}
}

// TestPlatinumEmperionLeavingUnlocksTheTotal — the grant is derived,
// so there is nothing to unwind and nothing to strand.
func TestPlatinumEmperionLeavingUnlocksTheTotal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	emperion := pushPermanentForTest(g, me.ID, "Platinum Emperion", platinumEmperionOracle,
		"Artifact Creature — Golem")
	if !seatLifeLocked(g, me) {
		t.Fatal("no lock with the Emperion out")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(emperion); err != nil {
			t.Fatalf("destroy the Emperion: %v", err)
		}
	})
	if seatLifeLocked(g, me) {
		t.Error("the lock outlived the Emperion")
	}
}

// TestPlatinumEmperionDoesNotStopDamageOrPoison — the two things the
// card is regularly misread as doing. Damage is still dealt (the life
// loss is what does not happen), and poison is a counter, not life.
func TestPlatinumEmperionDoesNotStopDamageOrPoison(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Platinum Emperion", platinumEmperionOracle,
		"Artifact Creature — Golem")
	source := pushPermanentForTest(g, g.Seats[1].ID, "Pinger", "test-life-lock-card-pinger", "Artifact")

	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, me.ID, 5); err != nil {
			t.Fatalf("damage at the Emperion's controller: %v", err)
		}
		if err := g.AddPlayerCounterForEffect(me.ID, game.CounterPoison, 2); err != nil {
			t.Fatalf("poison at the Emperion's controller: %v", err)
		}
	})
	if me.Life != life {
		t.Errorf("the Emperion's controller lost %d life to damage", life-me.Life)
	}
	if got := me.Counters[game.CounterPoison]; got != 2 {
		t.Errorf("poison counters = %d, want 2 — the Emperion is not a counter prohibition", got)
	}
}

// --- Teferi's Protection --------------------------------------------

// TestTeferisProtectionLocksYourLifeTotalUntilYourNextTurn is the
// clause the card carried a caveat for from #1197 until #1200, with
// the boundary: through three opponents' turns and gone as yours
// begins, at the same moment the board phases back in.
func TestTeferisProtectionLocksYourLifeTotalUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushPermanentForTest(g, me.ID, "Bear", "test-life-lock-teferi-bear", "Creature — Bear")

	castCatalogSpell(t, g, "Teferi's Protection", "Instant", teferisProtectionOracle, nil)
	passPriorityAroundTable(t, g)

	if !seatLifeLocked(g, me) {
		t.Fatal("Teferi's Protection did not lock the caster's life total")
	}
	if !phasedOutInCatalogGame(g, mine) {
		t.Fatal("the board did not phase out; the two clauses have to ship together")
	}

	// The half the protection never covered: a drain is not damage.
	start := me.Life
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -9); err != nil {
			t.Fatalf("drain at the protected player: %v", err)
		}
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 4); err != nil {
			t.Fatalf("gain at the protected player: %v", err)
		}
	})
	if me.Life != start {
		t.Errorf("life moved by %d under Teferi's Protection", me.Life-start)
	}

	for i := 0; i < 3; i++ {
		advanceOneTurnForTest(t, g)
		if !seatLifeLocked(g, me) {
			t.Fatalf("the lock ended after %d opponent turns; it lasts until YOUR next turn", i+1)
		}
	}
	advanceOneTurnForTest(t, g)
	if seatLifeLocked(g, me) {
		t.Error("the lock survived the beginning of its own player's next turn")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the board did not phase back in at the same untap step")
	}
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3); err != nil {
			t.Fatalf("drain after the lock expired: %v", err)
		}
	})
	if me.Life != start-3 {
		t.Errorf("life = %d after the lock expired, want %d", me.Life, start-3)
	}
}

// TestTeferisProtectionHasNoCaveatsLeft — the ADR 0037 coverage
// signal for the card this seam finishes. Three printed clauses,
// three ADRs, nothing declared weaker than printed.
func TestTeferisProtectionHasNoCaveatsLeft(t *testing.T) {
	spec, ok := Lookup(teferisProtectionOracle)
	if !ok {
		t.Fatal("Teferi's Protection is in the catalog")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Teferi's Protection is %v with %d caveats, want full with none",
			spec.Completeness, len(spec.Caveats))
	}
}

// --- Teferi's Reproach ----------------------------------------------

// TestTeferisReproachLocksAndShieldsTheTARGET is the whole reason
// this card is in the batch: every clause is the same machinery aimed
// at a seat that is not the caster, and the duration is counted on
// the TARGET's seat-turn counter.
func TestTeferisReproachLocksAndShieldsTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirCreature := pushPermanentForTest(g, them.ID, "Their Bear",
		"test-life-lock-reproach-bear", "Creature — Bear")
	theirLand := pushPermanentForTest(g, them.ID, "Their Forest",
		"test-life-lock-reproach-forest", "Basic Land — Forest")

	spell := castCatalogSpell(t, g, "Teferi's Reproach", "Instant", teferisReproachOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}})
	passPriorityAroundTable(t, g)

	if !seatLifeLocked(g, them) {
		t.Fatal("the target's life total is not locked")
	}
	if seatLifeLocked(g, me) {
		t.Error("the CASTER's life total was locked; the card names the target")
	}
	if got := playerAbilities(g, them); !hasPlayerAbility(got, ProtectionFromEverything) {
		t.Errorf("the target has %v, want protection from everything", got)
	}
	if !phasedOutInCatalogGame(g, theirCreature) {
		t.Error("the target's nonland permanent did not phase out")
	}
	if phasedOutInCatalogGame(g, theirLand) {
		t.Error("the target's LAND phased out; the card says nonland")
	}

	// "Exile Teferi's Reproach", not the graveyard.
	var inExile bool
	g.ReadSnapshot(func() { inExile = g.Exile.Contains(spell) })
	if !inExile {
		t.Error("Teferi's Reproach did not exile itself")
	}

	start := them.Life
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, them.ID, -5); err != nil {
			t.Fatalf("drain at the reproached player: %v", err)
		}
	})
	if them.Life != start {
		t.Errorf("the reproached player lost %d life", start-them.Life)
	}
}

// TestTeferisReproachEndsOnTheTargetsNextTurn — the duration is
// stamped against the TARGET's seat-turn counter, not the caster's,
// which is the one thing a copy-paste from Teferi's Protection would
// get wrong. Seat 1 is reproached on seat 0's turn, so the window is
// one seat-turn long rather than a full rotation.
func TestTeferisReproachEndsOnTheTargetsNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	them := g.Seats[1]
	castCatalogSpell(t, g, "Teferi's Reproach", "Instant", teferisReproachOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}})
	passPriorityAroundTable(t, g)
	if !seatLifeLocked(g, them) {
		t.Fatal("the target's life total is not locked")
	}

	advanceOneTurnForTest(t, g) // seat 1's own turn begins
	if seatLifeLocked(g, them) {
		t.Error("the lock survived the beginning of the TARGET's next turn")
	}
	if hasPlayerAbility(playerAbilities(g, them), ProtectionFromEverything) {
		t.Error("the protection survived the beginning of the TARGET's next turn")
	}
}
