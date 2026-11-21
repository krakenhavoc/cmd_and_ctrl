package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Animal Sanctuary — Land (EDHREC rank 3770):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Put a +1/+1 counter on target Bird, Cat, Dog, Goat, Ox,
//	 or Snake."
//
// The pet-typal utility land. The colourless mana is an ordinary tap
// ability; the counter is a real activated ability with a target
// clause over permanents carrying any of the six creature types —
// read post-layer, so a changeling qualifies — and the counter goes
// through the CR 614 placement pipeline, so a Hardened Scales sees
// it. Anyone's animal is a legal target, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3c40943-1d7c-4ea2-b34f-8df8b6775701",
		Name:         "Animal Sanctuary",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Put a +1/+1 counter on target Bird, Cat, Dog, Goat, Ox, or Snake.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target Bird, Cat, Dog, Goat, Ox, or Snake", effects.AnySubtype("Bird", "Cat", "Dog", "Goat", "Ox", "Snake")),
			Effect:  effects.B36CounterOnChosenAnimal,
		}},
	})
}
