package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// the dead card is read post-move, so a changeling counts and a
// Zombie that was one only through a layer effect does not; the life
// is LOST, not dealt, so no prevention shield sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f086a82d-8dbe-4d02-a076-03789704e9d4",
		Name:            "Plague Belcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets:   TargetCreature("target creature you control", YouControl()),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Plague Belcher — put two -1/-1 counters on target creature you control",
						b35PutMinusCountersOnChosen(2))
				},
			},
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return anotherZombieYouControlDied(ev, source, g)
			}, "Plague Belcher — each opponent loses 1 life", b35EachOpponentLosesOne),
		},
	})
}
