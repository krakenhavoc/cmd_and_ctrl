package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonded Herdbeast // Plated Kilnbeast — a Phyrexian transformer (March of the Machine):
//
//	Bonded Herdbeast {4}{G}, Creature — Beast, 4/5
//	  "{4}{R/P}: Transform this creature. Activate only as a sorcery."
//
//	Plated Kilnbeast, Creature — Phyrexian Beast, 7/5
//	  "Menace"
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with R or 2 life.
//
// No simplification.
func init() {
	const oracle = "b06067e7-330e-47de-be59-c2874eedba1d"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Bonded Herdbeast",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}{R/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{4}{R/P}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Plated Kilnbeast",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
