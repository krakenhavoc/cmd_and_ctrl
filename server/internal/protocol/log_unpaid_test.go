package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0118 §2 and owner decision 4 (#2188): a cast made without paying
// its mana cost (force_cast) says so in the log, in exactly these
// words, after the zone it was cast from. A paid or permissive cast's
// line is unchanged.
func TestUnpaidCastLogLine(t *testing.T) {
	cases := []struct {
		name   string
		zone   game.ZoneKind
		unpaid bool
		want   string
	}{
		{"unpaid from hand", game.ZoneHand, true, "P1 cast Craw Wurm without paying its mana cost"},
		{"unpaid from exile", game.ZoneExile, true, "P1 cast Craw Wurm from exile without paying its mana cost"},
		{"paid from hand", game.ZoneHand, false, "P1 cast Craw Wurm"},
		{"paid from exile", game.ZoneExile, false, "P1 cast Craw Wurm from exile"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := buildActiveGame(t)
			caster := g.Seats[0]
			wurm := uuid.New()
			g.WithWriteLock(func() {
				// A spell on the stack is public: every seat knows it, as
				// MoveCard marks a card arriving in a public zone.
				g.Stack.PushTop(game.Card{
					InstanceID: wurm, Name: "Craw Wurm", TypeLine: "Creature — Wurm",
					Owner: caster.ID, Controller: caster.ID,
					KnownBy: map[uuid.UUID]bool{g.Seats[0].ID: true, g.Seats[1].ID: true},
				})
				g.EmitEvent(game.Event{
					Kind: game.EventCast, Actor: caster.ID, Source: wurm, CardID: wurm,
					OldZone: tc.zone, NewZone: game.ZoneStack, Unpaid: tc.unpaid,
				})
			})
			for _, viewer := range []string{caster.ID.String(), g.Seats[1].ID.String()} {
				cast := findLog(t, ViewOfGameFor(g, viewer).Log, LogCast)
				if cast.Text != tc.want {
					t.Errorf("viewer %s: text %q, want %q", viewer, cast.Text, tc.want)
				}
				if cast.Unpaid != tc.unpaid {
					t.Errorf("viewer %s: unpaid %v, want %v", viewer, cast.Unpaid, tc.unpaid)
				}
			}
		})
	}
}

// The flag is public (the pool and the cost are), but the card's name
// is redacted per viewer as for any cast: a face-down spell cast
// without paying stays "a card" for an opponent and keeps the suffix.
func TestUnpaidCastLogLineRedactsAFaceDownSpell(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	other := g.Seats[1]
	morph := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{
			InstanceID: morph, Name: "Sagu Mauler", TypeLine: "Creature — Beast",
			Owner: owner.ID, Controller: owner.ID, FaceDown: true,
			KnownBy: map[uuid.UUID]bool{owner.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventCast, Actor: owner.ID, Source: morph, CardID: morph,
			OldZone: game.ZoneHand, NewZone: game.ZoneStack, Unpaid: true,
		})
	})

	ownerEntry := findLog(t, ViewOfGameFor(g, owner.ID.String()).Log, LogCast)
	if want := "P1 cast Sagu Mauler without paying its mana cost"; ownerEntry.Text != want {
		t.Errorf("owner text: %q, want %q", ownerEntry.Text, want)
	}
	otherLog := ViewOfGameFor(g, other.ID.String()).Log
	otherEntry := findLog(t, otherLog, LogCast)
	if want := "P1 cast a card without paying its mana cost"; otherEntry.Text != want {
		t.Errorf("opponent text: %q, want %q", otherEntry.Text, want)
	}
	if !otherEntry.Unpaid {
		t.Error("the unpaid flag was redacted; it is public")
	}
	buf, err := json.Marshal(otherLog)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(buf), "Sagu Mauler") {
		t.Errorf("card name leaked to a non-knower: %s", buf)
	}
	if !strings.Contains(string(buf), `"unpaid":true`) {
		t.Errorf("wire entry lacks unpaid: %s", buf)
	}
}

// A paid cast's wire entry carries no unpaid key at all (omitempty), so
// the field costs nothing on the ordinary line.
func TestPaidCastLogEntryOmitsUnpaid(t *testing.T) {
	g := buildActiveGame(t)
	caster := g.Seats[0]
	bolt := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", Owner: caster.ID, Controller: caster.ID})
		g.EmitEvent(game.Event{Kind: game.EventCast, Actor: caster.ID, Source: bolt, CardID: bolt, OldZone: game.ZoneHand, NewZone: game.ZoneStack})
	})
	buf, err := json.Marshal(findLog(t, ViewOfGame(g).Log, LogCast))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(buf), "unpaid") {
		t.Errorf("a paid cast's entry carries unpaid: %s", buf)
	}
}
