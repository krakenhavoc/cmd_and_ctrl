package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        effects.AangAtTheCrossroadsOracle + "#1",
		Name:            "Aang, Destined Savior",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsLand() && target.IsCreature()
			}, "vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat("Aang, Destined Savior — earthbend 2", effects.AvatarB22Earthbend(2)),
				effects.EarthbendTargets()),
		},
	})
}
