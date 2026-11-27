package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wicker Warcrawler — Artifact Creature — Scarecrow {5}, 6/6 (Shadowmoor, S82
// OLD b532, #5591):
//
//	"Whenever this creature attacks or blocks, put a -1/-1 counter on it at end
//	 of combat."
//
// The Clockwork creatures' trigger with a -1/-1 counter: one trigger for the
// attack or the block (a creature blocking several attackers triggers once, CR
// 509.3a), and a delayed trigger at the beginning of the end of combat step (CR
// 511.2) whose source is this object. It does nothing if the Warcrawler has left
// (CR 603.7c, 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0acadb47-17b2-47d8-9aad-5436931bc38c",
		Name:         "Wicker Warcrawler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, effects.ThisAttacksOrBlocks,
				"Wicker Warcrawler — put a -1/-1 counter on it at end of combat",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ScheduleDelayedTrigger{
						At:    game.StepEndCombat,
						Label: "Wicker Warcrawler — put a -1/-1 counter on it",
						Body:  warcrawlerBody,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// warcrawlerBody is the delayed trigger's body: a -1/-1 counter on the
// Warcrawler if it is still that object. A registered key (ADR 0041 phase 3), so
// a table with it waiting is a restore point. Assigned in init, as clockwork's is:
// a var initialiser would be an initialisation cycle through the counter
// primitives.
var warcrawlerBody game.BodyRef

func init() {
	warcrawlerBody = game.SimpleDelayedBody("old-b532/wicker-warcrawler-minus-one-counter-at-end-of-combat",
		func(g *game.Game, item *game.StackItem) error {
			if g.AbilitySourceGoneForEffect(item) {
				return nil
			}
			return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterMinusOne, N: 1}.Apply(effects.NewContext(g, item))
		})
}
