package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyword_cards_test.go covers the S18 sub-PR 4 vanilla keyword
// card batch. Each test pushes the catalog card onto the
// battlefield, forces a layer recompute via ReadSnapshot, and
// asserts the expected keyword strings land in
// Effective().Abilities. The actual combat behaviour (flying block
// restriction, trample overflow, etc.) is tested in
// server/internal/game/combat_test.go against manufactured
// battlefield state; these tests just prove the card files declare
// the right PrintedKeywords slot.

const (
	SerraAngelOracle         = "4b7ac066-e5c7-43e6-9e7e-2739b24a905d"
	ColossalDreadmawOracle   = "08c7db90-c0cf-4482-b7ee-bb033e5996d2"
	GiantSpiderOracle        = "e740ce2f-2134-473c-afa1-1b6d2d1e38ef"
	TyphoidRatsOracle        = "d6ee6cc1-902d-4f56-afa5-6fa4813bfbbc"
	YouthfulKnightOracle     = "ef2a24f5-ce5e-4054-843a-2cae0c66318a"
	FencingAceOracle         = "2f810936-2ba6-4c2b-84a0-ff4c1deb026b"
	LightningElementalOracle = "58aee5cb-7b88-446e-ab10-9f83c10d7227"
	WallOfStoneOracle        = "cd4cadb4-3156-49bd-b36e-12ba5c85938b"
)

func AssertKeywords(t *testing.T, g *game.Game, cardID uuid.UUID, want ...string) {
	t.Helper()
	got := EffectiveAbilities(t, g, cardID)
	for _, kw := range want {
		if !ContainsString(got, kw) {
			t.Errorf("card %s missing %q in effective abilities %v", cardID, kw, got)
		}
	}
}
