package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wayward Swordtooth — Creature — Dinosaur {2}{G}, 5/5 (EDHREC rank
// 1052):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 You may play an additional land on each of your turns.
//	 This creature can't attack or block unless you have the city's
//	 blessing."
//
// A three-mana 5/5 that ramps from the turn it lands and does nothing
// else until its controller's board reaches ten permanents. Three
// printed lines, three mechanisms:
//
//   - Ascend is the keyword the engine reads (citys_blessing.go in
//     game): the Swordtooth is itself an ascend permanent, so its
//     controller earns the blessing the moment they control ten.
//   - The extra land is `Spec.AdditionalLandPlays`, as on Exploration.
//   - The attack half is the same CR 508.1c restriction Goblin Goon
//     uses, with one clause about the attacker's controller (nothing
//     about the defending player); the block half is a CR 509.1b block
//     rule on the creature itself.
//
// The blessing is kept once earned (CR 702.131c), so a Swordtooth whose
// controller's board later shrinks below ten still attacks and blocks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "3875aef0-3102-4fbf-be90-e4139f7a2348",
		Name:                "Wayward Swordtooth",
		Completeness:        CompletenessFull,
		PrintedKeywords:     []string{game.KeywordAscend},
		AdditionalLandPlays: 1,
		Static:              []game.StaticAbility{CantAttackUnlessYouHaveTheCitysBlessing()},
		BlockRules:          []game.BlockRule{CantBlockUnlessYouHaveTheCitysBlessing()},
	})
}
