package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Miren, the Moaning Well — Legendary Land:
//
//	"{T}: Add {C}.
//	 {3}, {T}, Sacrifice a creature: You gain life equal to the
//	 sacrificed creature's toughness."
//
// Greater Good's sacrifice cost and last-known-information read, with
// toughness (CR 608.2h) and a life gain. Counters and an anthem's bonus
// count, because the toughness is the one the creature had on the
// battlefield. A creature whose toughness was zero or less gains
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03fe19bb-8e22-4030-8299-2ddd2d5a7eb2",
		Name:         "Miren, the Moaning Well",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice a creature: You gain life equal to the sacrificed creature's toughness.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{3}"), TapCost(), SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				fed, ok := b17PermanentSacrificedToPay(g, item)
				if !ok {
					return nil
				}
				toughness := departedCreatureToughness(g, fed)
				if toughness <= 0 {
					return nil
				}
				return GainLife{Player: item.Controller, Amount: toughness}.Apply(NewContext(g, item))
			},
		}},
	})
}
