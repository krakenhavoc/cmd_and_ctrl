package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// ignore_effect_mana_test.go — #6247, CR 116.2d: "Some effects from
// static abilities allow a player to take an action to ignore the effect
// from that ability for a duration. Doing so is a special action. A
// player can take such an action any time they have priority." Leonin
// Arbiter: "Players can't search libraries. Any player may pay {2} for
// that player to ignore this effect until end of turn."

const arbiterOracle = "probe-leonin-arbiter"

func withArbiterCatalog(t *testing.T) {
	t.Helper()
	withCatalogDefs(t, map[string]*CardDef{arbiterOracle: {PlayersCantSearchLibraries: true}})
	prev := CatalogIgnorableEffect
	CatalogIgnorableEffect = func(key string) *IgnorableEffect {
		if key != arbiterOracle {
			return nil
		}
		return &IgnorableEffect{
			Label:    "Pay {2}: ignore Leonin Arbiter's effect until end of turn",
			ManaCost: "{2}",
			Applies:  func(*Game, *Card, uuid.UUID) bool { return true },
		}
	}
	t.Cleanup(func() { CatalogIgnorableEffect = prev })
}

func arbiterTable(t *testing.T) (*Game, *Player, *Player, uuid.UUID) {
	t.Helper()
	withArbiterCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arb := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: arb, Name: "Leonin Arbiter", TypeLine: "Creature — Cat Cleric",
			OracleID: arbiterOracle, Owner: opp.ID, Controller: opp.ID, Power: 2, Toughness: 2})
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: arb, OldZone: ZoneHand, NewZone: ZoneBattlefield})
	})
	return g, me, opp, arb
}

func cantSearch(g *Game, p *Player) bool {
	var out bool
	g.WithWriteLock(func() { out = g.searchCantLocked(p.ID) })
	return out
}

// TestArbiterStopsEveryPlayerSearching — "Players can't search
// libraries": the controller too.
func TestArbiterStopsEveryPlayerSearching(t *testing.T) {
	g, me, opp, _ := arbiterTable(t)
	if !cantSearch(g, me) || !cantSearch(g, opp) {
		t.Fatalf("can't search: me %v, controller %v; want both true", cantSearch(g, me), cantSearch(g, opp))
	}
}

// TestPayingTwoIgnoresTheArbiterForThatPlayerOnly is the issue's
// acceptance: a player pays {2} then searches; another player still
// can't. The offer is a mana cost with no sacrifice, and a second ignore
// isn't offered while the first lasts.
func TestPayingTwoIgnoresTheArbiterForThatPlayerOnly(t *testing.T) {
	g, me, opp, arb := arbiterTable(t)
	var offers []IgnoreEffectOffer
	g.WithWriteLock(func() { offers = g.IgnoreEffectOffersLocked(me.ID) })
	if len(offers) != 1 || offers[0].ManaCost != "{2}" || len(offers[0].Candidates) != 0 {
		t.Fatalf("offers = %+v, want one {2} offer with no sacrifice", offers)
	}
	// A sacrifice named for a mana-cost ignore is refused before anything
	// is paid.
	if err := g.PerformSpecialAction(me.ID, arb, SpecialActionIgnoreEffect,
		SpecialActionParams{SacrificeID: arb}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("with a sacrifice_id: err = %v, want ErrInvalidParam", err)
	}
	// With nothing to pay, a strict payment is refused and nothing changes.
	if err := g.PerformSpecialAction(me.ID, arb, SpecialActionIgnoreEffect,
		SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("an unpaid ignore was taken")
	}
	if !cantSearch(g, me) {
		t.Fatal("a refused ignore switched the effect off")
	}
	fillPool(g, me, 2)
	if err := g.PerformSpecialAction(me.ID, arb, SpecialActionIgnoreEffect,
		SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("paying {2}: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("pool after paying {2}: %d left, want 0", n)
	}
	if cantSearch(g, me) {
		t.Error("the payer still can't search")
	}
	if !cantSearch(g, opp) {
		t.Error("another player is no longer bound")
	}
	g.WithWriteLock(func() { offers = g.IgnoreEffectOffersLocked(me.ID) })
	if len(offers) != 0 {
		t.Errorf("the ignore is offered again while it lasts: %+v", offers)
	}
	// Until end of turn (CR 514.2's sweep).
	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
	if !cantSearch(g, me) {
		t.Error("the ignore outlived the turn")
	}
}
