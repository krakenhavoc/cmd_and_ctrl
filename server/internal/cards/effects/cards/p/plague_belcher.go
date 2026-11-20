package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plague Belcher — Creature — Zombie Beast {2}{B}, 5/4 (EDHREC rank
// 3680):
//
//	"Menace
//	 When this creature enters, put two -1/-1 counters on target
//	 creature you control.
//	 Whenever another Zombie you control dies, each opponent loses 1
//	 life."
//
// A three-mana 5/4 with a drawback you aim: the entry trigger
// targets a creature the controller controls, the Belcher itself
// included — it is on the battlefield when the trigger goes on the
// stack, so it is a legal answer and the usual one with an empty
// board, making it a 3/2 menace. Menace rides PrintedKeywords. The
// drain is Undead Augur's Zombie-dies read narrowed to "another":
// the Zombie is read as it last existed on the battlefield (CR
// 603.10a, #1679), so a changeling counts and so does a creature
// that was a Zombie only through a grant; the life is LOST, not
// dealt, so no prevention shield sees it.
//
// The clause is Harmful (#4249, #4314): the counters hurt the creature
// picked, so a bot puts them on its least valuable one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f086a82d-8dbe-4d02-a076-03789704e9d4",
		Name:            "Plague Belcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.B06SelfETB,
				Targets:   effects.Harmful(effects.TargetCreature("target creature you control", effects.YouControl())),
				Key:       "Plague Belcher — put two -1/-1 counters on target creature you control",
				Effect:    effects.B35PutMinusCountersOnChosen(2),
			},
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AnotherZombieYouControlDied(ev, source, g)
			}, "Plague Belcher — each opponent loses 1 life", effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
	})
}
