package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tin Street Hooligan — Creature — Goblin Rogue {1}{R}, 2/1 (Guildpact,
// #5599):
//
//	"When this creature enters, if {G} was spent to cast it, destroy
//	 target artifact."
//
// Gruul Scrapper's intervening if (gptWhenEntersIfSpent) on Keldon
// Vandals' trigger (destroyFirstLegalCardTarget). With no green mana
// spent nothing triggers. ReadsManaSpent and WantsColors as on the
// Scrapper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "a33073bf-66c1-425b-9f25-748b343f33fe",
		Name:           "Tin Street Hooligan",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"G"},
		Triggered: []game.TriggeredAbility{
			effects.GptWhenEntersIfSpent("G", "Tin Street Hooligan — destroy target artifact",
				effects.TargetPermanent("target artifact", effects.Artifact()), effects.DestroyFirstLegalCardTarget),
		},
	})
}
