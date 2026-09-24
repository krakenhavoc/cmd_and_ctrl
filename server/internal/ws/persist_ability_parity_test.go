package ws

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// persist_ability_parity_test.go is #522's boot half: a card the new
// binary restores with fewer catalog abilities than its restore point
// recorded costs the table nothing but that card's automation — the
// game comes back — and the deploy log says so at ERROR, naming the
// game, the card and the counts.
func TestRestoreReportsACardThatLostItsCatalogEntry(t *testing.T) {
	const oracle = "00000000-0000-4000-8000-00000000522a"
	prev := game.CatalogLookup
	t.Cleanup(func() { game.CatalogLookup = prev })
	def := &game.CardDef{
		Triggered: []game.TriggeredAbility{{Watches: []game.EventKind{game.EventETB}}},
	}
	game.CatalogLookup = func(k string) *game.CardDef {
		if k == oracle {
			return def
		}
		return nil
	}

	dir := t.TempDir()
	mgr := restoreTestManager(t, dir)
	g := newPersistGame(t)
	room := mgr.Create(g)
	probe := uuid.New()
	if _, _, err := room.Apply(g.Seats[0].ID, func() error {
		g.WithWriteLock(func() {
			g.Battlefield.PushTop(game.Card{
				InstanceID: probe,
				Name:       "Parity Probe",
				OracleID:   oracle,
				TypeLine:   "Creature — Test",
				Owner:      g.Seats[0].ID,
				Controller: g.Seats[0].ID,
			})
		})
		return nil
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// --- the deploy: the next binary has no entry for the card -------
	game.CatalogLookup = func(string) *game.CardDef { return nil }
	var logs bytes.Buffer
	mgr = NewRoomManager(slog.New(slog.NewTextHandler(&logs, nil)), dir)
	outcomes := mgr.RestoreRooms()
	LogRestoreSummary(mgr.log, outcomes)

	if len(outcomes) != 1 || !outcomes[0].Restored() {
		t.Fatalf("the table was not restored: %+v", outcomes)
	}
	if sf := outcomes[0].AbilityShortfalls; len(sf) != 1 || sf[0].CardID != probe {
		t.Fatalf("outcome shortfalls = %+v, want the probe", sf)
	}
	back := mgr.Get(g.ID)
	var flagged bool
	back.Game.ReadSnapshot(func() {
		for _, c := range back.Game.Battlefield.Cards {
			if c.InstanceID == probe {
				flagged = c.AbilitiesLostOnRestore && game.Unimplemented(c)
			}
		}
	})
	if !flagged {
		t.Error("the restored card is not flagged manual")
	}

	out := logs.String()
	var line string
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "fewer catalog abilities") {
			line = l
		}
	}
	if line == "" {
		t.Fatalf("no per-card line in the boot log:\n%s", out)
	}
	for _, want := range []string{"level=ERROR", g.ID.String(), `card="Parity Probe"`, oracle, "entry_missing=true"} {
		if !strings.Contains(line, want) {
			t.Errorf("per-card line lacks %q:\n%s", want, line)
		}
	}
	if !strings.Contains(out, "cards_with_lost_abilities=1") {
		t.Errorf("the summary does not count the card:\n%s", out)
	}
}
