package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wojek Embermage — Creature — Human Wizard {3}{R}, 1/2:
//
//	"Radiance — {T}: This creature deals 1 damage to target creature and
//	 each other creature that shares a color with it."
//
// The creatures are read as the ability resolves (radianceSet, #5736),
// whoever controls them, the Embermage itself included when it shares a
// colour with the target, and each is dealt 1 damage by the Embermage as
// one instruction (radianceDamage). A {T} ability, so it waits out
// summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac8630ac-defb-4774-ac71-d35fe02a7324",
		Name:         "Wojek Embermage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target creature and each other creature that shares a color with it.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return radianceDamage(NewContext(g, item), 1, game.KeyedTail{})
			},
		}},
	})
}
