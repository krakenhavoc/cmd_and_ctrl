package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b1_test.go — Planechase plane cards, batch 1 (ADR 0216 slice 4,
// #4870). Every test plays the real catalog entry face up on a Planechase
// table (planes_harness_test.go) and drives the engine: the plane's static
// and triggered abilities from the command zone, and its chaos ability on
// a chaos roll.

var (
	PlaneAmysHome         = PlaneFixture{"5b200e41-126d-4017-8e01-f50582b31af0", "Amy's Home", "Plane — Earth"}
	PlaneAntarcticBase    = PlaneFixture{"d65ff607-1ff5-42bd-8db2-f082cc2f0b5e", "Antarctic Research Base", "Plane — Earth"}
	PlaneAplanMortarium   = PlaneFixture{"7c9d6193-1d50-4035-8c85-98c384f461be", "Aplan Mortarium", "Plane — Alfava Metraxis"}
	PlaneBadWolfBay       = PlaneFixture{"706720f6-bb04-4a7b-8471-0f3d662d6eeb", "Bad Wolf Bay", "Plane — Earth"}
	PlaneVikingVillage    = PlaneFixture{"75206757-dae7-4a1e-a391-862699a100d7", "Besieged Viking Village", "Plane — Earth"}
	PlaneBowieBaseOne     = PlaneFixture{"7076bc0c-9cc7-4cd7-aba5-f2f236aab0f7", "Bowie Base One", "Plane — Mars"}
	PlaneCityOfTheDaleks  = PlaneFixture{"a2ef6b79-faeb-4183-a510-f10b14dff417", "City of the Daleks", "Plane — Skaro"}
	PlaneCoalHillSchool   = PlaneFixture{"51a53042-6baa-442a-84ec-2f16fdc12087", "Coal Hill School", "Plane — Earth"}
	PlaneDalekCare        = PlaneFixture{"2edd018b-6d82-4ff1-b9d0-5da9b628af3f", "Dalek Intensive Care", "Plane — The Dalek Asylum"}
	PlaneGardensOfRepose  = PlaneFixture{"d45692d5-19e8-45ec-92bf-b79cb3919acd", "Gardens of Tranquil Repose", "Plane — Necros"}
	PlaneHotelOfFears     = PlaneFixture{"79d3e356-9eb9-49bb-b4ba-f4f71876869e", "Hotel of Fears", "Plane — Spacecraft"}
	PlaneKerblamWarehouse = PlaneFixture{"71bff56b-733c-4d74-9e0c-85371a8b3787", "Kerblam! Warehouse", "Plane — Kandoka"}
	PlaneLakeSilencio     = PlaneFixture{"4dffc1cb-e9bd-4410-9b08-184eabd2c4da", "Lake Silencio", "Plane — Earth"}
	PlaneMondassianShip   = PlaneFixture{"9cc34a5b-550d-4d3e-b3e5-4d4630649882", "Mondassian Colony Ship", "Plane — Spacecraft"}
)

func OpponentOf(g *game.Game) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}
