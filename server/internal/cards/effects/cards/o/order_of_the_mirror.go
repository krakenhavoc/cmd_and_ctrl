package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Order of the Mirror",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{W/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{W/P}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Order of the Alabaster Host",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.ViciousBattleragerBlocked,
				"Order of the Alabaster Host — the blocking creature gets -1/-1 until end of turn", effects.OrderOfTheAlabasterHostShrink),
		},
	})
}
