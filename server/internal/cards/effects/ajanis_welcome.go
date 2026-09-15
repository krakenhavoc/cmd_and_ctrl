package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ajani's Welcome — Enchantment {W} (EDHREC rank 2580):
//
//	"Whenever a creature you control enters, you gain 1 life."
//
// The one-mana Soul Warden that only watches your side. The
// condition is enteredUnderYourControl narrowed to creatures (post-
// layer type, so an animated artifact entering as a creature counts);
// one trigger per creature, so a mass token creation gains once per
// token, as printed. Lifegain fires the usual payoffs (Ajani's
// Pridemate is the printed pairing).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a782bf9-4051-4613-8852-33b0d85a0edd",
		Name:         "Ajani's Welcome",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl, "Ajani's Welcome — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}
