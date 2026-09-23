package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// retarget_choice_test.go — #1196's half of battle_choice_test.go's
// lesson. A prompt kind with no case in choiceMoves is a seat with
// zero legal moves, because the engine refuses pass_priority while a
// blocking choice is open; and an enumerated answer the engine then
// refuses is #544 in the other direction. Both are checked here by
// dispatching what the enumerator offered through the real router.

const retargetTestOracle = "legal-test-retarget"

// seedRetargetPrompt puts a one-target spell on the stack under
// `caster` and opens the CR 115.7 offer for `chooser`.
func seedRetargetPrompt(t *testing.T, g *game.Game, caster, chooser uuid.UUID, optional bool) (spellID uuid.UUID, victims []uuid.UUID) {
	t.Helper()
	prev := game.CatalogTargetSpec
	game.CatalogTargetSpec = func(oracleID string) *game.TargetSpec {
		if oracleID != retargetTestOracle {
			return nil
		}
		return &game.TargetSpec{
			Mode:  "creature",
			Label: "target creature",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.IsCreature()
			},
			Min: 1, Max: 1,
		}
	}
	t.Cleanup(func() { game.CatalogTargetSpec = prev })

	for i := 0; i < 3; i++ {
		id := uuid.New()
		victims = append(victims, id)
		g.WithWriteLock(func() {
			g.Battlefield.PushTop(game.Card{
				InstanceID: id, Name: "Victim", TypeLine: "Creature — Test",
				Power: 2, Toughness: 2, Owner: caster, Controller: caster,
			})
		})
	}
	spellID = uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{
			InstanceID: spellID, Name: "Test Bolt", TypeLine: "Instant",
			OracleID: retargetTestOracle, Owner: caster, Controller: caster,
		})
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*game.StackItem{}
		}
		g.StackMeta[spellID] = &game.StackItem{
			ID: spellID, Kind: game.StackItemSpell, Controller: caster, Owner: caster,
			SourceCardID: spellID,
			Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victims[0]}},
		}
		if err := g.OfferRetargetForEffect(game.RetargetOffer{
			ItemID:   spellID,
			Chooser:  chooser,
			Policy:   game.RetargetChangeOne,
			Optional: optional,
			Reason:   "Bolt Bend — change the target",
		}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
	return spellID, victims
}

// A seat owed a retarget prompt is offered one answer per alternative
// and nothing else, and the engine takes the first of them.
func TestRetargetPromptIsEnumeratedAndAccepted(t *testing.T) {
	g := newTable(t)
	caster, chooser := g.Seats[0], g.Seats[1]
	spellID, victims := seedRetargetPrompt(t, g, caster.ID, chooser.ID, false)

	moves := legal.EnumerateFor(g, chooser.ID)
	if len(moves) == 0 {
		t.Fatal("a seat owed a retarget prompt has no legal moves at all — the seat is stuck")
	}
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("offered %q while a choice is open; only resolve_choice is legal", m.Type)
		}
	}
	// Three creatures, one of them already targeted: two alternatives
	// and, since the change is mandatory, no decline.
	if len(moves) != len(victims)-1 {
		t.Fatalf("enumerated %d answers, want one per alternative (%d)", len(moves), len(victims)-1)
	}

	var p struct {
		ChoiceID string `json:"choice_id"`
		Targets  []struct {
			Kind string `json:"kind"`
			ID   string `json:"id"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(moves[0].Params, &p); err != nil {
		t.Fatalf("enumerated params are not resolve_choice params: %v", err)
	}
	if len(p.Targets) != 1 {
		t.Fatalf("a retarget answer is one ref, got %+v", p.Targets)
	}
	if err := actions.Dispatch(g, actions.Action{
		Type:   actions.TypeResolveChoice,
		Player: chooser.ID,
		Params: moves[0].Params,
	}); err != nil {
		t.Fatalf("the engine refused an enumerated retarget answer: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Error("the retarget prompt is still open after a legal answer")
	}
	moved := g.StackMeta[spellID].Targets
	if len(moved) != 1 || moved[0].ID == victims[0] {
		t.Errorf("the spell's target did not move: %+v", moved)
	}
}

// A "you may" retarget offers the DECLINE as well, and offers it
// first — the answer that always terminates.
func TestOptionalRetargetEnumeratesTheDecline(t *testing.T) {
	g := newTable(t)
	caster, chooser := g.Seats[0], g.Seats[1]
	spellID, victims := seedRetargetPrompt(t, g, caster.ID, chooser.ID, true)

	moves := legal.EnumerateFor(g, chooser.ID)
	if len(moves) != len(victims) {
		t.Fatalf("enumerated %d answers, want %d (two alternatives plus the decline)", len(moves), len(victims))
	}
	var p struct {
		Targets []struct{} `json:"targets"`
	}
	if err := json.Unmarshal(moves[0].Params, &p); err != nil {
		t.Fatalf("enumerated params: %v", err)
	}
	if len(p.Targets) != 0 {
		t.Fatalf("the decline should be enumerated first, got %+v", moves[0].Label)
	}
	if err := actions.Dispatch(g, actions.Action{
		Type:   actions.TypeResolveChoice,
		Player: chooser.ID,
		Params: moves[0].Params,
	}); err != nil {
		t.Fatalf("the engine refused an enumerated decline: %v", err)
	}
	if got := g.StackMeta[spellID].Targets; len(got) != 1 || got[0].ID != victims[0] {
		t.Errorf("a declined retarget must leave the target alone: %+v", got)
	}
}
