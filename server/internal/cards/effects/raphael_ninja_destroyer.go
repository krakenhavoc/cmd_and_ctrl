package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raphael, Ninja Destroyer — Legendary Creature — Mutant Ninja Turtle
// {2}{R}{R}, 4/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Raphael must be blocked if able.
//	 Enrage — Whenever Raphael is dealt damage, add that much {R}. Until
//	 end of turn, you don't lose this mana as steps and phases end."
//
// The first line is Gaea's Protector's CR 509.1c requirement. Enrage is
// an ability word (CR 207.2c); the trigger is the Incredible Hulk's: one
// trigger per damage event, not per source (combat damage is dealt at
// once, CR 510.2), waiting for the damage to finish (AtBatchEnd,
// ThatMuchDamage) so a Raphael blocked by three creatures triggers once
// for the total. It uses the stack (it is
// not a mana ability: it does not trigger from one, CR 605.1b), and the
// red mana carries Savage Ventmaw's keep-until-end-of-turn mark.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fdb8e862-c642-4b7b-b061-400de6061ffe",
		Name:         "Raphael, Ninja Destroyer",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{MustBeBlocked()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			},
			Key:        "Raphael, Ninja Destroyer — Enrage: add that much {R}; you don't lose it until end of turn",
			AtBatchEnd: ThatMuchDamage(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
					return nil
				}
				return AddMana{
					Amount: ManaN(item.Trigger.Event.Amount, "R"),
					Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
