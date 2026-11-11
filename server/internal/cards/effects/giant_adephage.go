package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant Adephage — Creature — Insect {5}{G}{G}, 7/7:
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, create a
//	 token that's a copy of this creature."
//
// The copy is the Adephage's copiable values (CR 707.2) as it last
// existed, so one that was itself a copy copies what it copied, and the
// token — which has the same trigger — copies again the next time it
// connects. The trigger is the creature's own and one per damage event;
// it makes the token even if the Adephage has left the battlefield by
// resolution (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8e1d261d-b8d3-4fb1-922d-47789854b5b3",
		Name:            "Giant Adephage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtCombatDamageToAPlayer,
				"Giant Adephage — create a token that's a copy of this creature",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
