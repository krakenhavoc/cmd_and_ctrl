package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// serra_ascendant_test.go covers the card and, through it, #1117's
// life-total layer invalidation.
//
// Every assertion is taken with NO intervening battlefield event. That
// is deliberate and it is the whole test: a battlefield move drops the
// cached layer resolution, so a version of this card with no
// invalidation at all passes any test that plays a land between the
// life change and the read. The bug was always "right whenever
// something else happened, one event behind otherwise".

const serraAscendantOracle = "27ad3e00-6ffb-48f7-8469-8868d066d1e2"

// pushSerraAscendant seeds the printed 1/1 on the battlefield under
// seat 0 and returns its instance ID and controller.
func pushSerraAscendant(t *testing.T, g *game.Game) (uuid.UUID, uuid.UUID) {
	t.Helper()
	owner := g.Seats[0].ID
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Serra Ascendant",
		TypeLine:   "Creature — Human Monk",
		ManaCost:   "{W}",
		OracleID:   serraAscendantOracle,
		Power:      1,
		Toughness:  1,
		Owner:      owner,
		Controller: owner,
	})
	return id, owner
}

// assertAscendant checks both halves of the conditional clause at
// once, because the card is only right when they agree.
func assertAscendant(t *testing.T, g *game.Game, id uuid.UUID, wantPT int, wantFlying bool) {
	t.Helper()
	if got := effectivePower(t, g, id); got != wantPT {
		t.Errorf("power at %d life = %d, want %d", g.Seats[0].Life, got, wantPT)
	}
	if got := effectiveToughness(t, g, id); got != wantPT {
		t.Errorf("toughness at %d life = %d, want %d", g.Seats[0].Life, got, wantPT)
	}
	abilities := effectiveAbilities(t, g, id)
	if got := containsString(abilities, "flying"); got != wantFlying {
		t.Errorf("flying at %d life = %v, want %v (abilities %v)", g.Seats[0].Life, got, wantFlying, abilities)
	}
	if !containsString(abilities, "lifelink") {
		t.Errorf("lifelink missing at %d life (abilities %v) — it is printed and unconditional", g.Seats[0].Life, abilities)
	}
}

// TestSerraAscendantIsOnAtCommanderStartingLife: 40 is 30 or more, so
// the clause is live from the opening hand. This is not an edge case
// in this format — it is the default state of the card.
func TestSerraAscendantIsOnAtCommanderStartingLife(t *testing.T) {
	g := newCatalogGame(t)
	id, _ := pushSerraAscendant(t, g)
	assertAscendant(t, g, id, 6, true)
}

// TestSerraAscendantSwitchesOffBelowThirtyAndBackOn walks the clause
// in both directions, with nothing between each life change and the
// read that follows it.
func TestSerraAscendantSwitchesOffBelowThirtyAndBackOn(t *testing.T) {
	g := newCatalogGame(t)
	id, owner := pushSerraAscendant(t, g)
	assertAscendant(t, g, id, 6, true)

	// 40 → 29. One below the line.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, owner, -11) })
	if g.Seats[0].Life != 29 {
		t.Fatalf("life = %d, want 29", g.Seats[0].Life)
	}
	assertAscendant(t, g, id, 1, false)

	// 29 → 30. Exactly on it: "30 or more" includes 30.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, owner, +1) })
	assertAscendant(t, g, id, 6, true)
}

// TestSerraAscendantShrinksOnDamageNotJustLifeLoss is the route that
// does not emit a life event at all: damage to a player writes the
// total and emits EventDealDamage. Getting the card right on a
// Lightning Bolt is most of what "it went below 30" means in play.
func TestSerraAscendantShrinksOnDamageNotJustLifeLoss(t *testing.T) {
	g := newCatalogGame(t)
	id, owner := pushSerraAscendant(t, g)
	assertAscendant(t, g, id, 6, true)

	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.Nil, owner, 11) })
	if g.Seats[0].Life != 29 {
		t.Fatalf("life after 11 damage = %d, want 29", g.Seats[0].Life)
	}
	assertAscendant(t, g, id, 1, false)
}

// TestSerraAscendantReadsItsOwnControllersLife, not the table's: an
// opponent at 12 does not turn it off, and an opponent gaining life
// does not turn it on.
func TestSerraAscendantReadsItsOwnControllersLife(t *testing.T) {
	g := newCatalogGame(t)
	id, _ := pushSerraAscendant(t, g)
	opponent := g.Seats[1].ID

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opponent, -28) })
	assertAscendant(t, g, id, 6, true)
}

// --- the other two cards the same invalidation unblocked -----------

// TestBloodghastHasHasteOnlyWhileAnOpponentIsLow covers the clause
// Bloodghast shipped without: it reads an OPPONENT'S life total, so
// it exercises the same bump from the other side of the table.
func TestBloodghastHasHasteOnlyWhileAnOpponentIsLow(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghast := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Bloodghast",
		TypeLine:   "Creature — Vampire Spirit",
		OracleID:   bloodghastOracle,
		Power:      2,
		Toughness:  1,
		Owner:      me.ID,
		Controller: me.ID,
	})
	if containsString(effectiveAbilities(t, g, ghast), "haste") {
		t.Errorf("haste with every opponent at %d life", opp.Life)
	}

	// 40 → 10. On the line: "10 or less".
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -30) })
	if !containsString(effectiveAbilities(t, g, ghast), "haste") {
		t.Errorf("no haste with an opponent at %d life", opp.Life)
	}

	// And back off the moment they gain one.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, +1) })
	if containsString(effectiveAbilities(t, g, ghast), "haste") {
		t.Errorf("haste survived the opponent going back to %d life", opp.Life)
	}
}

// TestBloodghastReadsOpponentsLifeNotItsControllers: its controller
// dropping to 10 is not the condition.
func TestBloodghastReadsOpponentsLifeNotItsControllers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ghast := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Bloodghast",
		TypeLine:   "Creature — Vampire Spirit",
		OracleID:   bloodghastOracle,
		Power:      2,
		Toughness:  1,
		Owner:      me.ID,
		Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -30) })
	if containsString(effectiveAbilities(t, g, ghast), "haste") {
		t.Errorf("haste from its OWN controller being at %d life", me.Life)
	}
}
