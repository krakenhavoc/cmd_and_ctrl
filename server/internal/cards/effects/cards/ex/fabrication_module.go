package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b057c8ff-f169-41cd-a594-026f1a6cb0f9",
		Name:         "Fabrication Module",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: You get {E}.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Energy: 1},
			Effect:  effects.Do(effects.GetEnergy{N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouGetEnergy("Fabrication Module — a +1/+1 counter on target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					for _, t := range effects.NewContext(g, item).LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := g.AddCounterByForEffect(item.Controller, t.ID, game.CounterPlusOne, 1); err != nil {
							return err
						}
					}
					return nil
				}), effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
