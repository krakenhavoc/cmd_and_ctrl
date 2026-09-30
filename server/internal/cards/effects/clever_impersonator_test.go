package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const cleverImpersonatorOracle = "d4ec0df9-a4a3-48cf-a0aa-b1aea5c49142"

// TestCleverImpersonatorCopiesANonlandNoncreaturePermanent — wider
// than Clone's creature-only candidate set: Clever Impersonator can
// become a copy of an artifact.
func TestCleverImpersonatorCopiesANonlandNoncreaturePermanent(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	ring := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Sol Ring",
		TypeLine:   "Artifact",
		OracleID:   "oracle-Sol Ring",
		Owner:      active.ID,
		Controller: active.ID,
	})

	impID := castCatalogSpell(t, g, "Clever Impersonator", "Creature — Shapeshifter", cleverImpersonatorOracle, nil)
	resolveWithCopyChoice(t, g, ring)

	got := copyBattlefieldCard(t, g, impID)
	if got.Name != "Sol Ring" {
		t.Errorf("name = %q, want %q", got.Name, "Sol Ring")
	}
	if !got.IsArtifact() || got.IsCreature() {
		t.Errorf("Clever Impersonator should copy Sol Ring's types (artifact, not creature)")
	}
	if !got.IsCopy() {
		t.Error("does not report as a copy")
	}
}

// TestCleverImpersonatorCannotCopyALand — with only a land on the
// battlefield there is nothing to copy, so no prompt is offered and
// Clever Impersonator stays its own printed self rather than copying
// the land.
func TestCleverImpersonatorCannotCopyALand(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Forest",
		TypeLine:   "Basic Land — Forest",
		OracleID:   "oracle-Forest",
		Owner:      active.ID,
		Controller: active.ID,
	})

	impID := castCatalogSpell(t, g, "Clever Impersonator", "Creature — Shapeshifter", cleverImpersonatorOracle, nil)
	passPriorityAroundTable(t, g)

	if copyPrompt(g) != nil {
		t.Error("a land is not a legal copy candidate; no prompt should be offered")
	}
	got := copyBattlefieldCard(t, g, impID)
	if got.Name != "Clever Impersonator" || got.IsCopy() {
		t.Errorf("with nothing to copy, Clever Impersonator should stay itself: %+v", got)
	}
}
