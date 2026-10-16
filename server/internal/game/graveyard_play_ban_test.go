package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// graveyard_play_ban_test.go — #1987, ADR 0066's amendment of
// 2026-10-16: the engine half of Shaman's Trance. The card itself is
// pinned end to end in cards/effects/shamans_trance_test.go.

// banOwnGraveyard writes the ModCantPlayFromOwnGraveyard record for
// `players`.
func banOwnGraveyard(t *testing.T, g *Game, players ...uuid.UUID) {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.CantPlayFromOwnGraveyardThisTurnForEffect(uuid.Nil, players, "Test Trance")
	})
	if !ok {
		t.Fatal("no ban record was written")
	}
}

// gates asks both gates about one card for one player.
func gates(g *Game, player uuid.UUID, card uuid.UUID, zone ZoneKind) (cast, land error) {
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		if c.IsLand() {
			land = g.LandPlayGateLocked(player, c, zone)
			return
		}
		cast = g.CastGateLocked(player, c, zone, CastSpellParams{})
	})
	return cast, land
}

// The ban reaches the banned player's OWN graveyard, for a spell and
// for a land, and nothing else: not another player's graveyard, not the
// hand, and not a player it does not name.
func TestGraveyardPlayBanRefusesOnlyTheBannedPlayersOwnGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirSpell := seedGraveyard(them, "Their Spell", "Instant", "{R}")
	theirLand := seedGraveyard(them, "Their Land", "Basic Land — Mountain", "")
	mySpell := seedGraveyard(me, "My Spell", "Instant", "{R}")
	banOwnGraveyard(t, g, them.ID)

	cast, _ := gates(g, them.ID, theirSpell, ZoneGraveyard)
	var cant *CantCastError
	if !errors.As(cast, &cant) || !strings.HasPrefix(cant.Reason, graveyardPlayBanReason) {
		t.Errorf("a spell from their own graveyard: %v, want the graveyard ban", cast)
	}
	_, land := gates(g, them.ID, theirLand, ZoneGraveyard)
	if !errors.Is(land, ErrCantPlayLand) {
		t.Errorf("a land from their own graveyard: %v, want ErrCantPlayLand", land)
	}
	// "Their graveyards": a card in MY graveyard is not in theirs.
	if cast, _ := gates(g, them.ID, mySpell, ZoneGraveyard); cast != nil {
		t.Errorf("the ban reached another player's graveyard: %v", cast)
	}
	// Not the hand, and not the zone a card is not in.
	if cast, _ := gates(g, them.ID, theirSpell, ZoneHand); cast != nil {
		t.Errorf("the ban reached the hand: %v", cast)
	}
	// Not a player it does not name.
	if cast, _ := gates(g, me.ID, mySpell, ZoneGraveyard); cast != nil {
		t.Errorf("the ban reached the player it does not name: %v", cast)
	}

	// It ends with the turn (CR 514.2).
	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
	if cast, _ := gates(g, them.ID, theirSpell, ZoneGraveyard); cast != nil {
		t.Errorf("the ban outlived the cleanup sweep: %v", cast)
	}
}

// uuid.Nil names nobody and an unseated player is skipped, so a card
// file that computed an empty list bans nobody rather than everybody.
func TestGraveyardPlayBanNamesNobodyByDefault(t *testing.T) {
	g := newActiveGame(t)
	var ok bool
	g.WithWriteLock(func() {
		ok = g.CantPlayFromOwnGraveyardThisTurnForEffect(uuid.Nil, []uuid.UUID{uuid.Nil, uuid.New()}, "Test Trance")
	})
	if ok || len(g.ScopedEffects) != 0 {
		t.Errorf("a ban naming no seated player wrote a record: %+v", g.ScopedEffects)
	}
}

// Rule 3 gives way to a permission at the printed cost: under a
// standing grant over the graveyard (Yawgmoth's Will's shape), a
// flashback card may be cast for its mana cost as well as flashed back.
// Without a grant the flashback cost is still owed
// (TestGraveyardCastRequiresItsBoundCost).
func TestAPrintedCostPermissionLetsAFlashbackCardBeCastForItsManaCost(t *testing.T) {
	const oracle = "test-1987-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, flashbackCost("{2}{R}")))
	id := looterInGraveyard(t, g, me, oracle)
	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(CastPermission{Player: me.ID, Zone: ZoneGraveyard, Scope: ScopeStanding})
	})

	var offers []string
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(id)
		perm := g.CastPermissionForLocked(me.ID, c, ZoneGraveyard)
		for _, o := range g.CastOffersForLocked(me.ID, c, ZoneGraveyard, perm) {
			if o == nil {
				offers = append(offers, "printed")
				continue
			}
			offers = append(offers, o.Key)
		}
	})
	if strings.Join(offers, ",") != "printed,flashback" {
		t.Errorf("offers = %v, want [printed flashback]", offers)
	}
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("printed-cost cast under the permission: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the printed {R} was not charged: pool %+v", me.ManaPool)
	}
	if item := g.StackMeta[id]; item == nil || item.AltCost != "" {
		t.Errorf("the cast claimed an alternative cost: %+v", item)
	}
}

// A stored standing grant over ANOTHER seat's graveyard is over that
// pile only. The public `exile_play` read (CastPermissionOnCardForEffect)
// must not find it on a card in any other graveyard, the holder's own
// included.
func TestAStoredGrantOverAnotherGraveyardIsOnlyOnThatPile(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirs := seedGraveyard(them, "Their Spell", "Instant", "{R}")
	mine := seedGraveyard(me, "My Spell", "Instant", "{R}")
	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(CastPermission{
			Player: me.ID, Zone: ZoneGraveyard, Scope: ScopeStanding, ZoneOwner: them.ID,
		})
	})
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(theirs)
		if g.CastPermissionOnCardForEffect(c, ZoneGraveyard) == nil {
			t.Error("the grant is not found on a card in the pile it names")
		}
		if g.CastPermissionForLocked(me.ID, c, ZoneGraveyard) == nil {
			t.Error("the holder may not cast a card in the pile the grant names")
		}
		c, _ = g.LookupCardForEffect(mine)
		if p := g.CastPermissionOnCardForEffect(c, ZoneGraveyard); p != nil {
			t.Errorf("the grant is stamped on a card in the holder's own graveyard: %+v", p)
		}
		if g.CastPermissionForLocked(me.ID, c, ZoneGraveyard) != nil {
			t.Error("a grant over another seat's graveyard opened the holder's own")
		}
	})
}
