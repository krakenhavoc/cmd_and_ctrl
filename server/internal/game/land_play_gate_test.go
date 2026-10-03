package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// land_play_gate_test.go — ADR 0109 §4 (#1895): "players can't play
// lands" (CR 101.2: "can't" beats "can") asked at the land branch of the
// play path, before the drop count, and at the land play a resolution
// instructs (CR 305.2a). The enumerator's and the view's halves are in
// internal/legal and internal/protocol, against the real cards.

const landBanOracle = "test-land-ban"

// stubLandPlayRestrictions wires CatalogLandPlayRestrictions for one
// oracle ID and restores the previous hook at the end of the test.
func stubLandPlayRestrictions(t *testing.T, oracleID string, rules []LandPlayRestriction) {
	t.Helper()
	prev := CatalogLandPlayRestrictions
	CatalogLandPlayRestrictions = func(id string) []LandPlayRestriction {
		if id == oracleID {
			return rules
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { CatalogLandPlayRestrictions = prev })
}

// seedLandBan puts a permanent on the battlefield whose static forbids
// every land play by everyone.
func seedLandBan(t *testing.T, g *Game, controller uuid.UUID) uuid.UUID {
	t.Helper()
	stubLandPlayRestrictions(t, landBanOracle, []LandPlayRestriction{{
		Label:   "Players can't play lands.",
		Forbids: func(LandPlayQuery) bool { return true },
	}})
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Test Dispute",
		TypeLine:   "Enchantment",
		OracleID:   landBanOracle,
		Owner:      controller,
		Controller: controller,
	})
	return id
}

func assertCantPlayLand(t *testing.T, err error, wantReason string) {
	t.Helper()
	if err == nil {
		t.Fatalf("the land play was allowed, want a refusal %q", wantReason)
	}
	if !errors.Is(err, ErrCantPlayLand) {
		t.Fatalf("refusal is %v, want it to wrap ErrCantPlayLand", err)
	}
	if errors.Is(err, ErrLandDropUnavailable) {
		t.Fatalf("refusal is the drop-count one (%v); the gate must be asked first", err)
	}
	var cant *CantPlayLandError
	if !errors.As(err, &cant) || cant.Reason != wantReason {
		t.Fatalf("refusal = %v, want a *CantPlayLandError with reason %q", err, wantReason)
	}
}

// A static on the battlefield refuses the play BEFORE the drop count,
// leaves the card in hand, and an extra drop does not lift it (CR 101.2's
// own example).
func TestLandPlayStaticRefusesBeforeTheDropCount(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	seedLandBan(t, g, active.ID)
	g.WithWriteLock(func() { g.GrantAdditionalLandPlayForEffect(active.ID, 2) })

	handBefore := len(active.Hand.Cards)
	err := dropLandFromHand(t, g, active)
	assertCantPlayLand(t, err, "Players can't play lands. — Test Dispute")
	if len(active.Hand.Cards) != handBefore {
		t.Errorf("a refused land play moved the card: hand %d, want %d", len(active.Hand.Cards), handBefore)
	}
	if g.LandsPlayedThisTurnFor(active.ID) != 0 {
		t.Errorf("a refused land play spent a land drop")
	}
}

// Nothing is stored for a static: the source leaving lifts the ban on the
// very next query.
func TestLandPlayStaticStopsWhenItsSourceLeaves(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	src := seedLandBan(t, g, active.ID)
	if err := dropLandFromHand(t, g, active); !errors.Is(err, ErrCantPlayLand) {
		t.Fatalf("under the ban: %v", err)
	}
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(src) })
	if err := dropLandFromHand(t, g, active); err != nil {
		t.Fatalf("the ban outlived its source: %v", err)
	}
}

// The "this turn" form is a stored ScopedEffect for one player, swept at
// cleanup (CR 514.2).
func TestCantPlayLandsThisTurnRecord(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[1-g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		if g.CantPlayLandsThisTurnForEffect(uuid.Nil, uuid.Nil, "x") {
			t.Error("a record was written for no player")
		}
		if !g.CantPlayLandsThisTurnForEffect(uuid.Nil, active.ID, "Turf Wound") {
			t.Fatal("no record written")
		}
	})
	assertCantPlayLand(t, dropLandFromHand(t, g, active), "You can't play lands this turn")

	var forActive, forOther error
	g.WithWriteLock(func() {
		land := Card{TypeLine: "Basic Land — Mountain"}
		forActive = g.LandPlayGateLocked(active.ID, land, ZoneHand)
		forOther = g.LandPlayGateLocked(other.ID, land, ZoneHand)
		if got := g.LandPlayBanFor(active.ID); got != "You can't play lands this turn" {
			t.Errorf("the seat's banner = %q", got)
		}
		if got := g.LandPlayBanFor(other.ID); got != "" {
			t.Errorf("the other seat's banner = %q, want none", got)
		}
	})
	if forActive == nil || forOther != nil {
		t.Errorf("gate: banned player %v, other player %v", forActive, forOther)
	}

	g.WithWriteLock(func() { g.sweepScopedEffectsLocked(true) })
	if err := dropLandFromHand(t, g, active); err != nil {
		t.Fatalf("the ban survived the end of the turn: %v", err)
	}
}

// CR 305.2a: a land played during a resolution still counts as a land
// play, so the ban stops it too, and nothing moves.
func TestLandPlayDuringAResolutionObeysTheGate(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	_, land := hideawaySetup(t, g, Card{Name: "Hidden Mountain", TypeLine: "Basic Land — Mountain"})
	seedLandBan(t, g, me.ID)
	g.WithWriteLock(func() {
		if g.CanPlayLandDuringResolutionForEffect(me.ID, land) {
			t.Error("the resolution-time play is offered under a land ban")
		}
		if played, err := g.PlayLandDuringResolutionForEffect(me.ID, land); played || err != nil {
			t.Errorf("played=%v err=%v, want (false, nil)", played, err)
		}
	})
	if g.Battlefield.Contains(land) || g.LandsPlayedThisTurnFor(me.ID) != 0 {
		t.Error("a land entered under a land ban")
	}
}

// A ban that names a zone binds only plays from it; one that names a
// player binds only that player.
func TestLandPlayRestrictionReadsTheQuery(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	stubLandPlayRestrictions(t, landBanOracle, []LandPlayRestriction{{
		Label: "Your opponents can't play land cards from graveyards.",
		Forbids: func(q LandPlayQuery) bool {
			return q.Source.Controller != q.Player && q.FromZone == ZoneGraveyard && q.Card.Name == "Mountain"
		},
	}})
	g.Battlefield.PushTop(Card{InstanceID: uuid.New(), Name: "Test Tomik", TypeLine: "Creature", OracleID: landBanOracle,
		Owner: active.ID, Controller: active.ID})
	opp := g.Seats[1-g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		mountain := Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"}
		if g.LandPlayGateLocked(opp.ID, mountain, ZoneGraveyard) == nil {
			t.Error("the opponent may play the land from a graveyard")
		}
		if err := g.LandPlayGateLocked(opp.ID, mountain, ZoneHand); err != nil {
			t.Errorf("the opponent is refused a hand play: %v", err)
		}
		if err := g.LandPlayGateLocked(active.ID, mountain, ZoneGraveyard); err != nil {
			t.Errorf("the controller is refused: %v", err)
		}
		if err := g.LandPlayGateLocked(opp.ID, Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, ZoneGraveyard); err != nil {
			t.Errorf("a differently named land is refused: %v", err)
		}
	})
}

// CR 206.3a's list folds the accents and the apostrophe the rules and a
// card database spell differently.
func TestArabianNightsNames(t *testing.T) {
	for _, name := range []string{"Desert", "Library of Alexandria", "Ifh-Bíff Efreet", "Ring of Ma'rûf",
		"Aladdin’s Lamp", "Aladdin's Lamp", "Dandân", "Juzám Djinn", "City in a Bottle"} {
		if !IsArabianNightsName(name) {
			t.Errorf("%q is not recognised as an Arabian Nights name", name)
		}
	}
	for _, name := range []string{"Forest", "Mountain", "Desert Warrior", "Serra Angel", ""} {
		if IsArabianNightsName(name) {
			t.Errorf("%q is wrongly an Arabian Nights name", name)
		}
	}
	if n := len(ArabianNightsNames); n != 77 {
		t.Errorf("CR 206.3a lists 77 names, the table has %d", n)
	}
}
