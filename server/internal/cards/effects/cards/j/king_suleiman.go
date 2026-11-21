package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// King Suleiman — Creature — Human Noble {1}{W}, 1/1 (Arabian Nights,
// #5641):
//
//	"{T}: Destroy target Djinn or Efreet."
//
// Royal Assassin's row with a subtype clause: any permanent that is a
// Djinn or an Efreet, whoever controls it, read off its current types
// (a changeling counts).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97f548b1-6398-4dc7-b5eb-da5a3f24ddb7",
		Name:         "King Suleiman",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target Djinn or Efreet.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target Djinn or Efreet", effects.Or(effects.HasSubtype("Djinn"), effects.HasSubtype("Efreet"))),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
