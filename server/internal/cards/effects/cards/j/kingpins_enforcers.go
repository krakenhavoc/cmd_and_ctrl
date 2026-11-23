package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kingpin's Enforcers — Creature — Human Villain {2}{B}, 2/3 (Marvel Super
// Heroes, #4455):
//
//	"Lifelink
//	 {2}{B}, Sacrifice an artifact or creature: Draw a card."
//
// Dockside Chef's sacrifice outlet (b30SacrificeAnArtifactOrCreature): the
// sacrifice is a cost paid at announcement, so the Enforcers are a legal
// payment themselves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "342df67f-44d8-4b0a-b1d7-4241966d0598",
		Name:            "Kingpin's Enforcers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice an artifact or creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.B30SacrificeAnArtifactOrCreature()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
