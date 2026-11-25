package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sky Theater Strix — Creature — Bird {1}{U}, 1/2 (War of the Spark, #5520):
//
//	"Flying
//	 Whenever you cast a noncreature spell, this creature gets +1/+0 until
//	 end of turn."
//
// Not prowess (no +1/+1): its own triggered ability on a noncreature spell
// you cast, pumping the Strix if it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "99e96711-90ea-4f6b-a92d-d8ddd06e8750",
		Name:            "Sky Theater Strix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou,
				"Sky Theater Strix — this creature gets +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Sky Theater Strix — +1/+0"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
