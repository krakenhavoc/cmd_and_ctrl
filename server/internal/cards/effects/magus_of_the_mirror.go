package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magus of the Mirror — Creature — Human Wizard {4}{B}{B}, 4/2 :
//
//	"{T}, Sacrifice this creature: Exchange life totals with target
//	 opponent. Activate only during your upkeep."
//
// Mirror Universe on a body, and the restrictions are the card. The
// tap means it has to survive a turn cycle; the sacrifice means the
// swap happens once; and "only during your upkeep" is the one that
// makes it fair — you cannot wait until an opponent has burned
// themselves low in response to something, you have to commit in the
// first step of your own turn, before you have drawn.
//
// That last clause is an activation instruction (CR 602.1b), not a
// timing speed: DuringYourUpkeep, which checks BOTH the step and that
// you are the active player. SorcerySpeed would have been wrong twice
// over — it means "your main phase with an empty stack", which
// forbids the upkeep entirely.
//
// The costs are paid at announce, so the Magus is already in the
// graveyard when the ability resolves and a dies-trigger it sets off
// sits ABOVE the exchange and resolves first. That is printed
// behaviour, and it is also why the exchange must not read anything
// off the source: it reads the item's controller.
//
// An opponent removed in response leaves the ability with no legal
// target and it is countered on resolution — the creature is still
// sacrificed, because a cost paid is paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f587676e-d45f-4b1d-be2c-b404d8f55fb3",
		Name:         "Magus of the Mirror",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}, Sacrifice this creature: Exchange life totals with target opponent. Activate only during your upkeep.",
			Cost:      Plus(TapCost(), SacrificeThis()),
			Condition: DuringYourUpkeep(),
			Targets:   TargetPlayer("target opponent", Opponent()),
			Effect:    magusOfTheMirrorExchange,
		}},
	})
}

// magusOfTheMirrorExchange swaps the activator's life total with the
// chosen opponent's. "With" makes the controller one of the two
// halves, read off the item rather than off the source — the source
// is in the graveyard by now, sacrificed as part of the cost.
func magusOfTheMirrorExchange(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return exchangeLifeTotals(ctx, item.Controller, victim)
}
