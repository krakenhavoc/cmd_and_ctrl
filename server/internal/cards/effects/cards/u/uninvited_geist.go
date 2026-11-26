package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uninvited Geist // Unimpeded Trespasser — {2}{U} Creature — Spirit 2/2 //
// Creature — Spirit 3/3 (Shadows over Innistrad, #5543):
//
//	Front: "Skulk (This creature can't be blocked by creatures with greater
//	        power.)
//	        When this creature deals combat damage to a player, transform it."
//	Back:  "This creature can't be blocked."
//
// Skulk is the engine keyword (CR 702.118). The transform turns the same
// object over in place (CR 701.27a) and is ignored if it has already turned
// over since the ability went on the stack (CR 701.27f). The back face's
// "can't be blocked" is the same restriction Bilbo, Luckwearer carries.
//
// No simplification.
func init() {
	const oracle = "5c975dbb-aabb-43fe-877c-c3950019a584"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Uninvited Geist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"skulk"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Uninvited Geist — transform it", effects.Do(effects.TransformThis{})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Unimpeded Trespasser",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
	})
}
