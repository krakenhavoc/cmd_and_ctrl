package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// U.S.S. Defiant, Tough Little Ship — Legendary Artifact — Spacecraft
// {6}, 7/4 (Star Trek Commander, #4974):
//
//	"When U.S.S. Defiant enters, destroy up to one target artifact or
//	 enchantment.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 5+.)
//	 5+ | Flying, hexproof"
//
// Hatut Zeraze Strike Force's enters trigger (up to one target artifact
// or enchantment, destroyed if still legal) on Galvanizing Sawship's
// station body: at five charge counters it is a 7/4 artifact creature
// with flying and hexproof.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87b8fef3-f40e-44a9-afd5-31d3917b1d59",
		Name:         "U.S.S. Defiant, Tough Little Ship",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(5, 7, 4),
			effects.ThresholdKeywords(5, "flying", "hexproof"),
		),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("U.S.S. Defiant — destroy up to one target artifact or enchantment", func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyEachLegalTarget(item, effects.NewContext(g, item))
			}),
				effects.TargetPermanent("up to one target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())).WithCount(0, 1)),
		},
	})
}
