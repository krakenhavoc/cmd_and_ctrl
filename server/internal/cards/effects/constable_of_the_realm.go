package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Constable of the Realm — Creature — Giant Soldier {4}{W}, 3/3:
//
//	"Renown 2 (When this creature deals combat damage to a player, if it
//	 isn't renowned, put two +1/+1 counters on it and it becomes renowned.)
//	 Whenever one or more +1/+1 counters are put on this creature, exile up
//	 to one other target nonland permanent until this creature leaves the
//	 battlefield."
//
// Renown is the engine's keyword trigger (game/renown.go), whose counters
// are an ordinary EventCounterPlaced, so they fire the second ability like
// any other counters (Hardened Scales and the rest of CR 614 apply first).
// The exile is Aang's Iceberg's until-this-leaves one-shot (CR 610.3): the
// permanent returns as a new object under its owner's control the moment
// the Constable leaves, and nothing is exiled if it left before the trigger
// resolved (CR 610.3b). The target is "up to one other", chosen as the
// trigger goes on the stack.
//
// No simplification.
func init() {
	const label = "Constable of the Realm — exile up to one other target nonland permanent until this creature leaves the battlefield"
	trig := On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return plusCountersOnThis(ev, source, g)
	}, label, exileChosenTargetUntilThisLeaves("Constable of the Realm — the exiled card returns when Constable of the Realm leaves the battlefield"))
	Register(Spec{
		OracleID:        "2a263e61-303f-449d-a431-c26cabdc68e0",
		Name:            "Constable of the Realm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"renown 2"},
		Triggered: []game.TriggeredAbility{
			Targeting(trig, Another(TargetPermanent("up to one other target nonland permanent", Nonland()).WithCount(0, 1))),
		},
	})
}
