package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// borrowed_abilities_view_test.go — ADR 0155 Decision 10 on the wire: a
// borrowed row's `ref`, its `borrowed_from: {id, name}`, its label with
// the lender's name read as the borrower's (CR 201.5b), and the borrow's
// clause in `granted_abilities`, for every viewer.

const (
	viewBorrowLender = "view-fixture-lender"
	viewBorrowOoze   = "view-fixture-ooze"
	viewBorrowClause = "Has all activated abilities of all creature cards in all graveyards."
	viewBorrowPing   = "{1}: Mana Elf pings."
)

func stubViewBorrows(t *testing.T) {
	t.Helper()
	graveyardCreatures := func(g *game.Game, _ *game.Card) []*game.Card {
		var out []*game.Card
		for _, p := range g.Seats {
			for i := range p.Graveyard.Cards {
				if p.Graveyard.Cards[i].IsCreature() {
					out = append(out, &p.Graveyard.Cards[i])
				}
			}
		}
		return out
	}
	fixtures := map[string]*game.CardDef{
		viewBorrowLender: {
			ManaAbilities: []game.ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}},
			Activated: []game.ActivatedAbilityShape{{
				Label:  viewBorrowPing,
				Cost:   game.AbilityCost{Mana: "{1}"},
				Effect: func(*game.Game, *game.StackItem) error { return nil },
			}},
		},
		viewBorrowOoze: {Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Borrows: &game.AbilityBorrow{Lenders: graveyardCreatures, Label: viewBorrowClause},
		}}},
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
	t.Cleanup(func() { game.CatalogLookup = prev })
}

func TestBorrowedAbilitiesReachTheWireForEveryViewer(t *testing.T) {
	stubViewBorrows(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := game.NewCard("Mana Elf", me.ID)
	elf.TypeLine = "Creature — Elf"
	elf.OracleID = viewBorrowLender
	ooze := game.NewCard("Necrotic Ooze", me.ID)
	ooze.TypeLine = "Creature — Ooze"
	ooze.OracleID = viewBorrowOoze
	g.WithWriteLock(func() {
		g.Seats[1].Graveyard.PushTop(elf)
		ooze.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
		g.Battlefield.PushTop(ooze)
	})
	g.BumpLayerVersionForTest()

	wantRef := "borrow:" + viewBorrowLender + ":0:0"
	v := ViewOfGame(g)
	for _, viewer := range []struct {
		name       string
		id         string
		controller bool
	}{{"controller", me.ID.String(), true}, {"opponent", opp.ID.String(), false}, {"spectator", "", false}} {
		t.Run(viewer.name, func(t *testing.T) {
			fv := FilterViewFor(v, viewer.id)
			var got *CardView
			for i := range fv.Battlefield.Cards {
				if fv.Battlefield.Cards[i].InstanceID == ooze.InstanceID.String() {
					got = &fv.Battlefield.Cards[i]
				}
			}
			if got == nil {
				t.Fatal("the ooze is missing from the view")
			}
			if len(got.ManaAbilities) != 1 {
				t.Fatalf("mana rows = %+v, want the lent {G}", got.ManaAbilities)
			}
			row := got.ManaAbilities[0]
			if row.Ref != wantRef {
				t.Errorf("mana ref = %q, want %q", row.Ref, wantRef)
			}
			if row.BorrowedFrom == nil || row.BorrowedFrom.ID != elf.InstanceID.String() || row.BorrowedFrom.Name != "Mana Elf" {
				t.Errorf("borrowed_from = %+v, want the elf by id and name", row.BorrowedFrom)
			}
			if row.GrantedBy == nil || row.GrantedBy.ID != ooze.InstanceID.String() {
				t.Errorf("granted_by = %+v, want the borrowing ooze", row.GrantedBy)
			}
			if len(got.GrantedAbilities) != 1 || got.GrantedAbilities[0].Text != viewBorrowClause {
				t.Errorf("granted_abilities = %+v, want the borrow's clause", got.GrantedAbilities)
			}
			if viewer.controller && len(got.ActivatedAbilities) != 1 {
				t.Fatalf("activated rows = %+v, want the lent ping", got.ActivatedAbilities)
			}
			for _, a := range got.ActivatedAbilities {
				if a.Ref != wantRef {
					t.Errorf("activated ref = %q, want %q", a.Ref, wantRef)
				}
				if a.Label != "{1}: Necrotic Ooze pings." {
					t.Errorf("label = %q, want the lender's name read as the borrower's (CR 201.5b)", a.Label)
				}
				if a.BorrowedFrom == nil || a.BorrowedFrom.Name != "Mana Elf" {
					t.Errorf("activated borrowed_from = %+v", a.BorrowedFrom)
				}
			}
		})
	}
}

// ADR 0155 owner decision 2: a lender the catalog has no entry for lends
// nothing, and the borrower's card names it so the menu can say
// "<lender>: not automated". A catalogued lender, a basic land and a token
// are not named.
func TestUnautomatedLendersReachTheWire(t *testing.T) {
	stubViewBorrows(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	known := game.NewCard("Mana Elf", me.ID)
	known.TypeLine = "Creature — Elf"
	known.OracleID = viewBorrowLender
	unknown := game.NewCard("Mystery Bear", me.ID)
	unknown.TypeLine = "Creature — Bear"
	unknown.OracleID = "no-such-card"
	ooze := game.NewCard("Necrotic Ooze", me.ID)
	ooze.TypeLine = "Creature — Ooze"
	ooze.OracleID = viewBorrowOoze
	g.WithWriteLock(func() {
		g.Seats[1].Graveyard.PushTop(known)
		g.Seats[1].Graveyard.PushTop(unknown)
		ooze.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
		g.Battlefield.PushTop(ooze)
	})
	g.BumpLayerVersionForTest()

	for _, viewer := range []string{me.ID.String(), opp.ID.String(), ""} {
		fv := FilterViewFor(ViewOfGame(g), viewer)
		var got *CardView
		for i := range fv.Battlefield.Cards {
			if fv.Battlefield.Cards[i].InstanceID == ooze.InstanceID.String() {
				got = &fv.Battlefield.Cards[i]
			}
		}
		if got == nil {
			t.Fatal("the ooze is missing from the view")
		}
		if len(got.UnautomatedLenders) != 1 || got.UnautomatedLenders[0].ID != unknown.InstanceID.String() ||
			got.UnautomatedLenders[0].Name != "Mystery Bear" {
			t.Errorf("viewer %q: unautomated_lenders = %+v, want only the uncatalogued bear", viewer, got.UnautomatedLenders)
		}
	}
}
