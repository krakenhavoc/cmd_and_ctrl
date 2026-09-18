package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// modal_enumeration_test.go — #764 / ADR 0065 §6: the enumerator
// expands a modal announcement's (modes × per-clause targets) product
// under one MaxExpansionPerSource budget, prefers the bullets that
// have legal targets, and offers the repeatable "same mode N times"
// selection first.

func withCatalogModes(t *testing.T, fn func(oracleID string) *game.ModeSpec) {
	t.Helper()
	prev := game.CatalogModeSpec
	game.CatalogModeSpec = fn
	t.Cleanup(func() { game.CatalogModeSpec = prev })
}

// modePickPrompt queues a mode_pick prompt with a fixed offer, without
// standing up a whole modal trigger: the enumerator reads the prompt,
// not the ability.
func modePickPrompt(t *testing.T, g *game.Game, chooser uuid.UUID, opts []int, labels []string, min, max int, repeat bool) *game.PendingChoice {
	t.Helper()
	var out *game.PendingChoice
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:            game.PendingChoiceModePick,
			Chooser:         chooser,
			Count:           1,
			Reason:          "Choose",
			ModeOptionIndex: opts,
			ModeOptionLabel: labels,
			ModeMin:         min,
			ModeMax:         max,
			ModeRepeatable:  repeat,
		})
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceModePick {
				out = c
			}
		}
	})
	if out == nil {
		t.Fatal("prompt not queued")
	}
	return out
}

func TestModePickIsEnumeratedForItsChooser(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	modePickPrompt(t, g, me.ID, []int{0, 2}, []string{"Draw a card.", "Gain 2 life."}, 1, 1, false)

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) == 0 {
		t.Fatal("a seat owing a mode_pick must be offered answers (#499 / #618)")
	}
	seen := map[int]bool{}
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Fatalf("an owed prompt crowds out everything else: %+v", m)
		}
		var p struct {
			Modes []int `json:"modes"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params: %v", err)
		}
		if len(p.Modes) != 1 {
			t.Fatalf("choose one: %v", p.Modes)
		}
		seen[p.Modes[0]] = true
		// The label is the bullet, not the index — a decision log
		// that says "mode 2" tells nobody anything.
		if m.Label == "" {
			t.Error("the move is labelled")
		}
	}
	if !seen[0] || !seen[2] {
		t.Errorf("both offered bullets enumerated, by ModeSpec index: %v", seen)
	}
	if seen[1] {
		t.Error("a bullet the prompt did not offer is never enumerated")
	}
}

func TestRepeatableModePickOffersTheAllOneSelectionFirst(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	c := modePickPrompt(t, g, me.ID, []int{1}, []string{"Draw a card."}, 3, 3, true)

	sels := game.ModePickSelections(c, 12)
	if len(sels) == 0 {
		t.Fatal("one legal bullet and CR 700.2d: 'that bullet three times' is a legal answer")
	}
	if len(sels[0]) != 3 || sels[0][0] != 1 || sels[0][2] != 1 {
		t.Errorf("the all-one-bullet selection comes first: %v", sels[0])
	}
	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) == 0 {
		t.Fatal("the seat is offered an answer")
	}
}

func TestModePickSelectionsRespectTheBudget(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	opts := []int{0, 1, 2, 3, 4, 5}
	labels := make([]string, len(opts))
	for i := range labels {
		labels[i] = "bullet"
	}
	c := modePickPrompt(t, g, me.ID, opts, labels, 1, 6, true)
	for _, budget := range []int{1, 3, 12} {
		if got := len(game.ModePickSelections(c, budget)); got > budget {
			t.Errorf("budget %d: %d selections", budget, got)
		}
	}
	// The enumerator's own cap holds too — an unbounded expansion here
	// is what ADR 0033 §1 exists to prevent.
	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) > 12 {
		t.Errorf("%d moves for one prompt, want ≤ MaxExpansionPerSource", len(moves))
	}
}

// A modal CAST expands modes × targets and never offers a bullet the
// engine would refuse for want of a target.
func TestModalCastEnumeratesPerModeTargets(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	const oracle = "test-two-bullet-charm"

	rock := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: rock, Name: "Rock", TypeLine: "Artifact",
			Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
		})
	})
	withCatalogModes(t, func(id string) *game.ModeSpec {
		if id != oracle {
			return nil
		}
		return &game.ModeSpec{
			Prompt: "Choose one", Min: 1, Max: 1,
			Options: []game.ModeOption{
				{
					Label: "Destroy target artifact.",
					Targets: &game.TargetSpec{
						Mode: "permanent", Label: "target artifact", Zones: []game.ZoneKind{game.ZoneBattlefield},
						CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
							return c.IsArtifact()
						},
						Min: 1, Max: 1,
					},
				},
				{
					Label: "Destroy target enchantment.",
					Targets: &game.TargetSpec{
						Mode: "permanent", Label: "target enchantment", Zones: []game.ZoneKind{game.ZoneBattlefield},
						CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
							return c.IsEnchantment()
						},
						Min: 1, Max: 1,
					},
				},
			},
		}
	})

	spell := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: spell, Name: "Test Charm", TypeLine: "Instant", ManaCost: "{R}",
			OracleID: oracle, Owner: me.ID, Controller: me.ID,
		})
		me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	})

	var casts []legal.Move
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == legal.TypeCastSpell && m.Source == spell {
			casts = append(casts, m)
		}
	}
	if len(casts) != 1 {
		t.Fatalf("one legal bullet, one legal target: %d casts", len(casts))
	}
	var p struct {
		Modes   []int `json:"modes"`
		Targets []struct {
			ID   string `json:"id"`
			Slot int    `json:"slot"`
			Mode int    `json:"mode"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(casts[0].Params, &p); err != nil {
		t.Fatalf("params: %v", err)
	}
	if len(p.Modes) != 1 || p.Modes[0] != 0 {
		t.Errorf("the enchantment bullet has no legal target and is not offered: %v", p.Modes)
	}
	if len(p.Targets) != 1 || p.Targets[0].ID != rock.String() {
		t.Fatalf("the artifact bullet's target: %+v", p.Targets)
	}
	if p.Targets[0].Mode != 0 || p.Targets[0].Slot != 0 {
		t.Errorf("the pick names the occurrence and clause it answers: %+v", p.Targets[0])
	}
}
