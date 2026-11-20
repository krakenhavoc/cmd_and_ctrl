package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// War Priest of Thune — Creature — Human Cleric {1}{W}, 2/2 (Magic 2011,
// #5579):
//
//	"When this creature enters, you may destroy target enchantment."
//
// Uktabi Orangutan's targeted enters trigger over an enchantment, made
// optional the way Boneclad Necromancer's is: the target is chosen as the
// trigger goes on the stack (CR 603.3d) and the "you may" is asked as it
// resolves (CR 603.5), so the controller can spare their own enchantment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e44fd5c-5c7f-4ca8-90cd-34015047eb26",
		Name:         "War Priest of Thune",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(WhenThisEnters("War Priest of Thune — destroy target enchantment", destroyFirstLegalCardTarget),
				TargetPermanent("target enchantment", Enchantment())),
				"War Priest of Thune — destroy target enchantment?"),
		},
	})
}
