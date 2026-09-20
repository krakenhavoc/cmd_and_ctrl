package protocol

// log_spawn_test.go — ADR 0075 §2.4. The spawn line is the condition
// the production spawner ships under: a spawned Treasure is
// indistinguishable from an earned one on the board, so if the log
// does not say where it came from, nothing does.

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func spawnLogOf(t *testing.T, g *game.Game) LogEvent {
	t.Helper()
	v := ViewOfGame(g)
	return findLog(t, v.Log, LogSpawn)
}

func TestSpawnIsNarrated(t *testing.T) {
	g := buildActiveGame(t)
	host, other := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventSpawned, Actor: host.ID, Target: other.ID,
			NewZone: game.ZoneBattlefield, Label: "Treasure", Amount: 2,
		})
	})

	e := spawnLogOf(t, g)
	if e.Amount != 2 || e.Label != "Treasure" || e.NewZone != string(game.ZoneBattlefield) {
		t.Fatalf("entry = %+v, want 2 x Treasure onto the battlefield", e)
	}
	if e.TargetSeat == nil || *e.TargetSeat != 1 {
		t.Fatalf("target_seat = %v, want seat 1", e.TargetSeat)
	}
	want := host.Name + " spawned 2 × Treasure onto " + other.Name + "'s battlefield"
	if e.Text != want {
		t.Errorf("text = %q, want %q", e.Text, want)
	}
	// Nothing has told the projection who holds the table, so the
	// marker is absent rather than assumed.
	if e.ActorIsHost {
		t.Error("actor_is_host set before the room stamped it")
	}
}

// The room stamps the marker on the same pass that stamps the seats'
// is_host flag, and the text is re-rendered with it.
func TestStampHostOnLogMarksTheSpawner(t *testing.T) {
	g := buildActiveGame(t)
	host := g.Seats[0]
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventSpawned, Actor: host.ID, Target: host.ID,
			NewZone: game.ZoneBattlefield, Label: "Sol Ring", Amount: 1,
		})
	})

	v := ViewOfGame(g)
	StampHostOnLog(&v) // no host flagged yet: a no-op
	if findLog(t, v.Log, LogSpawn).ActorIsHost {
		t.Fatal("marked a host on a view that flags none")
	}

	v.Seats[0].IsHost = true
	StampHostOnLog(&v)
	e := findLog(t, v.Log, LogSpawn)
	if !e.ActorIsHost {
		t.Fatal("actor_is_host not set for the host's own spawn")
	}
	want := host.Name + " (host) spawned Sol Ring onto " + host.Name + "'s battlefield"
	if e.Text != want {
		t.Errorf("text = %q, want %q", e.Text, want)
	}

	// A spawn by somebody who is NOT the host — the dev route lets
	// anyone at a preview table spawn — keeps the plain line.
	v2 := ViewOfGame(g)
	v2.Seats[1].IsHost = true
	StampHostOnLog(&v2)
	if strings.Contains(findLog(t, v2.Log, LogSpawn).Text, "(host)") {
		t.Error("marked a non-host spawner as host")
	}
}

// The admin has no seat, so renderLogText's "someone" would be the
// fallback. A spawn has exactly two authorised callers; naming the
// one it must have been is better than naming neither.
func TestAdminSpawnNamesTheAdmin(t *testing.T) {
	g := buildActiveGame(t)
	target := g.Seats[1]
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventSpawned, NewZone: game.ZoneGraveyard,
			Target: target.ID, Label: "Llanowar Elves", Amount: 1,
		})
	})
	e := spawnLogOf(t, g)
	want := "The admin spawned Llanowar Elves into " + target.Name + "'s graveyard"
	if e.Text != want {
		t.Errorf("text = %q, want %q", e.Text, want)
	}
}

// A spawn into a hidden zone names the ZONE and not the card. The
// redaction is at BUILD time because the entry carries no card
// reference for the per-viewer knower filter to key on — so it has to
// be dropped for everyone, the spawner included.
func TestSpawnIntoAHiddenZoneDoesNotNameTheCard(t *testing.T) {
	for _, zone := range []game.ZoneKind{game.ZoneHand, game.ZoneLibrary} {
		g2 := buildActiveGame(t)
		host, victim := g2.Seats[0], g2.Seats[1]
		g2.WithWriteLock(func() {
			g2.EmitEvent(game.Event{
				Kind: game.EventSpawned, Actor: host.ID, Target: victim.ID,
				NewZone: zone, Label: "Black Lotus", Amount: 1,
			})
		})
		e := spawnLogOf(t, g2)
		if e.Label != "" {
			t.Errorf("%s: label = %q, want it dropped", zone, e.Label)
		}
		if strings.Contains(e.Text, "Black Lotus") {
			t.Errorf("%s: text names the card: %q", zone, e.Text)
		}
		if !strings.Contains(e.Text, victim.Name) || !strings.Contains(e.Text, spawnZoneWord(string(zone))) {
			t.Errorf("%s: text = %q, want it to name the seat and the zone", zone, e.Text)
		}
	}
}

// A multi-card spawn into a hidden zone still says how many, because
// the count is public — the table watched the hand grow.
func TestHiddenZoneSpawnKeepsTheCount(t *testing.T) {
	g := buildActiveGame(t)
	host := g.Seats[0]
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventSpawned, Actor: host.ID, Target: host.ID,
			NewZone: game.ZoneHand, Label: "Black Lotus", Amount: 3,
		})
	})
	if got := spawnLogOf(t, g).Text; !strings.Contains(got, "3 cards") {
		t.Errorf("text = %q, want it to say 3 cards", got)
	}
}
