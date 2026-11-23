package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Migloz, Maze Crusher — Legendary Creature — Phyrexian Beast {1}{R}{G}, 4/4:
//
//	"Migloz enters with five oil counters on it.
//	 {1}, Remove an oil counter from Migloz: It gains vigilance and menace
//	 until end of turn.
//	 {2}, Remove two oil counters from Migloz: It gets +2/+2 until end of
//	 turn.
//	 {3}, Remove three oil counters from Migloz: Destroy target artifact or
//	 enchantment."
//
// Each counter cost is paid as the ability is activated (CR 602.2b), so a
// Migloz with too few cannot be activated and can spend them in response to
// removal. The keyword grant is one layer-6 effect for both words.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85eb890e-a5e1-48f8-b813-e8424d18b28a",
		Name:         "Migloz, Maze Crusher",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.S82OneB12EntersWithOil(5, "Migloz enters with five oil counters on it")},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, Remove an oil counter from Migloz: It gains vigilance and menace until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersFromThis(effects.S82OneB12Oil, 1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.UntilEndOfTurn(effects.NewContext(g, item), item.SourceCardID, nil,
						"Migloz — vigilance and menace until end of turn",
						game.AddKeywordsMod("vigilance", "menace"))
				},
			},
			{
				Label:   "{2}, Remove two oil counters from Migloz: It gets +2/+2 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.RemoveCountersFromThis(effects.S82OneB12Oil, 2)),
				Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Migloz — +2/+2 until end of turn"),
			},
			{
				Label:   "{3}, Remove three oil counters from Migloz: Destroy target artifact or enchantment.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.RemoveCountersFromThis(effects.S82OneB12Oil, 3)),
				Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
				},
			},
		},
	})
}
