package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fabrication Module — Artifact {3}:
//
//	"Whenever you get one or more {E} (energy counters), put a +1/+1
//	 counter on target creature you control.
//	 {4}, {T}: You get {E}."
//
// ADR 0129 §6 (#1995): one trigger per placement of energy (CR 603.2c),
// its own tap ability's included. The target is chosen as the trigger
// goes on the stack and re-checked as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b057c8ff-f169-41cd-a594-026f1a6cb0f9",
		Name:         "Fabrication Module",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: You get {E}.",
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Energy: 1},
			Effect:  Do(GetEnergy{N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouGetEnergy("Fabrication Module — a +1/+1 counter on target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					for _, t := range NewContext(g, item).LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := g.AddCounterByForEffect(item.Controller, t.ID, game.CounterPlusOne, 1); err != nil {
							return err
						}
					}
					return nil
				}), TargetCreature("target creature you control", YouControl())),
		},
	})
}
