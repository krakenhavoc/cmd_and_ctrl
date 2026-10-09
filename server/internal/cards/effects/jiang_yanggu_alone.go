package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jiang Yanggu, Alone — Legendary Creature — Human Berserker {4}{R},
// 4/4:
//
//	"Menace
//	 Whenever a creature you control attacks a player alone, discard a
//	 card, then draw a card. Then put a +1/+1 counter on that creature
//	 for each card you've discarded this turn."
//
// "Attacks a player alone" is Exalted's reading (CR 506.5, one declared
// attacker, whoever controls it) narrowed to an attack aimed at a
// player: one at a planeswalker or battle does not trigger it. The
// discard is the controller's own choice; the draw and the counters wait
// for the answer. The counters count every card discarded this turn, the
// one just pitched included, and go on the attacking creature if it is
// still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fa8b3557-b35a-4a68-ab85-cf09683f3bc5",
		Name:            "Jiang Yanggu, Alone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, rfCreatureCAttacksAPlayerAlone,
				"Jiang Yanggu, Alone — discard a card, then draw a card, then +1/+1 counters for each card discarded this turn",
				rfCreatureCJiangAloneEffect),
		},
	})
}
