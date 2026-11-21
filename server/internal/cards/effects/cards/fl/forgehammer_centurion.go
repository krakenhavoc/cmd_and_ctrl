package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "fa519a46-11ab-4ab8-9286-904ce2746c03",
		Name:         "Forgehammer Centurion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.S82OilOnDeparture, "Forgehammer Centurion — put an oil counter on it",
				effects.Do(effects.CounterOnThis{Kind: effects.S82OneOil, N: 1})),
			effects.Optional(effects.WheneverThisAttacks("Forgehammer Centurion — you may remove two oil counters from it; when you do, target creature can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) || c.Counters[effects.S82OneOil] < 2 {
						return nil
					}
					if err := g.AddCounterForEffect(item.SourceCardID, effects.S82OneOil, -2); err != nil {
						return err
					}
					return effects.WhenYouDo("Forgehammer Centurion — target creature can't block this turn", forgehammerCenturionBody).Apply(effects.NewContext(g, item))
				}),
				"Forgehammer Centurion — remove two oil counters from it?"),
		},
	})
}

// forgehammerCenturionBody is the reflexive trigger. An on-disk key:
// never renamed, never reused.
var forgehammerCenturionBody = game.ReflexiveBody("forgehammer-centurion/cant-block",
	effects.SimpleBody(effects.TargetCreatureCantBlockThisTurn("Forgehammer Centurion — can't block this turn")),
	effects.ConstTargets(func() *game.TargetSpec { return effects.TargetCreature("target creature") }))
