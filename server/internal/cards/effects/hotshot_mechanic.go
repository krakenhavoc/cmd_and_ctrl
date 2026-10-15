package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hotshot Mechanic — {W} Artifact Creature — Fox Pilot 2/1:
//
//	"This creature crews Vehicles as though its power were 2 greater."
//
// CrewsAsThoughPowerGreater for crew only (ADR 0186 Decision 4): the
// Mechanic counts as 4 toward a crew cost and its real power stays 2. It
// does NOT count extra for a saddle cost, which is why the use is
// CostPowerCrew and not both bits.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8576201b-186d-4e28-97d4-deb344f14436",
		Name:         "Hotshot Mechanic",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CrewsAsThoughPowerGreater(2, game.CostPowerCrew),
		},
	})
}
