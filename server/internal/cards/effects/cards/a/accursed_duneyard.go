package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Accursed Duneyard — Land:
//
//	"{T}: Add {C}.
//	 {2}, {T}: Regenerate target Shade, Skeleton, Specter, Spirit, Vampire,
//	 Wraith, or Zombie."
//
// A colourless mana ability and a regeneration shield (CR 701.19a) for a
// creature of one of the seven types, any controller. The target's types are
// read through the layered view, so a changeling qualifies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48edc348-93f6-4dce-9cc4-7244d76b6f4a",
		Name:         "Accursed Duneyard",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Regenerate target Shade, Skeleton, Specter, Spirit, Vampire, Wraith, or Zombie.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target Shade, Skeleton, Specter, Spirit, Vampire, Wraith, or Zombie",
				effects.Or(effects.HasSubtype("Shade"), effects.HasSubtype("Skeleton"), effects.HasSubtype("Specter"), effects.HasSubtype("Spirit"),
					effects.HasSubtype("Vampire"), effects.HasSubtype("Wraith"), effects.HasSubtype("Zombie"))),
			Effect: effects.RegenerateTheTargetPermanent,
		}},
	})
}
