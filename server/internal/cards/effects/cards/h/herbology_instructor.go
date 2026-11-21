package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herbology Instructor // Malady Invoker — a Phyrexian transformer (March of the Machine):
//
//	Herbology Instructor {1}{G}, Creature — Treefolk Druid, 1/3
//	  "When this creature enters, you gain 3 life.
//	   {6}{B/P}: Transform this creature. Activate only as a sorcery."
//
//	Malady Invoker, Creature — Phyrexian Treefolk, 3/3
//	  "When this creature transforms into Malady Invoker, target creature
//	   an opponent controls gets -0/-X until end of turn, where X is this
//	   creature's power."
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with B or 2 life.
//
// No simplification.
func init() {
	const oracle = "bf0ff4b9-3543-4cfa-bfcd-f9504d27f00f"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Herbology Instructor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Herbology Instructor — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{6}{B/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{6}{B/P}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Malady Invoker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventTransform, effects.ThisTransformedInto(1),
				"Malady Invoker — target creature an opponent controls gets -0/-X until end of turn, where X is this creature's power",
				effects.MaladyInvokerShrink),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
