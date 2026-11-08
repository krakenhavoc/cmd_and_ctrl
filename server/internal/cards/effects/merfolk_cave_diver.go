package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merfolk Cave-Diver — Creature — Merfolk Scout {2}{U}, 2/4 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Whenever a creature you control explores, this creature gets +1/+0
//	 until end of turn and can't be blocked this turn."
//
// The trigger watches game.EventExplored, which fires once an explore is
// complete (CR 701.44b): CardID is the explorer and Actor its controller.
// It is read as connive's payoff is (ACreatureYouControlConnived): the
// controller as the explore happened, and a creature still on the
// battlefield must still be a creature; one that left mid-explore still
// explored (CR 701.44c). The Cave-Diver itself counts when it explores.
// The pump and the evasion are Nimrodel Watcher's
// (thisGetsPlusOneZeroAndCantBeBlockedThisTurn), applied only while the
// Cave-Diver is still on the battlefield.
//
// No simplification.
func init() {
	const label = "Merfolk Cave-Diver — +1/+0 and can't be blocked this turn"
	Register(Spec{
		OracleID:     "2f1019df-3e76-446c-ba33-f9dfe56b10bf",
		Name:         "Merfolk Cave-Diver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventExplored, aCreatureYouControlExplored, label,
				thisGetsPlusOneZeroAndCantBeBlockedThisTurn(label)),
		},
	})
}

// aCreatureYouControlExplored is "whenever a creature you control
// explores" (CR 701.44b): the explorer's controller is the source's, and
// an explorer still around is a creature.
func aCreatureYouControlExplored(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventExplored || ev.Actor != source.Controller {
		return false
	}
	if c, ok := g.LookupCardForEffect(ev.CardID); ok {
		return c.IsCreature()
	}
	return true
}
