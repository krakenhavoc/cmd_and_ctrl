package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ravenous Baboons — Creature — Monkey {3}{R}, 2/2 (Exodus, #5626):
//
//	"When this creature enters, destroy target nonbasic land."
//
// A mandatory enters trigger aimed at any nonbasic land, yours
// included (CR 603.3d: with none on the battlefield it is removed).
// NonbasicLand reads the basic supertype, not the land type, so a
// Sacred Foundry is a legal target and a Snow-Covered Swamp is not. The
// shared destroyFirstLegalTarget re-checks the target at resolution
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d9bd0a5-3ba0-44e5-9cb3-966f6b9dc14a",
		Name:         "Ravenous Baboons",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Ravenous Baboons — destroy target nonbasic land", destroyFirstLegalTarget),
				TargetPermanent("target nonbasic land", NonbasicLand())),
		},
	})
}
