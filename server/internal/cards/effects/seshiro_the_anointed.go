package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seshiro the Anointed — Legendary Creature — Snake Monk {4}{G}{G}, 3/4
// (Champions of Kamigawa, #5603):
//
//	"Other Snake creatures you control get +2/+2.
//	 Whenever a Snake you control deals combat damage to a player, you
//	 may draw a card."
//
// The anthem is the tribal lord row (Diregraf Captain's), Seshiro
// excluded. The draw triggers once per Snake that deals combat damage to
// a player, Seshiro included, and is a CR 603.5 "you may" asked as it
// resolves (Bident of Thassa's optional draw).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "54fc373e-157f-4157-b7dd-74d689f4c0f9",
		Name:         "Seshiro the Anointed",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Snake"}, Others: true, YoursOnly: true}, 2, 2),
		},
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventDealDamage, chkB1SnakeDealtCombatDamageToAPlayer,
				"Seshiro the Anointed — draw a card", Do(DrawCards{N: 1})),
				"Seshiro the Anointed — draw a card?", 1),
		},
	})
}
