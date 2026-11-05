package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battle-Scarred Survivalist — Creature — Klingon Scout {1}{B}, 1/1
// (Star Trek):
//
//	Menace
//	Whenever another creature dies, put a +1/+1 counter on this creature.
//	This ability triggers only once each turn.
//
// "Only once each turn" is the per-turn trigger tally of this object
// (b11TriggeredThisTurn), and OncePerBatch keeps two creatures dying at
// once from both getting past it before the first is put on the stack.
//
// No simplification.
func init() {
	const label = "Battle-Scarred Survivalist — put a +1/+1 counter on this creature"
	Register(Spec{
		OracleID:        "0e0e15d9-bd29-495a-9975-d1bd2ebb4bff",
		Name:            "Battle-Scarred Survivalist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return AnotherCreatureDied(ev, source, lki, g) && !b11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return plusOneCountersOnThis(1)(g, item)
			})),
		},
	})
}
