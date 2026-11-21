package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overclocked Electromancer — Creature — Lizard Wizard {2}{R}, 2/2 (Modern
// Horizons 3 Commander, #5948):
//
//	"At the beginning of combat on your turn, you may pay {E}{E}{E}. If you
//	 do, put a +1/+1 counter on this creature.
//	 Whenever this creature attacks, double its power until end of turn.
//	 Whenever this creature deals combat damage to a creature, if that
//	 creature was dealt excess damage this turn, you get X {E}, where X is
//	 that excess damage."
//
// The first line is the energy shape of the Thriving cycle (the payment is
// asked as the trigger resolves, CR 118.12). The second doubles the
// power the creature has as the trigger resolves (counters and pumps
// included). The third reads excess damage (CR 120.10) as the damage
// events landed: the intervening "if" holds when the damaged creature was
// dealt any excess damage this turn, from any source, and X is the total
// excess damage dealt to it this turn. Both are checked as the combat
// damage is dealt and again as the trigger resolves (CR 603.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e576de0-0846-4665-a07c-83fe483ac196",
		Name:         "Overclocked Electromancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Overclocked Electromancer — you may pay "+EnergySymbols(3),
				mayPayEnergyThen("Overclocked Electromancer", 3, "put a +1/+1 counter on it", energyPlusOneCounterOnThis)),
			WheneverThisAttacks("Overclocked Electromancer — double its power until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return DoublePowerUntilEOT(NewContext(g, item), item.SourceCardID, "Overclocked Electromancer — double its power")
				}),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				Key:     "Overclocked Electromancer — you get {E} equal to the excess damage",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Source != source.InstanceID {
						return false
					}
					victim, ok := g.LookupCardForEffect(ev.Target)
					return ok && victim.IsCreature() && gapB45ExcessDealtTo(g, ev.Target) > 0
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					x := gapB45ExcessDealtTo(g, item.Trigger.Event.Target)
					if x <= 0 { // CR 603.4: the intervening "if" is rechecked
						return nil
					}
					return GetEnergy{N: x}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
