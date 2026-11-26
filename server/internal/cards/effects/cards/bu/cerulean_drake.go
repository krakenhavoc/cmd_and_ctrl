package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cerulean Drake — Creature — Drake {1}{U}, 1/1 (Core Set 2020):
//
//	"Flying
//	 Protection from red
//	 Sacrifice this creature: Counter target spell that targets you."
//
// Flying and protection from red are printed keywords. The ability is
// Hindering Light's counter with a narrower clause: the spell must target the
// Drake's controller (a permanent they control does not do), judged as the
// ability is activated and again as it resolves (CR 608.2b). The sacrifice is
// a cost (CR 118.12), so the Drake is gone whether or not the counter lands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ecd6587c-3548-4bc1-89e4-9f7fbbc2d2ad",
		Name:            "Cerulean Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from red"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Counter target spell that targets you.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetSpell("target spell that targets you", effects.OldB158SpellTargetsYou),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				return effects.CounterTarget{StackID: item.Targets[0].ID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
