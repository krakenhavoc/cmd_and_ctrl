package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vile Consumption — Enchantment {1}{U}{B} (Invasion, #5616):
//
//	"All creatures have "At the beginning of your upkeep, sacrifice this
//	 creature unless you pay 1 life.""
//
// A granted triggered ability (Plague Sliver's) on every creature at the table,
// the tribe filter left empty to match all of them. "Your upkeep" and "you" are
// the controller of the creature that has it, so each creature asks its own
// controller in their own upkeep, one trigger per creature, and the payment is
// 1 life (a player with no life to pay, or who declines, sacrifices it). A
// creature that left or came back as a new object asks nothing (CR 400.7).
//
// No simplification.
const vileConsumptionGrant = "vile-consumption/upkeep"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6a005db-e7e2-4b88-bc8a-c257bcc31cf5",
		Name:         "Vile Consumption",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: vileConsumptionGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Vile Consumption — sacrifice this creature unless you pay 1 life", vileConsumptionUpkeep),
			},
			Text: "At the beginning of your upkeep, sacrifice this creature unless you pay 1 life.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{}, vileConsumptionGrant),
		},
	})
}

func vileConsumptionUpkeep(g *game.Game, item *game.StackItem) error {
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	pay := effects.LifePayment(1)
	words, action := pay.Prompt()
	return effects.UpkeepPayUnless{
		Chooser:  item.Controller,
		Cost:     words,
		Action:   action,
		Question: "Vile Consumption — pay 1 life, or sacrifice this creature?",
		Else:     effects.UpkeepSacrificeSource,
	}.Apply(effects.NewContext(g, item))
}
