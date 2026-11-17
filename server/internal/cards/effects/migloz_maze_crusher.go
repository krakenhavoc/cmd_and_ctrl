package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "85eb890e-a5e1-48f8-b813-e8424d18b28a",
		Name:         "Migloz, Maze Crusher",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{s82OneB12EntersWithOil(5, "Migloz enters with five oil counters on it")},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, Remove an oil counter from Migloz: It gains vigilance and menace until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    Plus(ManaCost("{1}"), RemoveCountersFromThis(s82OneB12Oil, 1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil,
						"Migloz — vigilance and menace until end of turn",
						game.AddKeywordsMod("vigilance", "menace"))
				},
			},
			{
				Label:   "{2}, Remove two oil counters from Migloz: It gets +2/+2 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
				Cost:    Plus(ManaCost("{2}"), RemoveCountersFromThis(s82OneB12Oil, 2)),
				Effect:  thisGetsUntilEndOfTurn(2, 2, "Migloz — +2/+2 until end of turn"),
			},
			{
				Label:   "{3}, Remove three oil counters from Migloz: Destroy target artifact or enchantment.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    Plus(ManaCost("{3}"), RemoveCountersFromThis(s82OneB12Oil, 3)),
				Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return destroyTheTargetPermanent(item, NewContext(g, item))
				},
			},
		},
	})
}
