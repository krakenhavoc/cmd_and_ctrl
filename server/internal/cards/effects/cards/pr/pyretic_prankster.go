package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyretic Prankster // Glistening Goremonger — a Phyrexian transformer (March of the Machine):
//
//	Pyretic Prankster {1}{R}, Creature — Devil, 2/1
//	  "{3}{B/P}: Transform this creature. Activate only as a sorcery."
//
//	Glistening Goremonger, Creature — Phyrexian Devil, 3/2
//	  "When this creature dies, each opponent sacrifices an artifact or
//	   creature of their choice."
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with B or 2 life.
//
// No simplification.
func init() {
	const oracle = "f810e185-4498-483b-ab37-0c583dae0eaf"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Pyretic Prankster",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{B/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{B/P}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Glistening Goremonger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WhenThisDies("Glistening Goremonger — each opponent sacrifices an artifact or creature of their choice",
				effects.Do(effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Or(effects.Artifact(), effects.Creature()), Label: "an artifact or creature"}))),
		},
	})
}
