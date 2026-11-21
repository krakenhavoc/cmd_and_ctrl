package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// crewOnly activates ability 0 (the crew ability) and leaves it on the
// stack.
func (s CrewScene) CrewOnly(t *testing.T, vehicle uuid.UUID, crewers ...uuid.UUID) {
	t.Helper()
	ToMain(t, s.G)
	if err := s.G.ActivateCatalogAbility(s.Me.ID, vehicle, 0, game.ActivateAbilityParams{CrewIDs: crewers}); err != nil {
		t.Fatalf("crew: %v", err)
	}
}
