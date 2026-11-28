package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keeper of the Light — Creature — Human Wizard {W}{W}, 1/2 (Exodus, S82 OLD batch 743, #5626):
//
//	"{W}, {T}: Choose target opponent who has more life than you do as you
//	 activate this ability. You gain 3 life."
//
// The opponent is only a condition: the life totals are read as the ability is
// activated (CR 602.2b), and the gain is yours whatever happens to them after.
// So the clause carries StillResolves: an opponent whose life drops in
// response does not make the ability fizzle (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62d1300f-2b76-42b9-8e6c-f920712cd547",
		Name:         "Keeper of the Light",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Choose target opponent who has more life than you do as you activate this ability. You gain 3 life.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.S82OldB743TargetOpponentWithMoreLife(),
			Effect:  effects.Do(effects.GainLife{Amount: 3}),
		}},
	})
}
