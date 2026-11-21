package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// yshtola_deck_sweep_b_test.go — the second slice of the "Y'shtola,
// Night's Blessed" deck request (#3657): Alisaie and Alphinaud
// Leveilleur, the job-select Equipment (Reaper's Scythe, Sage's
// Nouliths, Astrologian's Planisphere), Champions from Beyond, Eye of
// Nidhogg and Observed Stasis.

const (
	YsBAlisaieOracle    = "4bd1d079-314e-4fb5-9b43-c9c83e5528bc"
	YsBAlphinaudOracle  = "d16783a4-9980-451b-9af0-33866980bc8d"
	YsBScytheOracle     = "b0f7c2a9-914a-48c1-871c-6f72946d61d8"
	YsBNoulithsOracle   = "97248734-6e99-42a2-8f11-f3337b25f8ab"
	YsBPlanisphereOracl = "4dcd0583-cee0-46a6-ba43-93bca5eb0428"
	YsBChampionsOracle  = "57e4a1c0-4218-4a07-88e7-d6083884c2b6"
	YsBEyeOracle        = "c288f184-0beb-4cc9-9b0a-5f3c4b9e70d7"
	YsBStasisOracle     = "0c729806-3a23-491e-847f-23a394d27291"
)

// ysBHero returns the Hero token a job-select Equipment made and
// whether it is the one the Equipment is attached to.
func YsBHero(t *testing.T, g *game.Game, equipment uuid.UUID) uuid.UUID {
	t.Helper()
	eq, ok := BattlefieldCard(g, equipment)
	if !ok {
		t.Fatal("the Equipment is not on the battlefield")
	}
	if eq.AttachedTo.Kind != game.TargetCard {
		t.Fatal("job select did not attach the Equipment to the Hero")
	}
	hero, ok := BattlefieldCard(g, eq.AttachedTo.ID)
	if !ok || hero.Name != "Hero" {
		t.Fatalf("attached to %q, want a Hero token", hero.Name)
	}
	return hero.InstanceID
}
