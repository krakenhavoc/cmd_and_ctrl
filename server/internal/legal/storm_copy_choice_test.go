package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// storm_copy_choice_test.go — #1238 / ADR 0086 Decision 4. Storm makes
// N copies and each one opens its OWN CR 707.10c "choose new targets"
// prompt, so a bot seat casting Grapeshot on the sixth spell of a turn
// has to walk five prompts in a row.
//
// The lesson this file is an instance of is battle_choice_test.go's:
// the engine refuses pass_priority while a blocking choice is open, so
// a prompt with no enumerated answer is a seat with zero legal moves
// and a table that stops. `pick_target` has had a case in choiceMoves
// since S27 and a single copy has been enumerable since Reverberate
// shipped; what is new here is that storm queues SEVERAL of them at
// once, which is the arrangement nothing had tested.

const stormCopyTestOracle = "legal-test-storm-copy"

// seedStormCopyPrompts puts a one-target spell on the stack under
// `caster` and opens `copies` re-target prompts against it — exactly
// what effects.CopySpell{Count: copies, ChooseNewTargets: true} does
// when a storm trigger resolves.
func seedStormCopyPrompts(t *testing.T, g *game.Game, caster uuid.UUID, copies int) (victims []uuid.UUID) {
	t.Helper()
	prev := game.CatalogTargetSpec
	game.CatalogTargetSpec = func(oracleID string) *game.TargetSpec {
		if oracleID != stormCopyTestOracle {
			return nil
		}
		return &game.TargetSpec{
			Mode:  "creature",
			Label: "any target",
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
	spellID := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{
			InstanceID: spellID, Name: "Test Grapeshot", TypeLine: "Sorcery",
			OracleID: stormCopyTestOracle, Owner: caster, Controller: caster,
		})
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*game.StackItem{}
		}
		g.StackMeta[spellID] = &game.StackItem{
			ID: spellID, Kind: game.StackItemSpell, Controller: caster, Owner: caster,
			SourceCardID: spellID,
			Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victims[0]}},
		}
		for i := 0; i < copies; i++ {
			if err := g.CopySpellForEffect(spellID, caster, true, nil); err != nil {
				t.Fatalf("CopySpellForEffect %d: %v", i, err)
			}
		}
	})
	return victims
}

// TestStormCopyPromptsAreEnumeratedOneAtATime — the seat is never
// stuck, and each answer it gives is one the engine accepts.
func TestStormCopyPromptsAreEnumeratedOneAtATime(t *testing.T) {
	g := newTable(t)
	caster := g.Seats[0]
	victims := seedStormCopyPrompts(t, g, caster.ID, 3)

	if len(g.PendingChoices) != 3 {
		t.Fatalf("seeded %d prompts, want 3 (one per copy)", len(g.PendingChoices))
	}

	for answered := 0; answered < 3; answered++ {
		moves := legal.EnumerateFor(g, caster.ID)
		if len(moves) == 0 {
			t.Fatalf("copy %d: the seat has no legal moves at all while a copy prompt is open", answered)
		}
		for _, m := range moves {
			if m.Type != legal.TypeResolveChoice {
				t.Errorf("copy %d: offered %q while a choice is open; only resolve_choice is legal",
					answered, m.Type)
			}
		}
		// One answer per creature on the board, for EVERY prompt still
		// open — the enumerator offers the whole open set rather than
		// only the newest, so a policy that wants to answer the
		// copies in a particular order can. The copy may keep the
		// original's target or move to either of the others.
		if want := len(victims) * (3 - answered); len(moves) != want {
			t.Fatalf("copy %d: enumerated %d answers, want %d (%d prompts × %d legal targets)",
				answered, len(moves), want, 3-answered, len(victims))
		}
		var p struct {
			ChoiceID string `json:"choice_id"`
			Target   *struct {
				Kind string `json:"kind"`
				ID   string `json:"id"`
			} `json:"target"`
		}
		if err := json.Unmarshal(moves[0].Params, &p); err != nil {
			t.Fatalf("copy %d: enumerated params are not resolve_choice params: %v", answered, err)
		}
		if p.Target == nil {
			t.Fatalf("copy %d: a one-slot copy answer is a single `target` ref, got %s",
				answered, string(moves[0].Params))
		}
		if err := actions.Dispatch(g, actions.Action{
			Type:   actions.TypeResolveChoice,
			Player: caster.ID,
			Params: moves[0].Params,
		}); err != nil {
			t.Fatalf("copy %d: the engine refused an enumerated copy answer: %v", answered, err)
		}
		if want := 2 - answered; len(g.PendingChoices) != want {
			t.Fatalf("copy %d: %d prompts left open, want %d", answered, len(g.PendingChoices), want)
		}
	}

	// Three answers, three copies on the stack beside the original.
	if got := g.Stack.Size(); got != 4 {
		t.Errorf("stack holds %d objects, want 4 (the spell plus three copies)", got)
	}
}

// TestAStormCopyPromptDoesNotLeakToAnotherSeat — the CR 707.10c
// choice belongs to the copy's controller, so nobody else is offered
// it and nobody else is blocked by it.
func TestAStormCopyPromptDoesNotLeakToAnotherSeat(t *testing.T) {
	g := newTable(t)
	caster, bystander := g.Seats[0], g.Seats[1]
	seedStormCopyPrompts(t, g, caster.ID, 2)

	for _, m := range legal.EnumerateFor(g, bystander.ID) {
		if m.Type == legal.TypeResolveChoice {
			t.Errorf("a bystander was offered the copy's target choice: %q", m.Label)
		}
	}
}
