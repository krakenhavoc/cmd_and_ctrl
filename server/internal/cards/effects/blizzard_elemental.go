package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "dd489f30-d1cf-4236-8390-9c21cd8659e0",
		Name:            "Blizzard Elemental",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			UntapSelf: true,
			Label:     "{3}{U}: Untap this creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      ManaCost("{3}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
