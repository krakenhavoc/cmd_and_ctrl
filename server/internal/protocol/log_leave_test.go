package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_leave_test.go — #4395, ADR 0209 §3: the public log tells a client
// why a permanent left the battlefield (`cause`), which card was behind it
// (`source`) and which exits happened at once (`batch`).

func leaveTestBear(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
	return id
}

// zoneEntriesFor returns the LogZone entries for ids, in log order.
func zoneEntriesFor(entries []LogEvent, ids ...uuid.UUID) []LogEvent {
	want := map[string]bool{}
	for _, id := range ids {
		want[id.String()] = true
	}
	var out []LogEvent
	for _, e := range entries {
		if e.Kind == LogZone && want[e.CardID] {
			out = append(out, e)
		}
	}
	return out
}

// The engine's own batch, cause and source reach the wire JSON under the
// documented names, and one wipe's exits share a batch and a source.
func TestLogForwardsBatchCauseAndSourceToTheWire(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	wipe := uuid.New()
	var victims []uuid.UUID
	g.WithWriteLock(func() {
		for _, name := range []string{"A", "B", "C"} {
			victims = append(victims, leaveTestBear(g, me.ID, name))
		}
		// The wipe is a card the whole table can see, as a spell on the
		// stack would be.
		known := map[uuid.UUID]bool{}
		for _, seat := range g.Seats {
			known[seat.ID] = true
		}
		g.Battlefield.PushTop(game.Card{InstanceID: wipe, Name: "Wipe", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID, KnownBy: known})
		for _, v := range victims {
			g.EmitEvent(game.Event{
				Kind: game.EventZoneMove, Actor: me.ID, CardID: v,
				OldZone: game.ZoneBattlefield, NewZone: game.ZoneGraveyard,
				LeaveCause: game.LeaveCauseDestroy, LeaveSource: wipe,
			})
		}
	})

	log := ViewOfGameFor(g, me.ID.String()).Log
	zones := zoneEntriesFor(log, victims...)
	if len(zones) != 3 {
		t.Fatalf("got %d zone entries, want 3: %v", len(zones), logKinds(log))
	}
	for _, e := range zones {
		if e.Batch == 0 || e.Batch != zones[0].Batch {
			t.Errorf("batch %d, want one non-zero batch for the whole wipe (%d)", e.Batch, zones[0].Batch)
		}
		if e.Cause != "destroy" {
			t.Errorf("cause = %q, want destroy", e.Cause)
		}
		if e.Source != wipe.String() {
			t.Errorf("source = %q, want the wipe %s", e.Source, wipe)
		}
	}

	// The names on the wire.
	buf, err := json.Marshal(zones[0])
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(buf, &raw); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]any{"cause": "destroy", "source": wipe.String(), "batch": float64(zones[0].Batch)} {
		if raw[key] != want {
			t.Errorf("wire field %q = %v, want %v (entry %s)", key, raw[key], want, buf)
		}
	}
}

// Events from the real engine movers carry their cause through to the wire,
// and two exits made by one effect are one batch.
func TestLogCarriesTheEnginesLeaveCauses(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	var destroyed, damaged, sacrificed, exiled, bounced uuid.UUID
	g.WithWriteLock(func() {
		destroyed = leaveTestBear(g, me.ID, "Destroyed")
		damaged = leaveTestBear(g, me.ID, "Damaged")
		sacrificed = leaveTestBear(g, me.ID, "Sacrificed")
		exiled = leaveTestBear(g, me.ID, "Exiled")
		bounced = leaveTestBear(g, me.ID, "Bounced")
	})
	g.WithWriteLock(func() {
		g.DestroyPermanentsForEffect([]uuid.UUID{destroyed})
		g.ExileCardsForEffect([]uuid.UUID{exiled})
		g.BounceCardsToHandForEffect([]uuid.UUID{bounced})
	})
	if err := g.MarkDamage(damaged, 2); err != nil {
		t.Fatal(err)
	}
	if err := g.SacrificePermanent(me.ID, sacrificed); err != nil {
		t.Fatal(err)
	}

	log := ViewOfGameFor(g, me.ID.String()).Log
	for id, want := range map[uuid.UUID]string{
		destroyed: "destroy", damaged: "sba", exiled: "exile_effect", bounced: "bounce",
	} {
		got := zoneEntriesFor(log, id)
		if len(got) != 1 || got[0].Cause != want {
			t.Errorf("zone entries for %s = %+v, want one with cause %q", id, got, want)
		}
	}
	// A sacrifice is its own line, whose kind says so; its zone move is not
	// repeated, but the line is grouped like the rest.
	var sac []LogEvent
	for _, e := range log {
		if e.Kind == LogSacrifice && e.CardID == sacrificed.String() {
			sac = append(sac, e)
		}
	}
	if len(sac) != 1 || sac[0].Batch == 0 {
		t.Errorf("sacrifice entries = %+v, want one carrying a batch", sac)
	}
	if got := zoneEntriesFor(log, sacrificed); len(got) != 0 {
		t.Errorf("the sacrificed card's zone move was repeated: %+v", got)
	}
}

// A source the viewer may not identify is dropped from the entry, as the
// card would be. The batch still groups it with the rest.
func TestLogSourceIsRedactedLikeACard(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0], g.Seats[1]
	secret := uuid.New()
	var victim uuid.UUID
	g.WithWriteLock(func() {
		victim = leaveTestBear(g, owner.ID, "Victim")
		// A face-down permanent whose ability did it: on the battlefield, but
		// known only to its controller.
		g.Battlefield.PushTop(game.Card{
			InstanceID: secret, Name: "Secret Sphinx", TypeLine: "Creature — Sphinx",
			Owner: owner.ID, Controller: owner.ID, FaceDown: true,
			KnownBy: map[uuid.UUID]bool{owner.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventZoneMove, Actor: owner.ID, CardID: victim,
			OldZone: game.ZoneBattlefield, NewZone: game.ZoneGraveyard,
			LeaveCause: game.LeaveCauseDestroy, LeaveSource: secret,
		})
	})

	ownerEntry := zoneEntriesFor(ViewOfGameFor(g, owner.ID.String()).Log, victim)
	otherEntry := zoneEntriesFor(ViewOfGameFor(g, other.ID.String()).Log, victim)
	if len(ownerEntry) != 1 || len(otherEntry) != 1 {
		t.Fatalf("entries: owner %d, other %d, want 1 each", len(ownerEntry), len(otherEntry))
	}
	if ownerEntry[0].Source != secret.String() {
		t.Errorf("the controller lost the source: %+v", ownerEntry[0])
	}
	if otherEntry[0].Source != "" {
		t.Errorf("a non-knower was handed the face-down source: %+v", otherEntry[0])
	}
	if otherEntry[0].Batch == 0 || otherEntry[0].Batch != ownerEntry[0].Batch || otherEntry[0].Cause != "destroy" {
		t.Errorf("redaction dropped more than the source: %+v", otherEntry[0])
	}
	buf, _ := json.Marshal(otherEntry[0])
	if strings.Contains(string(buf), secret.String()) || strings.Contains(string(buf), "Sphinx") {
		t.Errorf("the face-down source leaked to a non-knower: %s", buf)
	}
	// Re-filtering an already filtered log must not widen it.
	once := FilterViewFor(ViewOfGame(g), owner.ID.String())
	twice := FilterViewFor(once, other.ID.String())
	if e := zoneEntriesFor(twice.Log, victim); len(e) != 1 || e[0].Source != "" {
		t.Errorf("a second filter pass re-revealed the source: %+v", e)
	}
}
