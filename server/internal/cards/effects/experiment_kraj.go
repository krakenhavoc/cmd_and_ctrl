package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Experiment Kraj — Legendary Creature — Ooze Mutant {2}{G}{G}{U}{U},
// 4/6:
//
//	"Experiment Kraj has all activated abilities of each other creature
//	 with a +1/+1 counter on it.
//	 {T}: Put a +1/+1 counter on target creature."
//
// A layer-6 borrow over the battlefield (ADR 0155). The counter is read
// after layers 1-5 settle, so a creature gains the lent rows the moment
// the counter lands and loses them when it goes. Kraj's own ability is
// what puts the counter on, and a second Kraj with a counter lends its
// rows back, which the rules count as two instances of the same ability
// (CR 113.2c, ADR 0155 owner decision 4); the client shows them once.
//
// No simplification.
func init() {
	const clause = "Experiment Kraj has all activated abilities of each other creature with a +1/+1 counter on it."
	Register(Spec{
		OracleID:     "93cfa771-067f-44ed-9e29-6caf698ee9aa",
		Name:         "Experiment Kraj",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on target creature.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  b36CounterOnChosenAnimal,
		}},
		Static: []game.StaticAbility{HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label: clause,
			Lenders: BattlefieldLenders(func(_ *game.Game, _ *game.Card, c game.Card) bool {
				return c.IsCreature() && c.Counters[game.CounterPlusOne] > 0
			}),
		})},
	})
}
