package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rooftop Bypass — Enchantment {1}{U}{B} (Assassin's Creed, #5948):
//
//	"Whenever one or more nontoken creatures you control deal combat damage to
//	 a player, create a 1/1 black Assassin creature token with menace. (It can't
//	 be blocked except by two or more creatures.)"
//
// The batch trigger over combat damage to a player (CR 603.2c: once per player
// hit per damage step, however many creatures connect), narrowed to nontoken
// creatures as the dealers were when the damage was dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba0d177f-c8e9-4db3-b5cf-40ed2b63e111",
		Name:         "Rooftop Bypass",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(Not(IsTokenPredicate()),
				"Rooftop Bypass — create a 1/1 black Assassin creature token with menace",
				Do(CreateToken{Template: TokenCard("1/1 black Assassin with menace"), N: 1})),
		},
	})
}
