package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Transmogrifying Wand — Artifact {3} (Commander Masters):
//
//	"This artifact enters with three charge counters on it.
//	 {1}, {T}, Remove a charge counter from this artifact: Destroy target
//	 creature. Its controller creates a 2/4 white Ox creature token.
//	 Activate only as a sorcery."
//
// Tumble Magnet's three charge counters and counter-removal cost, with
// Harsh Annotation's destroy-then-its-controller-creates body: the
// controller is read before the destroy, so the Ox goes to whoever held the
// creature as the ability resolved, and is made whether or not the destroy
// did anything (an indestructible creature). The ability is sorcery-speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "921345ce-2eaf-4b95-b65d-1c27204665ed",
		Name:         "Transmogrifying Wand",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterCharge, 3, "Transmogrifying Wand: enters with three charge counters"),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}, {T}, Remove a charge counter from this artifact: Destroy target creature. Its controller creates a 2/4 white Ox creature token. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}"), TapCost(), RemoveCountersFromThis(game.CounterCharge, 1)),
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Purpose:      ForTargets(RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 2, TokenToughness: 4})),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return socB26DestroyThenItsControllerCreates("2/4 white Ox")(item, NewContext(g, item))
			},
		}},
	})
}
