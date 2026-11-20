package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	ThreeVisitsOracle         = "1b882a0e-0ede-4d1a-bd1a-9b7cffbcde8e"
	SkyshroudClaimOracle      = "376c9d3f-21d3-4251-bb6a-026fa9e1b0e1"
	ExplosiveVegetationOracle = "a0abd957-01a0-4aa9-8fc8-0e6840d21606"
	SakuraTribeElderOracle    = "e3afc704-220f-498f-9eaa-0821b17dc24c"
	BurnishedHartOracle       = "893fed41-c144-433f-af88-bc7d419b7fb3"
	WayfarersBaubleOracle     = "31f15274-301b-47c5-ba19-0ced04520878"
	PalladiumMyrOracle        = "7b0767b8-b504-456e-93bd-218502f73b3d"
	NaturalizeOracle          = "bdb3ca68-ec1f-4e16-81cc-d23f8f52c728"
	MortifyOracle             = "faa01ed1-ccfa-4e58-951f-cd81f9068027"
	PutrefyOracle             = "9b271430-f53d-42d6-a547-2f286dd9bcb6"
	DesparkOracle             = "bd16434d-55ea-4c5a-a9ef-752971a4af16"
	VandalblastOracle         = "3567c3c8-b3c7-45b7-935b-b1fdbc973720"
	CyclonicRiftOracle        = "d75b9c82-1b49-4c3e-a1b5-aeef57d6644b"
	MercilessEvictionOracle   = "3c8d4999-e18b-48d8-8ed9-f2feaa38300d"
)

// seedBasicsInLibrary pushes n basics to the bottom of p's library so
// the sandbox picker finds them deterministically.
func SeedBasicsInLibrary(p *game.Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, PushLibraryCardForTest(p, game.Card{
			Name: "Forest", TypeLine: "Basic Land — Forest",
		}))
	}
	return ids
}
