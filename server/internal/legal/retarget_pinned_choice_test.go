package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// retarget_pinned_choice_test.go — #1743. A pinned retarget (Spellskite)
// reuses the retarget prompt kind, so it reuses the enumerator's case
// too; what changes is what the options MEAN — the current targets
// that could become the destination, not new destinations. These pin
// that every enumerated answer is one the engine accepts (#544) and
// that each moves its own slot onto the destination.

const pinnedTestOracle = "legal-test-pinned-retarget"

// seedPinnedPrompt puts a two-clause spell ("target creature, then
// another target creature") on the stack under `caster` at two of
// three creatures, and opens a pinned offer for `chooser` onto the
// third. Returns the spell, its two targets and the destination.
func seedPinnedPrompt(t *testing.T, g *game.Game, caster, chooser uuid.UUID, optional bool) (spellID uuid.UUID, targets []uuid.UUID, dest uuid.UUID) {
	t.Helper()
	creature := func() *game.TargetSpec {
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
	prev := game.CatalogTargetSpec
	game.CatalogTargetSpec = func(oracleID string) *game.TargetSpec {
		if oracleID != pinnedTestOracle {
			return nil
		}
		second := creature()
		second.Label = "another target creature"
		second.Distinct = true
		return creature().Then(second)
	}
	t.Cleanup(func() { game.CatalogTargetSpec = prev })

	var ids []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		ids = append(ids, id)
		g.WithWriteLock(func() {
			g.Battlefield.PushTop(game.Card{
				InstanceID: id, Name: "Creature", TypeLine: "Creature — Test",
				Power: 2, Toughness: 2, Owner: chooser, Controller: chooser,
			})
		})
	}
	spellID = uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{
			InstanceID: spellID, Name: "Test Two-Target", TypeLine: "Instant",
			OracleID: pinnedTestOracle, Owner: caster, Controller: caster,
		})
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*game.StackItem{}
		}
		g.StackMeta[spellID] = &game.StackItem{
			ID: spellID, Kind: game.StackItemSpell, Controller: caster, Owner: caster,
			SourceCardID: spellID,
			Targets: []game.TargetRef{
				{Kind: game.TargetCard, ID: ids[0], Slot: 0},
				{Kind: game.TargetCard, ID: ids[1], Slot: 1},
			},
		}
		if err := g.OfferRetargetForEffect(game.RetargetOffer{
			ItemID:   spellID,
			Chooser:  chooser,
			Policy:   game.RetargetChangeOne,
			Optional: optional,
			To:       game.TargetRef{Kind: game.TargetCard, ID: ids[2]},
			Reason:   "Spellskite — choose the target to change to Spellskite",
		}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
	return spellID, ids[:2], ids[2]
}

// Every enumerated answer is accepted, and each changes the slot whose
// current target it names — never the other one, and never to anything
// but the destination.
func TestPinnedRetargetPromptIsEnumeratedAndEveryAnswerAccepted(t *testing.T) {
	for pick := 0; pick < 2; pick++ {
		g := newTable(t)
		caster, chooser := g.Seats[0], g.Seats[1]
		spellID, targets, dest := seedPinnedPrompt(t, g, caster.ID, chooser.ID, false)

		moves := legal.EnumerateFor(g, chooser.ID)
		if len(moves) != len(targets) {
			t.Fatalf("enumerated %d answers, want one per eligible target (%d) and no decline", len(moves), len(targets))
		}
		for _, m := range moves {
			if m.Type != legal.TypeResolveChoice {
				t.Errorf("offered %q while a choice is open", m.Type)
			}
		}
		if err := actions.Dispatch(g, actions.Action{
			Type: actions.TypeResolveChoice, Player: chooser.ID, Params: moves[pick].Params,
		}); err != nil {
			t.Fatalf("the engine refused enumerated answer %d: %v", pick, err)
		}
		if len(g.PendingChoices) != 0 {
			t.Error("the prompt is still open after a legal answer")
		}
		got := g.StackMeta[spellID].Targets
		moved := 0
		for i, ref := range got {
			if ref.ID == dest {
				moved++
				continue
			}
			if ref.ID != targets[i] {
				t.Errorf("slot %d holds %v, neither its old target nor the destination", i, ref.ID)
			}
		}
		if moved != 1 {
			t.Errorf("answer %d: %d slots moved onto the destination, want exactly 1: %+v", pick, moved, got)
		}
	}
}

// A "you may" (Mizzium Meddler) enumerates the decline first, as every
// optional retarget does.
func TestOptionalPinnedRetargetEnumeratesTheDecline(t *testing.T) {
	g := newTable(t)
	caster, chooser := g.Seats[0], g.Seats[1]
	spellID, targets, _ := seedPinnedPrompt(t, g, caster.ID, chooser.ID, true)
	moves := legal.EnumerateFor(g, chooser.ID)
	if len(moves) != len(targets)+1 {
		t.Fatalf("enumerated %d answers, want %d (each target plus the decline)", len(moves), len(targets)+1)
	}
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.TypeResolveChoice, Player: chooser.ID, Params: moves[0].Params,
	}); err != nil {
		t.Fatalf("the engine refused the enumerated decline: %v", err)
	}
	got := g.StackMeta[spellID].Targets
	if got[0].ID != targets[0] || got[1].ID != targets[1] {
		t.Errorf("a decline moved a target: %+v", got)
	}
}
