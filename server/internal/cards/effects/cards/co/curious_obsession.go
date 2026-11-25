package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const curiousObsessionDraw = "curious-obsession/draw"

// Curious Obsession — Enchantment — Aura {U} (Rivals of Ixalan, Jumpstart):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has 'Whenever this creature deals
//	 combat damage to a player, you may draw a card.'
//	 At the beginning of your end step, if you didn't attack with a
//	 creature this turn, sacrifice this Aura."
//
// The draw is GRANTED to the creature (ADR 0093), so its controller is the
// one who may draw, and a creature that loses all abilities loses it. The
// sacrifice is the Aura's own trigger and "you" is the Aura's controller.
// "If you didn't attack with a creature this turn" is an intervening if
// (CR 603.4): checked as the end step begins, and again as the trigger
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22a23706-a5fe-46f3-b845-6f58d558c723",
		Name:         "Curious Obsession",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: curiousObsessionDraw,
			Triggered: []game.TriggeredAbility{
				effects.OptionalDraw(effects.WheneverThisDealsCombatDamageToAPlayer("Curious Obsession — you may draw a card",
					effects.Do(effects.DrawCards{N: 1})), "Draw a card?", 1),
			},
			Text: "Whenever this creature deals combat damage to a player, you may draw a card.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantAbilitiesToAttached(curiousObsessionDraw),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepSacrificeThisIfYouDidntAttack("Curious Obsession — sacrifice this Aura (you didn't attack with a creature this turn)"),
		},
	})
}
