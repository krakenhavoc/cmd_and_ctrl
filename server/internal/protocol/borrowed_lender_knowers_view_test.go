package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// borrowed_lender_knowers_view_test.go — ADR 0155 Decision 10's per-viewer
// half (#3202): a borrowed row's lender in a hidden zone (a library top, a
// hand) is named on the wire to the seats that may see it there and to no
// one else. The row itself, with the borrower's label, is public.

const (
	viewTopBorrower  = "view-fixture-top-borrower"
	viewHandBorrower = "view-fixture-hand-borrower"
	viewTopRevealer  = "view-fixture-top-revealer"
	viewTopPeeker    = "view-fixture-top-peeker"
)

func stubViewHiddenLenders(t *testing.T) {
	t.Helper()
	stubViewBorrows(t)
	self := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID
	}
	libraryTop := func(g *game.Game, source *game.Card) []*game.Card {
		for _, p := range g.Seats {
			if p.ID == source.Controller && len(p.Library.Cards) > 0 {
				return []*game.Card{&p.Library.Cards[len(p.Library.Cards)-1]}
			}
		}
		return nil
	}
	hand := func(g *game.Game, source *game.Card) []*game.Card {
		var out []*game.Card
		for _, p := range g.Seats {
			if p.ID != source.Controller {
				continue
			}
			for i := range p.Hand.Cards {
				if p.Hand.Cards[i].OracleID == viewBorrowLender {
					out = append(out, &p.Hand.Cards[i])
				}
			}
		}
		return out
	}
	fixtures := map[string]*game.CardDef{
		viewTopBorrower: {Static: []game.StaticAbility{{
			Layer: game.Layer6Ability, AppliesTo: self, DependsOnLibraryTop: true,
			Borrows: &game.AbilityBorrow{Lenders: libraryTop, Label: "Has all activated abilities of the top card of your library."},
		}}},
		viewHandBorrower: {Static: []game.StaticAbility{{
			Layer: game.Layer6Ability, AppliesTo: self,
			Borrows: &game.AbilityBorrow{Lenders: hand, Label: "Has all activated abilities of the Elves in your hand."},
		}}},
		viewTopRevealer: {},
		viewTopPeeker:   {},
	}
	prev := game.CatalogLookup
	game.CatalogLookup = func(key string) *game.CardDef {
		if d, ok := fixtures[key]; ok {
			return d
		}
		if prev == nil {
			return nil
		}
		return prev(key)
	}
	prevTop := game.CatalogLibraryTopVisible
	game.CatalogLibraryTopVisible = func(key string) game.LibraryTopVisibility {
		switch key {
		case viewTopRevealer:
			return game.LibraryTopRevealed
		case viewTopPeeker:
			return game.LibraryTopOwner
		}
		if prevTop == nil {
			return game.LibraryTopHidden
		}
		return prevTop(key)
	}
	t.Cleanup(func() {
		game.CatalogLookup = prev
		game.CatalogLibraryTopVisible = prevTop
	})
}

// lenderSeenBy reports, for each viewer, whether the borrower's lent mana
// row names its lender, and fails if the row itself is missing.
func lenderSeenBy(t *testing.T, g *game.Game, borrower uuid.UUID, viewers ...string) []bool {
	t.Helper()
	v := ViewOfGame(g)
	out := make([]bool, 0, len(viewers))
	for _, viewer := range viewers {
		fv := FilterViewFor(v, viewer)
		var got *CardView
		for i := range fv.Battlefield.Cards {
			if fv.Battlefield.Cards[i].InstanceID == borrower.String() {
				got = &fv.Battlefield.Cards[i]
			}
		}
		if got == nil || len(got.ManaAbilities) != 1 {
			t.Fatalf("viewer %q: the borrower's lent mana row is missing: %+v", viewer, got)
		}
		row := got.ManaAbilities[0]
		if row.Ref == "" || row.Label == "" {
			t.Errorf("viewer %q: the lent row lost its ref or label: %+v", viewer, row)
		}
		seen := row.BorrowedFrom != nil
		if seen && row.BorrowedFrom.Name != "Mana Elf" {
			t.Errorf("viewer %q: borrowed_from = %+v", viewer, row.BorrowedFrom)
		}
		for _, a := range got.ActivatedAbilities {
			if (a.BorrowedFrom != nil) != seen {
				t.Errorf("viewer %q: the activated row and the mana row disagree on the lender", viewer)
			}
		}
		out = append(out, seen)
	}
	return out
}

func TestBorrowedFromALibraryTopReachesOnlyTheSeatsThatSeeIt(t *testing.T) {
	stubViewHiddenLenders(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := game.NewCard("Mana Elf", me.ID)
	elf.TypeLine = "Creature — Elf"
	elf.OracleID = viewBorrowLender
	borrower := game.NewCard("Snoop", me.ID)
	borrower.TypeLine = "Creature — Goblin"
	borrower.OracleID = viewTopBorrower
	borrower.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	g.WithWriteLock(func() {
		me.Library.PushTop(elf)
		g.Battlefield.PushTop(borrower)
	})
	g.BumpLayerVersionForTest()
	viewers := []string{me.ID.String(), opp.ID.String(), SpectatorViewerID, ""}

	// Nobody may look at the top: only the admin frame names it.
	if got := lenderSeenBy(t, g, borrower.InstanceID, viewers...); got[0] || got[1] || got[2] || !got[3] {
		t.Errorf("hidden top: seen by [me opp spectator admin] = %v, want only the admin", got)
	}

	// "You may look at the top card of your library": the owner only.
	peeker := game.NewCard("Peeker", me.ID)
	peeker.OracleID = viewTopPeeker
	peeker.TypeLine = "Artifact"
	g.WithWriteLock(func() { g.Battlefield.PushTop(peeker) })
	g.BumpLayerVersionForTest()
	if got := lenderSeenBy(t, g, borrower.InstanceID, viewers...); !got[0] || got[1] || got[2] || !got[3] {
		t.Errorf("owner may look: seen by [me opp spectator admin] = %v, want me and the admin", got)
	}

	// "Play with the top card of your library revealed": every seat.
	revealer := game.NewCard("Revealer", me.ID)
	revealer.OracleID = viewTopRevealer
	revealer.TypeLine = "Artifact"
	g.WithWriteLock(func() { g.Battlefield.PushTop(revealer) })
	g.BumpLayerVersionForTest()
	if got := lenderSeenBy(t, g, borrower.InstanceID, viewers...); !got[0] || !got[1] || !got[2] || !got[3] {
		t.Errorf("revealed top: seen by [me opp spectator admin] = %v, want everyone", got)
	}
}

func TestBorrowedFromAHandCardReachesItsOwnerAndTheSeatsItWasRevealedTo(t *testing.T) {
	stubViewHiddenLenders(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := game.NewCard("Mana Elf", me.ID)
	elf.TypeLine = "Creature — Elf"
	elf.OracleID = viewBorrowLender
	borrower := game.NewCard("Hand Reader", me.ID)
	borrower.TypeLine = "Creature — Wizard"
	borrower.OracleID = viewHandBorrower
	borrower.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	g.WithWriteLock(func() {
		me.Hand.PushTop(elf)
		g.Battlefield.PushTop(borrower)
	})
	g.BumpLayerVersionForTest()
	viewers := []string{me.ID.String(), opp.ID.String(), SpectatorViewerID}
	if got := lenderSeenBy(t, g, borrower.InstanceID, viewers...); !got[0] || got[1] || got[2] {
		t.Errorf("unrevealed: seen by [me opp spectator] = %v, want only me", got)
	}
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == elf.InstanceID {
				me.Hand.Cards[i].KnownBy = map[uuid.UUID]bool{opp.ID: true}
			}
		}
	})
	g.BumpLayerVersionForTest()
	if got := lenderSeenBy(t, g, borrower.InstanceID, viewers...); !got[0] || !got[1] || !got[2] {
		t.Errorf("revealed to the opponent: seen by [me opp spectator] = %v, want every seat", got)
	}
}
