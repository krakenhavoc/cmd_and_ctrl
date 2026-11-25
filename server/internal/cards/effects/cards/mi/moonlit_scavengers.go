package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moonlit Scavengers — Creature — Merfolk Rogue {5}{U}, 4/5 (Throne of Eldraine,
// S82 old sets):
//
//	"When this creature enters, if you control an artifact or enchantment,
//	 return target creature an opponent controls to its owner's hand."
//
// An intervening "if" (CR 603.4): checked as it enters, so no target is chosen
// without an artifact or enchantment, and again as it resolves. The target is
// chosen as the trigger goes on the stack; a target that has gone is not
// returned (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8403ea46-e6ac-458c-9c17-4c8a81f21659",
		Name:         "Moonlit Scavengers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InterveningIfYou("you control an artifact or enchantment",
					func(g *game.Game, you uuid.UUID) bool {
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == you && (c.IsArtifact() || c.IsEnchantment()) {
								return true
							}
						}
						return false
					},
					effects.WhenThisEnters("Moonlit Scavengers — return target creature an opponent controls to its owner's hand", effects.RetractionHelixBounceTarget)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
