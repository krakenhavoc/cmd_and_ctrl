package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Conclave Evangelist — Creature — Elephant Cleric {3}{G/W}{G/W}, 4/4
// (Ravnica: Clue Edition):
//
//	"Myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and
//	 attacking that player or a planeswalker they control. Exile the
//	 tokens at end of combat.)
//	 Whenever this creature deals combat damage to a player, create a token
//	 that's a copy of this creature."
//
// Myriad is the shared trigger Chittering Dispatcher carries (CR 702.116a).
// The second ability is Giant Adephage's copy: a token copy of the source,
// from its last-known information if it has left (CR 608.2h, 707.2), under
// the trigger's controller. A copy made off a myriad token is a permanent
// copy, since only myriad's own tokens are exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13c81fc5-dd35-481d-8e32-4d404fb0e4ac",
		Name:         "Conclave Evangelist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			GrantedMyriad("Conclave Evangelist"),
			WheneverThisDealsCombatDamageToAPlayer("Conclave Evangelist — create a token that's a copy of this creature",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
