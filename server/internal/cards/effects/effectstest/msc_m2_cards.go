package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mscM2PickTargets answers the open pick_target prompt with cards.
func MscM2PickTargets(t *testing.T, g *game.Game, ids ...uuid.UUID) {
	t.Helper()
	ch := PickTargetPrompt(t, g)
	refs := make([]game.TargetRef, len(ids))
	for i, id := range ids {
		refs[i] = game.TargetRef{Kind: game.TargetCard, ID: id}
	}
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	PassPriorityAroundTable(t, g)
}

func MscM2Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

// --- Zuri, Warrior of Wakanda ---

func MscM2CastArtifact(t *testing.T, g *game.Game, cost, mana string) {
	t.Helper()
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Gizmo", TypeLine: "Artifact", ManaCost: cost,
		Owner: me.ID, Controller: me.ID,
	})
	AdvanceToMain(t, g)
	FloatMana(t, g, me, mana)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	PassPriorityAroundTable(t, g)
}
