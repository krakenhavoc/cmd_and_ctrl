package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quillspike — Creature — Beast {2}{B/G} 1/1:
//
//	"{B/G}, Remove a -1/-1 counter from a creature you control: This creature
//	 gets +3/+3 until end of turn."
//
// Ghave's cost shape: the counter comes off any creature you control, chosen
// as the cost is paid, so the ability needs one to activate (CR 118.3).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8c748ed-7ba2-4f92-a9de-5080382c2fc2",
		Name:         "Quillspike",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B/G}, Remove a -1/-1 counter from a creature you control: This creature gets +3/+3 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 3, Toughness: 3}},
			Cost:    effects.Plus(effects.ManaCost("{B/G}"), effects.RemoveCountersFrom(game.CounterMinusOne, 1, "a creature you control", effects.Creature())),
			Effect:  effects.ThisGetsUntilEndOfTurn(3, 3, "Quillspike — +3/+3 until end of turn"),
		}},
	})
}
