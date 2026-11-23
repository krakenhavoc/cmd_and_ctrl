package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ongoing Investigation — Enchantment {1}{U} (Murders at Karlov Manor
// Commander, #5368):
//
//	"Whenever one or more creatures you control deal combat damage to a
//	 player, investigate.
//	 {1}{G}, Exile a creature card from your graveyard: Investigate. You
//	 gain 2 life."
//
// The trigger is one Clue per player your creatures hit in a combat
// damage step (CR 603.2c). The exile is a cost (Great Arashin City's),
// so a graveyard with no creature card offers nothing to pay with. The
// {G} in the cost is why its colour identity is green and blue.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68227832-2916-49b9-92bb-c09b8a0c13f5",
		Name:         "Ongoing Investigation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Ongoing Investigation — investigate", effects.Do(effects.Investigate{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, Exile a creature card from your graveyard: Investigate. You gain 2 life.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.ExileFromGraveyard(1, "a creature card", effects.MatchCreature)),
			Effect:  effects.Do(effects.Investigate{N: 1}, effects.GainLife{Amount: 2}),
		}},
	})
}
