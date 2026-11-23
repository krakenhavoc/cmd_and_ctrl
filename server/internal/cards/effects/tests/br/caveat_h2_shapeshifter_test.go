package testsbr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// #3172 cav-h2: Unstable Shapeshifter copies the LAST-KNOWN copiable values
// of a creature that left before its trigger resolved. A Clone that entered
// as a Grizzly Bears and was destroyed in response leaves the Shapeshifter a
// Grizzly Bears, not a Clone.
func TestUnstableShapeshifterCopiesWhatAClonedCreatureWasCopyingAfterItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shifter := effectstest.DcSeed(g, me.ID, "Unstable Shapeshifter", "Creature — Shapeshifter", effectstest.OracleUnstableShapeshifter, 0, 1)
	bears := effectstest.DcSeed(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)

	cloneID := effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)
	for i := 0; i < 16 && effectstest.CopyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := effectstest.CopyPrompt(g)
	if prompt == nil {
		t.Fatal("no copy-target prompt")
	}
	if err := g.ResolveCopyTarget(prompt.ID, prompt.Chooser, bears); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	if effectstest.TriggerOnStack(g, shifter) == nil {
		t.Fatal("the Shapeshifter's trigger is not on the stack")
	}
	// In response, the Clone leaves.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(cloneID) })
	if g.Battlefield.Contains(cloneID) {
		t.Fatal("the Clone is still on the battlefield")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.DcCard(t, g, shifter); c.Name != "Grizzly Bears" {
		t.Errorf("the Shapeshifter became %q, want what the Clone was copying (Grizzly Bears)", c.Name)
	}
}
