package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fear of Failed Tests — Enchantment Creature — Nightmare {4}{U}, 2/7
// (Duskmourn: House of Horror, #4466):
//
//	"Whenever this creature deals combat damage to a player, draw that
//	 many cards."
//
// Professor Hulk's trigger: the shared "deals combat damage to a player"
// condition, and "that many" is the damage the triggering event carries
// (lichsMasteryDraw), fixed as the trigger fires, so a pump or shrink after
// damage does not change it. The draw is mandatory.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34fd461b-a913-4a60-9477-ce2bd5c6fc5a",
		Name:         "Fear of Failed Tests",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Fear of Failed Tests — draw that many cards", lichsMasteryDraw),
		},
	})
}
