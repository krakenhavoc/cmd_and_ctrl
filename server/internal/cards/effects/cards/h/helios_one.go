package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// HELIOS One — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}: You get {E} (an energy counter).
//	 {3}, {T}, Pay X {E}, Sacrifice this land: Destroy target nonland
//	 permanent with mana value X. Activate only as a sorcery."
//
// ADR 0129 §2: "Pay X {E}" is PayXEnergy — X is announced with the
// activation (CR 107.3a), may not exceed the activator's energy
// (CR 118.3), and binds the target clause's "with mana value X", which
// the engine checks at announce (CR 601.2c via CR 602.2b) and again at
// resolution (CR 608.2b). X = 0 is a real announcement: a token or a
// land-less zero-cost permanent has mana value 0, so XMatters is not
// declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfb1a656-0bf1-484d-b099-33087914250b",
		Name:         "HELIOS One",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: You get {E} (an energy counter).",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerValue, Energy: 1},
				Effect:  effects.Do(effects.GetEnergy{N: 1}),
			},
			{
				Label: "{3}, {T}, Pay X {E}, Sacrifice this land: Destroy target nonland permanent with mana value X. Activate only as a sorcery.",
				Cost:  effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.PayXEnergy(), effects.SacrificeThis()),
				Targets: effects.TargetPermanent("target nonland permanent with mana value X",
					effects.Nonland()).WithManaValueEqualsX(),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyEachLegalTarget(item, effects.NewContext(g, item))
				},
			},
		},
	})
}
