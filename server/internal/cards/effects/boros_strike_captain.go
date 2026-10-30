package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boros Strike-Captain — Creature — Minotaur Soldier {1}{R/W}{R/W},
// 3/3:
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, exile the top card of your library. During any turn you
//	 attacked with three or more creatures, you may play that card."
//
// Battalion is not a keyword with a shape of its own: the trigger is the
// Captain's own attack declaration with two OTHER attackers in the same
// declaration, which every attacker already is by the time the first
// EventAttack goes out (attackersYouControlOtherThan). The permission is
// a conditional exile play (#3192, ADR 0066's second 2026-10-13
// amendment): it lasts for as long as the card stays exiled and is live
// only during a turn its holder attacked with three or more DIFFERENT
// creatures (game.CondAttackedWithAtLeast, counted over the whole turn,
// so a creature attacking again in a second combat counts once, and an
// attacker that has since died still counts).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30b20932-0d9a-447f-b934-1daa8c44a678",
		Name:         "Boros Strike-Captain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, battalionAttack, "Boros Strike-Captain — exile the top card of your library (battalion)", func(g *game.Game, item *game.StackItem) error {
				return ExileTopWithPermission{
					From: item.Controller, GrantTo: item.Controller, N: 1,
					WhileExiled: true,
					Condition:   &game.PermissionCondition{Key: game.CondAttackedWithAtLeast, N: 3},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
