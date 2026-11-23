package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cadaverous Knight — Creature — Zombie Knight {2}{B}, 2/2 (Mirage,
// #5633):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 {1}{B}{B}: Regenerate this creature."
//
// Flanking is flanking.go's trigger with the declared keyword (#5757,
// CR 702.25a). The regeneration is Restless Dead's row at {1}{B}{B}:
// Regenerate on the ability's own source, a shield that replaces the
// next destruction this turn (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "250a027f-2b3e-4eaf-899a-a9da8be20070",
		Name:            "Cadaverous Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{effects.RetiredFlankingRow("Cadaverous Knight")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}{B}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{B}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
