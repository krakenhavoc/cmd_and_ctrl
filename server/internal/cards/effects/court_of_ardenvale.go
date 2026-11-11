package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Court of Ardenvale — Enchantment {2}{W}{W} (Wilds of Eldraine
// Commander, #5368):
//
//	"When this enchantment enters, you become the monarch.
//	 At the beginning of your upkeep, return target permanent card with
//	 mana value 3 or less from your graveyard to your hand. If you're the
//	 monarch, return that permanent card to the battlefield instead."
//
// The target is chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b); the destination is read then too, since "if
// you're the monarch … instead" is a clause of the effect (CR 608.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be14d3be-295e-431c-b891-590164b730dd",
		Name:         "Court of Ardenvale",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Court of Ardenvale"),
			Targeting(AtYourUpkeep("Court of Ardenvale — return target permanent card with mana value 3 or less from your graveyard to your hand, or to the battlefield if you're the monarch",
				cmdr2S48CourtOfArdenvaleUpkeep),
				TargetCardInGraveyard("target permanent card with mana value 3 or less from your graveyard",
					YouOwn(), Permanent(), ManaValueLE(3))),
		},
	})
}
