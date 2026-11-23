package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sejiri Merfolk — Creature — Merfolk Soldier {1}{U}, 2/1 (Worldwake, #5582):
//
//	"As long as you control a Plains, this creature has first strike and
//	 lifelink. (Damage dealt by a creature with lifelink also causes its
//	 controller to gain that much life.)"
//
// Two layer-6 grants to itself while its controller controls a Plains,
// Kargan Dragonrider's shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74231a46-7e2d-497f-ae66-b40f8989551c",
		Name:         "Sejiri Merfolk",
		Completeness: effects.CompletenessFull,
		Static:       sejiriMerfolkGrants(),
	})
}

func sejiriMerfolkGrants() []game.StaticAbility {
	whilePlains := effects.SelfWhileYouControlA(effects.Subtype("Plains"))
	firstStrike := effects.KeywordGrant(whilePlains, "first strike")
	firstStrike.Label = "As long as you control a Plains, this creature has first strike."
	lifelink := effects.KeywordGrant(whilePlains, "lifelink")
	lifelink.Label = "As long as you control a Plains, this creature has lifelink."
	return []game.StaticAbility{firstStrike, lifelink}
}
