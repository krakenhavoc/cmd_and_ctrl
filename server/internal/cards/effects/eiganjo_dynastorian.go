package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eiganjo Dynastorian // Replenish — a preparation card (CR 722,
// ADR 0090), Secrets of Strixhaven Commander (#4834):
//
//	Eiganjo Dynastorian {2}{W}, Creature — Fox Advisor 2/3:
//	"Vigilance
//	 Whenever you attack with two or more creatures, this creature
//	 becomes prepared."
//	Replenish {3}{W}, Sorcery:
//	"Return all enchantment cards from your graveyard to the
//	 battlefield."
//
// "Attack with two or more creatures" is Spinneret and Spiderling's
// count: the EventAttack events of one declaration by you, settled as a
// batch, one trigger when there are at least two
// (ThatManyEventsAtLeast). The Dynastorian need not be one of the
// attackers, and a second combat's declaration is its own batch. The
// prepare spell is Replenish's body (replenish.go).
//
// No simplification.
const eiganjoDynastorianOracleID = "09b7ea91-2620-4c45-af3c-60bbeab89966"

func init() {
	Register(Spec{
		OracleID:        eiganjoDynastorianOracleID,
		Name:            "Eiganjo Dynastorian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			withAtBatchEnd(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && attackDeclaredByYou(ev, source.Controller)
			}, "Eiganjo Dynastorian — becomes prepared (you attacked with two or more creatures)", socB4ThisBecomesPrepared),
				ThatManyEventsAtLeast(2)),
		},
	})
	Register(Spec{
		OracleID:     eiganjoDynastorianOracleID + "#1",
		Name:         "Replenish",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b36ReturnAllEnchantmentCards(ctx)
		},
	})
}
