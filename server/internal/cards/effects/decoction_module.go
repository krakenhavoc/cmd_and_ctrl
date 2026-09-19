package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Decoction Module — Artifact {2} (EDHREC rank 4202):
//
//	"Whenever a creature you control enters, you get {E} (an energy
//	 counter).
//	 {4}, {T}: Return target creature you control to its owner's hand."
//
// One of the three Kaladesh Modules. On its own it is a two-mana
// blink outlet that stockpiles energy; the reason it is played is the
// other two Modules, which SPEND that energy — and that half is not
// in this batch.
//
// Energy is a real player counter in this engine (game.CounterEnergy,
// CR 107.14 / 122.1), so "you get {E}" is an ordinary player-counter
// placement and the total is visible on the player. There is no rules
// content to get wrong: energy has no state-based action, no maximum,
// and no loss condition.
//
// What DOES need saying is the seam on the other side. Nothing in the
// catalog PAYS energy yet — game.AbilityCost has no energy component
// — so a Decoction Module in play today accumulates a counter that no
// card in the catalog can spend. That is not a simplification of THIS
// card: every word of Decoction Module is implemented, and the gap is
// in the cards that would use it. Registering it now means the
// tracking is already right when a spender lands.
//
// The trigger reads "a creature YOU CONTROL enters", not "another" —
// the Module is not a creature, so the distinction never comes up,
// but a creature that enters and immediately dies has still entered
// and still pays out. One event per creature, so a mass token
// creation pays once per token.
//
// The activated ability is a real CR 602 ability on the stack, so a
// creature bounced with it can be answered in response, and "your
// own" is the target restriction — it is a blink/save outlet, not
// removal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1daba4f6-2a7d-426a-97d1-0298a7100c45",
		Name:         "Decoction Module",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			}, "Decoction Module — you get {E}", func(g *game.Game, item *game.StackItem) error {
				return g.AddPlayerCounterForEffect(item.Controller, game.CounterEnergy, 1)
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Return target creature you control to its owner's hand.",
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return BounceToHand{Target: id}.Apply(ctx)
			},
		}},
	})
}
