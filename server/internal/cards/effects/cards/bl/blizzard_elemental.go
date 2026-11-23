package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blizzard Elemental — Creature — Elemental {5}{U}{U}, 5/5 (Urza's
// Destiny, #5621):
//
//	"Flying
//	 {3}{U}: Untap this creature."
//
// Depth Charge Colossus' untap row at {3}{U} on a 5/5 flier: it can
// attack and still block, or untap after a tap cost. The cost has no
// {T}, so it can be activated the turn the Elemental arrives, any
// number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dd489f30-d1cf-4236-8390-9c21cd8659e0",
		Name:            "Blizzard Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			UntapSelf: true,
			Label:     "{3}{U}: Untap this creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{3}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
