package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const instillFurorGrant = "instill-furor/sacrifice-unless-attacked"

// Instill Furor — Enchantment — Aura {1}{R} (Ravnica: City of Guilds, S82 OLD
// batch 581, #5600):
//
//	"Enchant creature
//	 Enchanted creature has 'At the beginning of your end step, sacrifice this
//	 creature unless it attacked this turn.'"
//
// A granted triggered ability (Withercrown's shape): "your" is the creature's
// controller, so a stolen creature is sacrificed at its new controller's end
// step. The attack record is per object (CR 400.7): a creature that was put
// onto the battlefield attacking, or that came back this turn, has not
// attacked. The trigger checks as it resolves, so a creature that left or
// changed identity meanwhile is not sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5726c4b1-0613-4724-b1fc-2e7f398bf1af",
		Name:         "Instill Furor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key: instillFurorGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourEndStep("Instill Furor — sacrifice this creature unless it attacked this turn", instillFurorEndStep),
			},
			Text: "At the beginning of your end step, sacrifice this creature unless it attacked this turn.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(instillFurorGrant)},
	})
}

func instillFurorEndStep(g *game.Game, item *game.StackItem) error {
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) || g.AttackedThisTurn(item.SourceCardID) {
		return nil
	}
	return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
}
