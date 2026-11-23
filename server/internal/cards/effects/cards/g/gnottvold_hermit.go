package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnottvold Hermit // Chrome Host Hulk — a Phyrexian transformer (March of the Machine):
//
//	Gnottvold Hermit {3}{G}, Creature — Troll, 4/4
//	  "{5}{U/P}: Transform this creature. Activate only as a sorcery."
//
//	Chrome Host Hulk, Creature — Phyrexian Troll, 5/5
//	  "Whenever this creature attacks, up to one other target creature has
//	   base power and toughness 5/5 until end of turn."
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with U or 2 life.
//
// No simplification.
func init() {
	const oracle = "3bafc004-1c90-4939-9dcb-72667c170e39"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Gnottvold Hermit",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{5}{U/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{5}{U/P}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Chrome Host Hulk",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Chrome Host Hulk — up to one other target creature has base power and toughness 5/5 until end of turn",
				effects.ChromeHostHulkReshape),
				effects.Another(effects.UpToOneTargetCreature("up to one other target creature"))),
		},
	})
}
