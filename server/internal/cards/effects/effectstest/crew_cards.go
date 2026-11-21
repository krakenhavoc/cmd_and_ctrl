package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// crew_cards_test.go — ADR 0186 slice 11 (#3553): the Vehicles and Pilots
// that read the crew record or the counted-power rule, one proof per card
// for its crew-specific clause.

const (
	TurtleVanOracle          = "aeb9a2e6-ad11-4e2c-81d0-caf279d31ece"
	GetawayCarOracle         = "954b30b2-b88c-4558-b06c-4518b49d7b20"
	SchoonerOracle           = "7f82bae3-feaa-4967-a843-32a4f72806d5"
	GoldenArgosyOracle       = "b636a16d-f5ca-443d-8e6f-a2c240700c36"
	LocomotiveOracle         = "c68c2a98-9a6d-4199-af3a-78eb92beda48"
	VeteranMotoristOracle    = "3e6f51ae-7239-4fcc-919f-aa2437e8ef15"
	GearshiftAceOracle       = "dde61c4c-1815-45b3-9e68-3097c28c363c"
	SpeedwayFanaticOracle    = "5e7e1b23-ac4b-4f95-9388-feb099d6779f"
	HotshotMechanicOracle    = "8576201b-186d-4e28-97d4-deb344f14436"
	GiantOxOracle            = "b42bdf54-9055-405c-87e6-0912dbe2bd68"
	ShorikaiOracle           = "e2962883-6daf-4a84-af73-917afc73092c"
	ProdigysPrototypeOracle  = "405a54b3-fb9f-4503-a4fc-3ba6a5e58205"
	BornToDriveOracle        = "7ddd84d2-1a6d-43f5-a42d-3c66eee5f7e5"
	CrewCardsVehicleTypeLine = "Artifact — Vehicle"
)

// crewScene is the common table: the active player's crewed Vehicle
// attacking, with its attack trigger waiting.
type CrewScene struct {
	G       *game.Game
	Me, Opp *game.Player
}

func NewCrewScene(t *testing.T) CrewScene {
	t.Helper()
	g := NewCatalogGame(t)
	return CrewScene{G: g, Me: g.Seats[g.Turn.ActiveSeat], Opp: g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]}
}

// pushVehicle is a catalog Vehicle with its printed P/T, past summoning
// sickness, carried through the timestamping path.
func (s CrewScene) PushVehicle(name, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(s.G, game.Card{
		InstanceID: uuid.New(), Name: name, OracleID: oracle, TypeLine: CrewCardsVehicleTypeLine,
		Power: power, Toughness: toughness, Owner: s.Me.ID, Controller: s.Me.ID,
	})
}

func (s CrewScene) PushCreature(name, typeLine string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(s.G, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: s.Me.ID, Controller: s.Me.ID,
	})
}

// crewAndResolve crews `vehicle` (ability index `idx`) with the creatures
// and lets the ability resolve.
func (s CrewScene) CrewAndResolve(t *testing.T, vehicle uuid.UUID, idx int, crewers ...uuid.UUID) {
	t.Helper()
	ToMain(t, s.G)
	if err := s.G.ActivateCatalogAbility(s.Me.ID, vehicle, idx, game.ActivateAbilityParams{CrewIDs: crewers}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	PassPriorityAroundTable(t, s.G)
	if c, ok := BattlefieldCardByID(s.G, vehicle); !ok || !c.IsCreature() {
		t.Fatal("the crewed Vehicle is not a creature")
	}
}

func RequireFull(t *testing.T, oracle, name string) {
	t.Helper()
	spec, ok := effects.Lookup(oracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("%s: want complete with no caveat, got %+v", name, spec)
	}
}
