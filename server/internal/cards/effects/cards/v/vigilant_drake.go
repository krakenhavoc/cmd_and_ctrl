package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vigilant Drake — Creature — Drake {4}{U}, 3/3 (Urza's Legacy, S82 OLD b716, #5623):
//
//	"Flying
//	{2}{U}: Untap this creature."
//
// Flying is a printed keyword; the untap is Depth Charge Colossus' (no {T}, so
// it can be activated any number of times), a no-op once the Drake has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4f5eaf5a-dd52-4d48-91b5-d48763d13159",
		Name:            "Vigilant Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			UntapSelf: true,
			Label:     "{2}{U}: Untap this creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{2}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
