package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Campaign of Vengeance — Enchantment {3}{W}{B}:
//
//	"Whenever a creature you control attacks, defending player loses 1
//	 life and you gain 1 life."
//
// The same trigger as Brutal Hordechief's (attackDrainTrigger): one per
// attacking creature (CR 508.3a), the defending player read off the
// attack as it is declared, so a creature attacking a planeswalker
// drains the planeswalker's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "627e9668-5e1c-4ab6-83f4-612a42d0afce",
		Name:         "Campaign of Vengeance",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{attackDrainTrigger("Campaign of Vengeance")},
	})
}
