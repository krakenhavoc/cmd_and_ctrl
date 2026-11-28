package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mangara's Equity — Enchantment {1}{W}{W} (Mirage, #5633):
//
//	"As this enchantment enters, choose black or red.
//	 At the beginning of your upkeep, sacrifice this enchantment unless you pay
//	 {1}{W}.
//	 Whenever a creature of the chosen color deals damage to you or a white
//	 creature you control, this enchantment deals that much damage to that
//	 creature."
//
// The chosen colour is read as the creature last was when it dealt the damage;
// the damage Mangara's Equity deals is as much as the creature dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7986014-30e0-4999-ab95-b86620471419",
		Name:         "Mangara's Equity",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseOneOfColorsAsEnters(game.ColorForProtection, "Mangara's Equity", "black or red", "B", "R"),
		Triggered: []game.TriggeredAbility{
			effects.UpkeepSacrificeUnlessYouPay("Mangara's Equity", "{1}{W}"),
			effects.On(game.EventDealDamage, effects.OldB795ChosenColorCreatureHurtYouOrAWhiteCreature,
				"Mangara's Equity — it deals that much damage to that creature",
				func(g *game.Game, item *game.StackItem) error {
					ev := item.Trigger.Event
					if !effects.OnBattlefield(g, ev.Source) {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: ev.Source, Amount: ev.Amount}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
