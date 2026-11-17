package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forgehammer Centurion — Creature — Phyrexian Warrior {2}{R}, 3/2
// (Phyrexia: All Will Be One):
//
//	"Whenever another creature or artifact you control is put into a
//	 graveyard from the battlefield, put an oil counter on this creature.
//	 Whenever this creature attacks, you may remove two oil counters from
//	 it. When you do, target creature can't block this turn."
//
// The counter trigger is Vat of Rebirth's reading. The attack trigger is
// a "you may" whose removal is made as it resolves; only two counters
// really removed make the reflexive trigger (CR 603.12), which chooses
// its target, any creature, as it goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa519a46-11ab-4ab8-9286-904ce2746c03",
		Name:         "Forgehammer Centurion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, s82OilOnDeparture, "Forgehammer Centurion — put an oil counter on it",
				Do(CounterOnThis{Kind: s82OneOil, N: 1})),
			Optional(WheneverThisAttacks("Forgehammer Centurion — you may remove two oil counters from it; when you do, target creature can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) || c.Counters[s82OneOil] < 2 {
						return nil
					}
					if err := g.AddCounterForEffect(item.SourceCardID, s82OneOil, -2); err != nil {
						return err
					}
					return WhenYouDo("Forgehammer Centurion — target creature can't block this turn", forgehammerCenturionBody).Apply(NewContext(g, item))
				}),
				"Forgehammer Centurion — remove two oil counters from it?"),
		},
	})
}

// forgehammerCenturionBody is the reflexive trigger. An on-disk key:
// never renamed, never reused.
var forgehammerCenturionBody = game.ReflexiveBody("forgehammer-centurion/cant-block",
	simpleBody(targetCreatureCantBlockThisTurn("Forgehammer Centurion — can't block this turn")),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))
