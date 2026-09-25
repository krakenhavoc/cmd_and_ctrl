package ws

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// persist_stack_ability_test.go is ADR 0041 P9's boot half, the owner's
// answer to Q3 (#1497): an ability waiting on the stack whose catalog
// row the new binary cannot find costs the table nothing but that
// ability's automation. The game comes back, the item stays on the
// stack to be resolved by hand, the card is flagged, and the deploy log
// says so in exactly one ERROR line naming the game, the card and the
// row.
func TestRestoreReportsAStackAbilityWhoseRowIsGone(t *testing.T) {
	const oracle = "00000000-0000-4000-8000-000000001497"
	const label = "{0}: Target player loses 1 life."
	prev := game.CatalogLookup
	t.Cleanup(func() { game.CatalogLookup = prev })
	row := func(label string) game.ActivatedAbilityShape {
		return game.ActivatedAbilityShape{
			Label:   label,
			Targets: &game.TargetSpec{Mode: "player", Label: "target player", Players: true, Min: 1, Max: 1},
			Effect:  func(*game.Game, *game.StackItem) error { return nil },
		}
	}
	def := &game.CardDef{Activated: []game.ActivatedAbilityShape{row(label)}}
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
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	probe := uuid.New()
	if _, _, err := room.Apply(me, func() error {
		g.WithWriteLock(func() {
			g.Battlefield.PushTop(game.Card{
				InstanceID: probe, Name: "Stack Probe", OracleID: oracle,
				TypeLine: "Artifact", Owner: me, Controller: me,
			})
		})
		return g.ActivateCatalogAbility(me, probe, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp}},
		})
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// --- the deploy: the row was reworded; the count is unchanged, so
	// this is not a #522 shortfall, only a lost stack ability ---------
	reworded := &game.CardDef{Activated: []game.ActivatedAbilityShape{row(label + " (reworded)")}}
	game.CatalogLookup = func(k string) *game.CardDef {
		if k == oracle {
			return reworded
		}
		return nil
	}
	var logs bytes.Buffer
	mgr = NewRoomManager(slog.New(slog.NewTextHandler(&logs, nil)), dir)
	outcomes := mgr.RestoreRooms()
	LogRestoreSummary(mgr.log, outcomes)

	if len(outcomes) != 1 || !outcomes[0].Restored() {
		t.Fatalf("the table was not restored: %+v", outcomes)
	}
	if lost := outcomes[0].LostStackAbilities; len(lost) != 1 || lost[0].CardID != probe {
		t.Fatalf("outcome lost stack abilities = %+v, want the probe's", lost)
	}
	back := mgr.Get(g.ID)
	var flagged, onStack bool
	back.Game.ReadSnapshot(func() {
		for _, c := range back.Game.Battlefield.Cards {
			if c.InstanceID == probe {
				flagged = c.AbilitiesLostOnRestore
			}
		}
		for _, it := range back.Game.StackMeta {
			if it.SourceCardID == probe && it.Label == label {
				onStack = true
			}
		}
	})
	if !flagged {
		t.Error("the source card is not flagged AbilitiesLostOnRestore")
	}
	if !onStack {
		t.Error("the item is not on the restored stack")
	}

	var lines []string
	for _, l := range strings.Split(logs.String(), "\n") {
		if strings.Contains(l, "level=ERROR") && strings.Contains(l, g.ID.String()) && strings.Contains(l, "Stack Probe") {
			lines = append(lines, l)
		}
	}
	if len(lines) != 1 {
		t.Fatalf("%d ERROR lines name the game and the card, want exactly 1:\n%s", len(lines), logs.String())
	}
	for _, want := range []string{"ability_ref=own:0", "ability_name=\"" + label + "\"", "ability_key=" + oracle} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("the line lacks %q:\n%s", want, lines[0])
		}
	}
	if !strings.Contains(logs.String(), "cards_with_lost_abilities=1") {
		t.Errorf("the summary does not count it:\n%s", logs.String())
	}
}
