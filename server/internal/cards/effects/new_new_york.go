package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// New New York — Plane — New Earth (Doctor Who, #4870):
//
//	"At the beginning of combat on your turn, until end of turn,
//	 noncreature artifacts you control become 3/3 Vehicles in addition to
//	 their other types and gain flying, haste, and crew 1.
//	 Whenever chaos ensues, create a Treasure token and a 2/2 white Alien
//	 creature token."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// See planesB2BecomeVehicles. The chaos ability (CR 311.7) creates both tokens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f8cf1ab-1359-44af-a989-dd254fa6c7b5",
		Name:         "New New York",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{planesB2CrewGrant(newNewYorkCrewGrant, 1)},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("New New York — until end of turn, noncreature artifacts you control become 3/3 Vehicles and gain flying, haste, and crew 1",
				planesB2BecomeVehicles("New New York — 3/3 Vehicles with flying, haste and crew 1", newNewYorkCrewGrant,
					Not(Creature()), 3, 3, "flying", "haste")),
			WheneverChaosEnsues("New New York — create a Treasure token and a 2/2 white Alien creature token",
				Do(CreateToken{Template: TreasureToken(), N: 1}, CreateToken{Template: planesB2Alien(), N: 1})),
		},
	})
}

// newNewYorkCrewGrant is the granted bundle's key. An on-disk key.
const newNewYorkCrewGrant = "new-new-york/crew-1"
