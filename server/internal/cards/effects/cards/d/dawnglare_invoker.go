package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dawnglare Invoker — Creature — Kor Wizard {2}{W}, 2/1 (Commander 2015, S82
// OLD #5546):
//
//	"Flying
//	 {8}: Tap all creatures target player controls."
//
// The target player is chosen on activation and every creature they control
// at resolution is tapped (Crystal Fragments' tap of everything a player
// controls); a target that has left the game is an illegal target and nothing
// is tapped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d603aef-2b92-43fb-81ba-f89953a198df",
		Name:            "Dawnglare Invoker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}: Tap all creatures target player controls.",
			Cost:    effects.ManaCost("{8}"),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return effects.B29TapAllCreaturesControlledBy(ctx, player)
			},
		}},
	})
}
