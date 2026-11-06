package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aang, Destined Savior — Legendary Creature — Avatar Ally, the back face
// of Aang, at the Crossroads (aang_at_the_crossroads.go), reached by its
// delayed transform:
//
//	"Flying
//	 Land creatures you control have vigilance.
//	 At the beginning of combat on your turn, earthbend 2."
//
// The grant is a layer-6 keyword static (CR 613.1f) over permanents you
// control that are both lands and creatures as the layers stand, so an
// earthbent land has vigilance the turn it is animated. The combat
// trigger earthbends 2 ("target land you control").
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        aangAtTheCrossroadsOracle + "#1",
		Name:            "Aang, Destined Savior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsLand() && target.IsCreature()
			}, "vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat("Aang, Destined Savior — earthbend 2", avatarB22Earthbend(2)),
				EarthbendTargets()),
		},
	})
}
