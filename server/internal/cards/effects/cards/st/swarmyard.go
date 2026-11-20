package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swarmyard — Land:
//
//	"{T}: Add {C}.
//	 {T}: Regenerate target Insect, Rat, Spider, or Squirrel."
//
// Asceticism's regeneration shield with a creature-type clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "4b508087-99da-4eb1-8b12-29162f2ec85d",
		Name:          "Swarmyard",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Regenerate target Insect, Rat, Spider, or Squirrel.",
			Cost:  effects.TapCost(),
			Targets: effects.TargetCreature("target Insect, Rat, Spider, or Squirrel",
				effects.Or(effects.HasSubtype("Insect"), effects.HasSubtype("Rat"), effects.HasSubtype("Spider"), effects.HasSubtype("Squirrel"))),
			Effect: effects.RegenerateTheTargetPermanent,
		}},
	})
}
