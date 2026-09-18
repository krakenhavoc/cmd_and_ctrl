package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emblem_view_test.go is the wire half of CR 114 (#623, ADR 0064).
// Two things have to be true and neither is obvious from the field's
// declaration: an emblem reaches EVERY seat unredacted (it is face up
// in the command zone and anybody may read it), and it does NOT reach
// the command-zone ZoneView, which the cast surface and the
// hover-zoom both read.

const (
	emblemViewSource = "test-view-walker"
	emblemViewKey    = "emblem:" + emblemViewSource
)

// stubEmblemCatalogForView wires a one-card catalog whose emblem has
// a label and a text, so the projection has something to read.
func stubEmblemCatalogForView(t *testing.T) {
	t.Helper()
	defs := map[string]*game.CardDef{
		emblemViewSource: {},
		emblemViewKey: {
			Static: []game.StaticAbility{{
				Layer:     game.Layer6Ability,
				AppliesTo: func(*game.Card, *game.Game, *game.Card) bool { return false },
				Apply:     func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
			}},
			Emblem: &game.EmblemDef{
				Label: "Test Walker emblem",
				Text:  "Creatures you control get +2/+2 and have flying.",
			},
		},
	}
	prev := game.CatalogLookup
	game.CatalogLookup = func(key string) *game.CardDef { return defs[key] }
	t.Cleanup(func() { game.CatalogLookup = prev })
}

func giveEmblemForView(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	src := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: src,
		Name:       "Test Walker",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleID:   emblemViewSource,
		Owner:      owner,
		Controller: owner,
	})
	var err error
	g.WithWriteLock(func() { err = g.CreateEmblemForEffect(owner, src) })
	if err != nil {
		t.Fatalf("CreateEmblemForEffect: %v", err)
	}
}

func seatViewFor(t *testing.T, g *game.Game, viewer uuid.UUID, seatIdx int) PlayerView {
	t.Helper()
	v := FilterViewFor(ViewOfGame(g), viewer.String())
	if seatIdx >= len(v.Seats) {
		t.Fatalf("seat %d is out of range (%d seats)", seatIdx, len(v.Seats))
	}
	return v.Seats[seatIdx]
}

// TestEmblemsReachEverySeatUnredacted — an emblem is public, so the
// owner's own view and every opponent's view carry the same label and
// text.
func TestEmblemsReachEverySeatUnredacted(t *testing.T) {
	stubEmblemCatalogForView(t)
	g := buildActiveGame(t)
	owner, opponent := g.Seats[0], g.Seats[1]
	giveEmblemForView(t, g, owner.ID)

	for _, viewer := range []*game.Player{owner, opponent} {
		seat := seatViewFor(t, g, viewer.ID, 0)
		if len(seat.Emblems) != 1 {
			t.Fatalf("viewer %s sees %d emblems on seat 0, want 1", viewer.Name, len(seat.Emblems))
		}
		e := seat.Emblems[0]
		if e.Label != "Test Walker emblem" {
			t.Errorf("viewer %s: label = %q", viewer.Name, e.Label)
		}
		if e.Text != "Creatures you control get +2/+2 and have flying." {
			t.Errorf("viewer %s: text = %q", viewer.Name, e.Text)
		}
		if e.InstanceID == "" {
			t.Errorf("viewer %s: no instance id", viewer.Name)
		}
	}
	// A seat with no emblems sends the field at all only when it has
	// one — omitempty keeps every ordinary frame the size it was.
	if seat := seatViewFor(t, g, owner.ID, 1); len(seat.Emblems) != 0 {
		t.Errorf("the opponent's seat carries %d emblems", len(seat.Emblems))
	}
	raw, err := json.Marshal(seatViewFor(t, g, owner.ID, 1))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := string(raw); strings.Contains(got, `"emblems"`) {
		t.Errorf("an emblem-less seat still ships the key: %s", got)
	}
}

// TestEmblemsAreNotInTheCommandZoneView — the command-zone ZoneView
// is the cast surface (stampLegalTargets walks it unconditionally)
// and the commander pile the client renders. An emblem in it would be
// offered as a castable card with a commander tax.
func TestEmblemsAreNotInTheCommandZoneView(t *testing.T) {
	stubEmblemCatalogForView(t)
	g := buildActiveGame(t)
	owner := g.Seats[0]
	before := len(ViewOfGame(g).Seats[0].Command.Cards)
	giveEmblemForView(t, g, owner.ID)

	seat := seatViewFor(t, g, owner.ID, 0)
	if len(seat.Command.Cards) != before {
		t.Errorf("the command ZoneView went %d → %d cards", before, len(seat.Command.Cards))
	}
	if seat.Command.Count != before {
		t.Errorf("the command ZoneView count is %d, want %d", seat.Command.Count, before)
	}
	for _, c := range seat.Command.Cards {
		if c.Name == "Test Walker emblem" {
			t.Error("the emblem is in the command ZoneView — the cast enumerator would offer it")
		}
	}
	v := ViewOfGame(g)
	for _, c := range v.Battlefield.Cards {
		if c.Name == "Test Walker emblem" {
			t.Error("the emblem is on the battlefield view — it is not a permanent (CR 114.4)")
		}
	}
}
