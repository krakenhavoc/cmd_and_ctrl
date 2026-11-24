package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// exile_links_6092_test.go — #6092's two engine seams, against stubbed
// catalog hooks; the cards are pinned end to end in the effects tests
// (Coram, the Undertaker; Shelob, Dread Weaver).

// CastPermission.AllGraveyards: a standing graveyard permission that
// reaches every player's graveyard, not only its holder's. Without the
// flag the same permission reaches only the holder's own pile.
func TestStandingPermissionAllGraveyardsReachesEveryGraveyard(t *testing.T) {
	for _, all := range []bool{false, true} {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		withCatalogCastPermissions(t, func(id string) []CastPermission {
			if id != "test-coram" {
				return nil
			}
			return []CastPermission{{
				Zone: ZoneGraveyard, Scope: ScopeStanding, Duration: WhileInZoneDuration(),
				AllGraveyards: all,
			}}
		})
		source := permanentFor(g, me, "Test Coram", "Creature", "{B}")
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == source {
					g.Battlefield.Cards[i].OracleID = "test-coram"
				}
			}
		})
		mine := seedGraveyard(me, "My Instant", "Instant", "{U}")
		theirs := seedGraveyard(opp, "Their Instant", "Instant", "{U}")
		if perm := grantOn(g, me.ID, mine, ZoneGraveyard); !perm.Granted() {
			t.Errorf("all=%v: the holder's own graveyard is not reached", all)
		}
		if got := grantOn(g, me.ID, theirs, ZoneGraveyard).Granted(); got != all {
			t.Errorf("all=%v: another player's graveyard reached = %v", all, got)
		}
		// The holder is the permanent's controller: the opponent gets
		// nothing from it, whatever the flag says.
		if perm := grantOn(g, opp.ID, theirs, ZoneGraveyard); perm.Granted() {
			t.Errorf("all=%v: the permission reached a player who does not hold it", all)
		}
	}
}

// ExileCost.ExiledWithSource: a processing cost narrowed to the cards
// exiled with the paying object, any owner's — the paying player's own
// card included — and never a card linked to another object, or to an
// earlier object the source was (CR 400.7).
func TestProcessCostExiledWithSourceReadsTheLinkNotTheOwner(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source := permanentFor(g, me, "Test Shelob", "Creature", "{B}")
	other := permanentFor(g, me, "Other Linker", "Creature", "{B}")
	var srcEpoch, otherEpoch int
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			switch c.InstanceID {
			case source:
				srcEpoch = c.ObjectEpoch
			case other:
				otherEpoch = c.ObjectEpoch
			}
		}
	})
	exiled := func(name string, owner *Player, link PermissionCardRef) Card {
		c := NewCard(name, owner.ID)
		c.TypeLine = "Creature — Bear"
		c.ExiledWith = link
		g.Exile.PushTop(c)
		return c
	}
	theirs := exiled("Their Linked", opp, PermissionCardRef{ID: source, Epoch: srcEpoch})
	mine := exiled("My Linked", me, PermissionCardRef{ID: source, Epoch: srcEpoch})
	stale := exiled("Earlier Object", opp, PermissionCardRef{ID: source, Epoch: srcEpoch - 1})
	elsewhere := exiled("Other Link", opp, PermissionCardRef{ID: other, Epoch: otherEpoch})
	unlinked := exiled("Unlinked", opp, PermissionCardRef{})

	linked := &ExileCost{N: 1, Label: "a creature card exiled with this", From: ZoneExile, ExiledWithSource: true}
	plain := &ExileCost{N: 1, Label: "a card an opponent owns from exile", From: ZoneExile}
	var gotLinked, gotPlain []string
	g.WithWriteLock(func() {
		for _, id := range g.ExileCostOptionsForEffect(me.ID, source, linked) {
			c, _ := g.LookupCardForEffect(id)
			gotLinked = append(gotLinked, c.Name)
		}
		for _, id := range g.ExileCostOptionsForEffect(me.ID, source, plain) {
			c, _ := g.LookupCardForEffect(id)
			gotPlain = append(gotPlain, c.Name)
		}
	})
	slices.Sort(gotLinked)
	if want := []string{mine.Name, theirs.Name}; !slices.Equal(gotLinked, want) {
		t.Errorf("linked options = %v, want %v", gotLinked, want)
	}
	// The plain processing clause is unchanged: every opponent-owned card.
	for _, c := range []Card{theirs, stale, elsewhere, unlinked} {
		if !slices.Contains(gotPlain, c.Name) {
			t.Errorf("plain processing clause lost %q (%v)", c.Name, gotPlain)
		}
	}
	if slices.Contains(gotPlain, mine.Name) {
		t.Errorf("plain processing clause offered the activator's own card")
	}
	// Paying: the validator accepts exactly the linked picks.
	g.WithWriteLock(func() {
		if _, err := g.validateExileCardsCostLocked(me.ID, source, linked, []uuid.UUID{theirs.InstanceID}, nil); err != nil {
			t.Errorf("a linked card was refused: %v", err)
		}
		if _, err := g.validateExileCardsCostLocked(me.ID, source, linked, []uuid.UUID{elsewhere.InstanceID}, nil); err == nil {
			t.Errorf("a card linked to another object paid the cost")
		}
	})
}
