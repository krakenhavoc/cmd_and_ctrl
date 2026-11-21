package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Naomi, Pillar of Order — Legendary Creature — Human Advisor {3}{W}{B}, 4/4
// (Kamigawa: Neon Dynasty, #5493):
//
//	"Whenever Naomi enters or attacks, if you control an artifact and an
//	 enchantment, create a 2/2 white Samurai creature token with vigilance."
//
// One ability with two trigger conditions (WhenThisEntersOrAttacks, as Agrus
// Kos), with an intervening "if" (CR 603.4): checked as it triggers and again
// as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd6018de-5791-4c13-b134-38c410f04872",
		Name:         "Naomi, Pillar of Order",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.NeoB08CondArtifactAndEnchantment,
				effects.OnAny([]game.EventKind{game.EventETB, game.EventAttack}, effects.AllOf(effects.Self, effects.NeoB08WhenArtifactAndEnchantment),
					"Naomi — create a 2/2 white Samurai creature token with vigilance",
					effects.Do(effects.CreateToken{Template: effects.NeoB08SamuraiToken(), N: 1}))),
		},
	})
}
