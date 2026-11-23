package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// isCyberman is the whole listed body, asserted through the layered
// characteristics every reader sees.
func IsCyberman(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	if !c.FaceDown || c.FaceDownKind != game.FaceDownTurned {
		t.Errorf("state = (%v, %q), want face down, turned", c.FaceDown, c.FaceDownKind)
	}
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Cyberman") {
		t.Errorf("artifact %v, creature %v, Cyberman %v — want a Cyberman artifact creature",
			c.IsArtifact(), c.IsCreature(), c.HasSubtype("Cyberman"))
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("P/T = %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	if eff := c.Effective(); eff.Name != "" || len(eff.Abilities) != 0 {
		t.Errorf("the body has a name %q or abilities %v — CR 708.2 lists neither", eff.Name, eff.Abilities)
	}
}
