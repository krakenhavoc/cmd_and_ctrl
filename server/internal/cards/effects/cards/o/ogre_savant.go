package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Savant — Creature — Ogre Wizard {4}{R}, 3/2 (Guildpact, #5599):
//
//	"When this creature enters, if {U} was spent to cast it, return
//	 target creature to its owner's hand."
//
// Gruul Scrapper's intervening if (gptWhenEntersIfSpent) on Kiri-Onna's
// bounce (bounceTheTargetToHand). With no blue mana spent nothing
// triggers. Any creature may be the target, the Savant itself included.
// ReadsManaSpent and WantsColors as on the Scrapper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "2ce12817-a7b5-475a-87d0-718d77349a6d",
		Name:           "Ogre Savant",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"U"},
		Triggered: []game.TriggeredAbility{
			effects.GptWhenEntersIfSpent("U", "Ogre Savant — return target creature to its owner's hand",
				effects.TargetCreature("target creature"), effects.BounceTheTargetToHand),
		},
	})
}
