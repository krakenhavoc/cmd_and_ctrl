package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frostweb Spider — Snow Creature — Spider {2}{G}, 1/3 (Coldsnap, #5597):
//
//	"Reach
//	 Whenever this creature blocks a creature with flying, put a +1/+1 counter
//	 on this creature at end of combat."
//
// One trigger per flying attacker blocked (CR 509.3b); its resolution sets up
// a delayed trigger at the beginning of this combat's end of combat step
// (Clockwork Beetle's shape), which does nothing if the Spider has left.
//
// No simplification.
// frostwebSpiderCounterBody is "put a +1/+1 counter on this creature at end of
// combat", the delayed trigger the Spider schedules. A registered key (ADR 0041
// phase 3), so a table with it waiting is a restore point; assigned in init, as
// Clockwork Beetle's body is.
var frostwebSpiderCounterBody game.BodyRef

func init() {
	frostwebSpiderCounterBody = game.SimpleDelayedBody("old/s82-b565-frostweb-spider-counter-at-end-of-combat",
		func(g *game.Game, item *game.StackItem) error {
			if g.AbilitySourceGoneForEffect(item) {
				return nil
			}
			return effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(effects.NewContext(g, item))
		})
	effects.Register(effects.Spec{
		OracleID:        "42176320-ee7f-44c9-916e-5ee4a6108be3",
		Name:            "Frostweb Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.S82OldB565BlocksAFlyer, "Frostweb Spider — put a +1/+1 counter on it at end of combat",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ScheduleDelayedTrigger{
						At:    game.StepEndCombat,
						Label: "Frostweb Spider — put a +1/+1 counter on it",
						Body:  frostwebSpiderCounterBody,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
