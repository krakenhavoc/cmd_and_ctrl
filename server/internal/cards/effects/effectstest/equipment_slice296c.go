package effectstest

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// equipment_slice296c_test.go covers roadmap slice 296-c (Equipment):
// Helm of the Host, Commander's Plate, The Reaver Cleaver, Brotherhood
// Regalia, Hammer of Nazahn, Buster Sword, Tarrian's Soulcleaver,
// Conqueror's Flail, Embercleave, Caduceus Staff of Hermes, Bloodforged
// Battle-Axe, Sting the Glinting Dagger, and Illusionist's Bracers.
// The Reality Chip and Lizard Blades waited on reconfigure, which #2639
// built; they are tested in reconfigure_cards_test.go.

const (
	HelmOfTheHostOracle        = "83b43aba-bf9c-4da2-967d-9daa632e97d2"
	CommandersPlateOracle      = "cae166de-e681-40a0-83a8-3c17cf40e2fc"
	TheReaverCleaverOracle     = "37a2a31d-51e6-4c07-b4c2-b206ad40eb42"
	BrotherhoodRegaliaOracle   = "d169738e-af0b-4488-a399-3aa1d1933aa1"
	HammerOfNazahnOracle       = "e3955573-3db5-490e-903e-65e0172a9202"
	BusterSwordOracle          = "5e060d58-4d6e-425c-b7d4-727669fcce5b"
	TarriansSoulcleaverOracle  = "2ffb38ec-5852-4e91-85a5-cfccd1f23556"
	ConquerorsFlailOracle      = "9aace0d3-89e6-4254-b96f-ee3a878f2f91"
	EmbercleaveOracle          = "4d6120d6-fcce-40bc-9fc6-e1f5beb6c728"
	CaduceusOracle             = "7cea5b12-9483-4142-9efa-735305581e73"
	BloodforgedBattleAxeOracle = "81a7f559-5edd-47ec-91f7-51e5ee998ed0"
	StingOracle                = "973c49a1-8425-40b9-8bdc-2c2434222314"
	IllusionistsBracersOracle  = "1d1d78af-7982-419d-b9be-2bf4c149d97d"
)

// currentPowerFor reads Card.CurrentPower() — the layer-effective
// power PLUS any +1/+1 / -1/-1 counters (CR 613's counters step,
// which effectivePower's plain Effective().Power deliberately
// excludes; counter math is never modelled as a layer, per docs/adding-cards.md).
func CurrentPowerFor(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	var p int
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				p, found = c.CurrentPower(), true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return p
}

// --- Illusionist's Bracers -----------------------------------------------

func NewBigLibraryCatalogGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 40)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}
