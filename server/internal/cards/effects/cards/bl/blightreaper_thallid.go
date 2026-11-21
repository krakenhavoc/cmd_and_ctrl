package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blightreaper Thallid // Blightsower Thallid — a Phyrexian transformer (March of the Machine):
//
//	Blightreaper Thallid {1}{B}, Creature — Fungus, 2/2
//	  "{3}{G/P}: Transform this creature. Activate only as a sorcery."
//
//	Blightsower Thallid, Creature — Phyrexian Fungus, 3/3
//	  "When this creature transforms into Blightsower Thallid or dies,
//	   create a 1/1 green Phyrexian Saproling creature token."
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with G or 2 life.
//
// No simplification.
func init() {
	const oracle = "ea405cb4-a21b-4a26-98f3-251e65ea2aef"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Blightreaper Thallid",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{G/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{G/P}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Blightsower Thallid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.OnAny([]game.EventKind{game.EventTransform, game.EventLTB}, effects.BlightsowerThallidCondition,
				"Blightsower Thallid — create a 1/1 green Phyrexian Saproling creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Phyrexian Saproling"), N: 1})), game.Purpose{Tokens: 1}),
		},
	})
}
