package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Order of the Mirror // Order of the Alabaster Host — a Phyrexian transformer (March of the Machine):
//
//	Order of the Mirror {1}{U}, Creature — Human Knight, 2/1
//	  "{3}{W/P}: Transform this creature. Activate only as a sorcery."
//
//	Order of the Alabaster Host, Creature — Phyrexian Knight, 3/3
//	  "Whenever this creature becomes blocked by a creature, the blocking
//	   creature gets -1/-1 until end of turn."
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with W or 2 life.
//
// No simplification.
func init() {
	const oracle = "d849be9d-0447-4f0b-a382-3f4ceb98f9a5"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Order of the Mirror",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{3}{W/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{3}{W/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Order of the Alabaster Host",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, viciousBattleragerBlocked,
				"Order of the Alabaster Host — the blocking creature gets -1/-1 until end of turn", orderOfTheAlabasterHostShrink),
		},
	})
}
