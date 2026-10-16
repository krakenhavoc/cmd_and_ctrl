package game

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// #3859: a payload entry can record a permanent OBJECT (a card and its
// CR 400.7 epoch), and it survives the snapshot's JSON.
func TestPayloadObjectRefRoundTrips(t *testing.T) {
	o := ObjectRef{ID: uuid.New(), Epoch: 3}
	ref := PayloadObjectRef(o)
	if ref.Kind != TargetCard || ref.Object() != o {
		t.Fatalf("PayloadObjectRef(%v) = %+v, read back %v", o, ref, ref.Object())
	}
	raw, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	var back TargetRef
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if back.Object() != o {
		t.Errorf("after a JSON round trip the object is %v, want %v", back.Object(), o)
	}
	if plain, _ := json.Marshal(TargetRef{Kind: TargetCard, ID: o.ID}); strings.Contains(string(plain), "epoch") {
		t.Errorf("a target with no recorded object writes an epoch: %s", plain)
	}
}

// trigger_zones_library_test.go — #3859: the library and the command zone
// are declared trigger zones, so a "when you discard this card" ability
// follows a card Library of Leng put on top of its owner's library or
// CR 903.9 put into the command zone.

func TestLibraryAndCommandTriggersFireFromTheirOwnZoneOnly(t *testing.T) {
	const oracle = "test-library-zone-oracle"
	for _, tc := range []struct {
		name    string
		declare ZoneKind
		zone    func(g *Game) *Zone
		want    int
	}{
		{"library", ZoneLibrary, func(g *Game) *Zone { return g.Seats[0].Library }, 1},
		{"command", ZoneCommand, func(g *Game) *Zone { return g.Seats[0].Command }, 1},
		{"library card, graveyard declared", ZoneGraveyard, func(g *Game) *Zone { return g.Seats[0].Library }, 0},
		{"graveyard card, library declared", ZoneLibrary, func(g *Game) *Zone { return g.Seats[0].Graveyard }, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			owner := g.Seats[0]
			var w zoneWatch
			withZoneTriggerCard(t, oracle, []TriggeredAbility{w.ability(tc.declare, EventDrawCard)})
			tc.zone(g).PushTop(Card{
				InstanceID: uuid.New(),
				Name:       "Zone Watcher",
				OracleID:   oracle,
				TypeLine:   "Creature — Test",
				Owner:      owner.ID,
				Controller: owner.ID,
			})

			g.WithWriteLock(func() {
				g.EmitEvent(Event{Kind: EventDrawCard, Actor: owner.ID})
			})

			if w.fired != tc.want {
				t.Fatalf("trigger fired %d times, want %d", w.fired, tc.want)
			}
			if tc.want == 1 && w.seenSource.Controller != owner.ID {
				t.Errorf("source controller = %s, want the owner %s (CR 108.4)", w.seenSource.Controller, owner.ID)
			}
		})
	}
}
