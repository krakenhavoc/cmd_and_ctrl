package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rimeshield Frost Giant — Creature — Giant Warrior {3}{U}{U}, 4/5:
//
//	"Ward {3}"
//
// A Kaldheim common, and registered for exactly that reason: ward is
// the whole card. Every other ward card in the format carries a
// second ability, so this is the one place in the catalog where a
// test can assert "ward works" without a second mechanic in the
// frame — and where a future reader can see the shape of a ward
// declaration with nothing else around it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec47cf67-2580-464f-8118-7eabea5be11c",
		Name:         "Rimeshield Frost Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{3}"), "Rimeshield Frost Giant — ward {3}"),
		},
	})
}
