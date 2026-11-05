package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sulu, Heroic Helmsman — {1}{W} Legendary Creature — Human Pilot 2/2
// (Star Trek Commander, #4974):
//
//	"Sulu stations Spacecraft and crews Vehicles as though his power were
//	 tripled."
//
// A layer 6 self-only static (crewsAndStationsAsThoughPowerTripled) that
// writes a CostPowerRule with a Multiplier of 3 for crew and for station,
// the shape of Hotshot Mechanic's "+2" (ADR 0186 Decision 4). Only the
// payment counts it (Card.CountedPower): his power stays 2 for everything
// else. CR 702.184c: a static may change what a station ability reads;
// the station of a Spacecraft reads Game.PaidTapStationPowerForEffect, and
// the station of a Planet still reads his plain power, because the card
// names Spacecraft. The station count is read as the ability resolves, or
// from his last-known information if he has left (CR 608.2h), so a pump
// in response is tripled too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9565feaf-252d-43c2-a28e-4291c9ff5872",
		Name:         "Sulu, Heroic Helmsman",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{crewsAndStationsAsThoughPowerTripled()},
	})
}
