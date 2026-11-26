package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Jailbreaker — Creature — Ogre Rogue {3}{B}, 4/4 (Return to Ravnica,
// #5566):
//
//	"Defender
//	 This creature can attack as though it didn't have defender as long as you
//	 control a Gate."
//
// Backstreet Bruiser's conditional defender waiver (CR 702.3b): the keyword
// stays and the waiver is read live at declaration, with Garrison Sergeant's
// "as long as you control a Gate".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "08d27af8-7178-4694-8621-fabac54c8f0a",
		Name:            "Ogre Jailbreaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{effects.AvB18CanAttackAsThoughNoDefenderWhile(
			"This creature can attack as though it didn't have defender as long as you control a Gate.",
			func(g *game.Game, source *game.Card) bool {
				return len(effects.PermanentsControlledByMatching(g, source.Controller, effects.HasSubtype("Gate"))) > 0
			})},
	})
}
