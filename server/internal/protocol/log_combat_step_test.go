package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_combat_step_test.go — #187, ADR 0053 Decision 1: the public log
// carries game.Event.CombatStep onto LogEvent.combat_step, and a tagged
// damage line says which combat damage step dealt it. The engine side
// (when the tag is set) is pinned in internal/game/combat_step_test.go;
// these tests feed the projection events directly.

// damageEntries returns every LogDamage entry, oldest first.
func damageEntries(entries []LogEvent) []LogEvent {
	var out []LogEvent
	for _, e := range entries {
		if e.Kind == LogDamage {
			out = append(out, e)
		}
	}
	return out
}

func TestPublicLogCarriesCombatStep(t *testing.T) {
	g := buildActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	everyone := map[uuid.UUID]bool{atk.ID: true, def.ID: true}

	ace, bears, rogue, bolt := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: ace, Name: "Fencing Ace", TypeLine: "Creature — Human Soldier",
			Owner: atk.ID, Controller: atk.ID, KnownBy: everyone,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: bears, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
			Owner: def.ID, Controller: def.ID, KnownBy: everyone,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: rogue, Name: "Vanilla Rogue", TypeLine: "Creature — Rogue",
			Owner: atk.ID, Controller: atk.ID, KnownBy: everyone,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
			Owner: atk.ID, Controller: atk.ID, KnownBy: everyone,
		})
		g.EmitEvent(game.Event{
			Kind: game.EventStepBegan, Actor: atk.ID, Amount: 5, Label: string(game.StepCombatDamage),
		})
		// Acceptance criterion 1's log: Ace (first_strike, regular) and
		// the Bears (regular).
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: atk.ID, Source: ace, Target: bears, Amount: 1,
			Combat: true, CombatStep: game.CombatStepFirstStrike,
		})
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: atk.ID, Source: ace, Target: bears, Amount: 1,
			Combat: true, CombatStep: game.CombatStepRegular,
		})
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: def.ID, Source: bears, Target: ace, Amount: 2,
			Combat: true, CombatStep: game.CombatStepRegular,
		})
		// Untagged combat damage: a combat with no first-strike step.
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: atk.ID, Source: rogue, Target: def.ID, Amount: 3,
			Combat: true,
		})
		// Non-combat damage.
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: atk.ID, Source: bolt, Target: def.ID, Amount: 3,
		})
	})

	want := []struct {
		step string
		text string
	}{
		{game.CombatStepFirstStrike, "Fencing Ace dealt 1 combat damage to Grizzly Bears (first strike)"},
		{game.CombatStepRegular, "Fencing Ace dealt 1 combat damage to Grizzly Bears (regular damage)"},
		{game.CombatStepRegular, "Grizzly Bears dealt 2 combat damage to Fencing Ace (regular damage)"},
		{"", "Vanilla Rogue dealt 3 combat damage to P2"},
		{"", "Lightning Bolt dealt 3 damage to P2"},
	}
	check := func(label string, log []LogEvent) {
		t.Helper()
		got := damageEntries(log)
		if len(got) != len(want) {
			t.Fatalf("%s: %d damage entries, want %d: %+v", label, len(got), len(want), got)
		}
		for i, w := range want {
			if got[i].CombatStep != w.step {
				t.Errorf("%s: entry %d combat_step = %q, want %q", label, i, got[i].CombatStep, w.step)
			}
			if got[i].Text != w.text {
				t.Errorf("%s: entry %d text = %q, want %q", label, i, got[i].Text, w.text)
			}
		}
	}
	check("unfiltered", ViewOfGame(g).Log)
	// FilterViewFor re-renders Text; the tag and the suffix must survive.
	check("seat view", FilterViewFor(ViewOfGame(g), def.ID.String()).Log)
}

// A redacted source re-renders without its name but keeps the step: the
// tag is public (combat damage is between battlefield permanents and
// players), only the card's identity is not.
func TestPublicLogCombatStepSurvivesRedaction(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0], g.Seats[1]
	morph := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: morph, Name: "Secret Striker", TypeLine: "Creature — Knight",
			Owner: owner.ID, Controller: owner.ID, FaceDown: true,
			KnownBy: map[uuid.UUID]bool{owner.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: owner.ID, Source: morph, Target: other.ID, Amount: 2,
			Combat: true, CombatStep: game.CombatStepFirstStrike,
		})
	})
	entry := findLog(t, ViewOfGameFor(g, other.ID.String()).Log, LogDamage)
	if want := "a card dealt 2 combat damage to P2 (first strike)"; entry.Text != want {
		t.Errorf("redacted text = %q, want %q", entry.Text, want)
	}
	if entry.CombatStep != game.CombatStepFirstStrike {
		t.Errorf("redacted entry combat_step = %q, want %q", entry.CombatStep, game.CombatStepFirstStrike)
	}
}

// The wire shape: combat_step is omitempty, so untagged entries — every
// combat without first strike, and all non-combat damage — cost no
// bytes, and a tagged one carries the snake_case value.
func TestLogEventCombatStepWireShape(t *testing.T) {
	tagged, err := json.Marshal(LogEvent{Kind: LogDamage, Combat: true, CombatStep: game.CombatStepRegular})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(tagged), `"combat_step":"regular"`) {
		t.Errorf("tagged entry JSON %s lacks \"combat_step\":\"regular\"", tagged)
	}
	untagged, err := json.Marshal(LogEvent{Kind: LogDamage, Combat: true})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(untagged), "combat_step") {
		t.Errorf("untagged entry JSON %s carries a combat_step key", untagged)
	}
}
