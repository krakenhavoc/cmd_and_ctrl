package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// airbend_test.go — S22: the three airbend cards the engine can now
// express. Aang, the Last Airbender is the mechanic's proof; Appa
// adds "cast a spell from exile" and Monk Gyatso adds "becomes the
// target".

const (
	AangTheLastAirbenderOracle  = "70564c3a-858f-498e-8b92-acb3ca54ae7e"
	AppaSteadfastGuardianOracle = "03c141ca-11e4-4927-a6bd-980ee1203c73"
	MonkGyatsoOracle            = "ff92fa60-f0fe-496e-8155-d9d6f5af651b"
)

// airbendGrantOn reads the permission covering a card in exile.
func AirbendGrantOn(g *game.Game, id uuid.UUID) (*game.CastPermission, bool) {
	for _, c := range g.Exile.Cards {
		if c.InstanceID != id {
			continue
		}
		if perm := g.CastPermissionOnCardByIDForEffect(id); perm != nil {
			return perm, true
		}
		return &game.CastPermission{}, true
	}
	return nil, false
}

// assertAirbent checks the shape of a finished airbend: the card is
// in exile, and its OWNER may cast it for {2} with no expiry.
func AssertAirbent(t *testing.T, g *game.Game, id, owner uuid.UUID) {
	t.Helper()
	if g.Battlefield.Contains(id) {
		t.Fatalf("the permanent is still on the battlefield")
	}
	grant, ok := AirbendGrantOn(g, id)
	if !ok {
		t.Fatalf("the card never reached exile")
	}
	if grant.Player != owner {
		t.Errorf("grant holder = %v, want the card's owner %v", grant.Player, owner)
	}
	if grant.Duration.Kind != game.WhileInZone {
		t.Errorf("airbend grant expires; it should last while the card is exiled")
	}
	if grant.Cost != effects.AirbendCost {
		t.Errorf("cost override = %q, want %q", grant.Cost, effects.AirbendCost)
	}
	if !grant.CastOnly {
		t.Errorf("airbend says CAST, so the grant should be cast-only")
	}
}

// latestTriggerPromptFor returns the most recent trigger prompt
// addressed to chooser, or nil.
func LatestTriggerPromptFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			return c
		}
	}
	return nil
}
