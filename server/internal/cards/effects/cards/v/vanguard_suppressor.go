package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vanguard Suppressor — Creature — Astartes Warrior {3}{U}, 3/2 (Warhammer
// 40,000 Commander, S82 GAP batch 195, #5948):
//
//	"Squad {2} (As an additional cost to cast this spell, you may pay {2} any
//	 number of times. When this creature enters, create that many tokens that are
//	 copies of it.)
//	 Flying
//	 Suppressing Fire — Whenever this creature deals combat damage to a player,
//	 draw a card."
//
// Squad is Wasteland Raider's optional cost and enters trigger; each copy has
// the combat damage trigger too, so each that connects draws a card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b7c48f87-bdef-4cef-a1c4-5b4d0b7a26cd",
		Name:            "Vanguard Suppressor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.SquadMana("{2}", 20)},
		Triggered: []game.TriggeredAbility{
			effects.Squad("Vanguard Suppressor"),
			effects.WheneverThisDealsCombatDamageToAPlayer("Vanguard Suppressor — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
