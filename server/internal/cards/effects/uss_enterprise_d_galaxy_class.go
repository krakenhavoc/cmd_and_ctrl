package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// U.S.S. Enterprise-D, Galaxy-Class — {3} Legendary Artifact —
// Spacecraft, 4/5 (Star Trek, #4973):
//
//	"Whenever one or more charge counters are put on U.S.S. Enterprise-D
//	 for the first time each turn, exile the top card of your library.
//	 You may play that card this turn.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 7+.)
//	 7+ | Flying, vigilance"
//
// The station frame (Station, SpacecraftAt, ThresholdKeywords) with a
// trigger on EventCounterPlaced: a placement of charge counters on this
// object, from station or from anything else (Warp Core), and the
// first such placement on this object this turn
// (trekB28FirstChargeCountersOnThisThisTurn). It triggers below the
// threshold too, since the ability is printed above the bar. The body
// is Outpost Siege's impulse exile, the PLAY grant, so a land can be
// played from it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d95af032-3efd-40c7-8229-ade9d974934f",
		Name:         "U.S.S. Enterprise-D, Galaxy-Class",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return trekB28FirstChargeCountersOnThisThisTurn(ev, source, g)
			}, "U.S.S. Enterprise-D — exile the top card of your library; you may play it this turn",
				outpostSiegeImpulse),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(7, 4, 5),
			ThresholdKeywords(7, "flying", "vigilance"),
		),
	})
}
