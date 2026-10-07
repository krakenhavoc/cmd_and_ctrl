package protocol

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// named_activator_view_test.go — the view half of ADR 0106 §1's
// 2026-10-07 amendment (#1947): an opponents-only row reaches every seat
// marked `opponents_only`, an owner-only one marked `owner_only`, and the
// active opponent's digest lists the row the controller's does not.

const (
	oracleClergyView      = "66566999-f70a-4f14-9bf0-23325295a977"
	oracleIncarnationView = "6e49a5b8-6bc4-4c7b-82c1-957f1fb0ca5f"
)

func TestOpponentsOnlyRowIsMarkedAndInTheOpponentsDigest(t *testing.T) {
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clergy := put(g.Battlefield, controller, game.Card{
		Name: "Clergy of the Holy Nimbus", TypeLine: "Creature — Human Cleric",
		Power: 1, Toughness: 1, OracleID: oracleClergyView,
	})
	knowTheTable(g)

	av := ViewOfGameFor(g, active.ID.String())
	row := rowOf(t, av, clergy)
	if !row.OpponentsOnly || row.AnyPlayer || row.OwnerOnly {
		t.Errorf("row flags = any %v, opponents %v, owner %v; want opponents only", row.AnyPlayer, row.OpponentsOnly, row.OwnerOnly)
	}
	if av.LegalActions == nil {
		t.Fatal("the active seat got no digest")
	}
	if e := av.LegalActions.Sources[clergy.String()]; e == nil || !slices.Equal(e.Abilities, []string{"own:0"}) {
		t.Errorf("active opponent's digest entry = %+v, want abilities [own:0]", e)
	}
	if !rowOf(t, ViewOfGameFor(g, controller.ID.String()), clergy).OpponentsOnly {
		t.Error("the controller's copy of the row is not marked opponents_only, so the client cannot grey it")
	}
}

func TestOpponentsOnlyRowIsAbsentFromTheControllersDigest(t *testing.T) {
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	clergy := put(g.Battlefield, active, game.Card{
		Name: "Clergy of the Holy Nimbus", TypeLine: "Creature — Human Cleric",
		Power: 1, Toughness: 1, OracleID: oracleClergyView,
	})
	knowTheTable(g)
	av := ViewOfGameFor(g, active.ID.String())
	if av.LegalActions != nil {
		if e := av.LegalActions.Sources[clergy.String()]; e != nil && len(e.Abilities) != 0 {
			t.Errorf("the controller's digest lists its own opponents-only row: %+v", e)
		}
	}
}

// The per-seat copy: an opponents-only row is stamped with the VIEWER as
// activator, so its hand-read option list is the viewer's own hand and
// reaches that seat alone (#1369), exactly as for an any-player row.
func TestOpponentsOnlyRowIsStampedForTheOpponent(t *testing.T) {
	const oracle = "test-protocol-opponents-only-discard"
	prev := game.CatalogActivatedAbilities
	game.CatalogActivatedAbilities = func(key string) []game.ActivatedAbilityShape {
		if key == oracle {
			return []game.ActivatedAbilityShape{{
				Label:         "Discard a card: Nothing. Only your opponents may activate this ability.",
				Cost:          game.AbilityCost{DiscardCards: &game.DiscardCost{N: 1, Label: "a card"}},
				OpponentsOnly: true,
				Effect:        func(*game.Game, *game.StackItem) error { return nil },
			}}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogActivatedAbilities = prev })

	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%4]
	src := put(g.Battlefield, controller, game.Card{Name: "Test Warden", TypeLine: "Enchantment", OracleID: oracle})
	knowTheTable(g)

	var want []string
	for _, c := range active.Hand.Cards {
		want = append(want, c.InstanceID.String())
	}
	if got := rowOf(t, ViewOfGameFor(g, active.ID.String()), src).DiscardCostOptions; !slices.Equal(got, want) {
		t.Errorf("the opponent's discard options %v, want their own hand %v", got, want)
	}
	if got := rowOf(t, ViewOfGameFor(g, controller.ID.String()), src).DiscardCostOptions; len(want) > 0 && slices.Equal(got, want) {
		t.Errorf("the controller's copy carries the opponent's hand %v", got)
	}
}

func TestOwnerOnlyRowIsMarked(t *testing.T) {
	g := busyTable(t, 1)
	owner := g.Seats[g.Turn.ActiveSeat]
	thief := g.Seats[(g.Turn.ActiveSeat+1)%4]
	inc := put(g.Battlefield, owner, game.Card{
		Name: "Personal Incarnation", TypeLine: "Creature — Avatar Incarnation",
		Power: 6, Toughness: 6, OracleID: oracleIncarnationView,
	})
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == inc {
			g.Battlefield.Cards[i].Controller = thief.ID
		}
	}
	knowTheTable(g)
	av := ViewOfGameFor(g, owner.ID.String())
	if row := rowOf(t, av, inc); !row.OwnerOnly || row.OpponentsOnly || row.AnyPlayer {
		t.Errorf("row flags = %+v, want owner only", row)
	}
	if e := av.LegalActions.Sources[inc.String()]; e == nil || !slices.Equal(e.Abilities, []string{"own:0"}) {
		t.Errorf("the owner's digest entry = %+v, want abilities [own:0]", e)
	}
}
