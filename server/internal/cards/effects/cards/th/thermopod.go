package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thermopod — Snow Creature — Slug {4}{R}, 4/3 (Coldsnap, #5597):
//
//	"{S}: This creature gains haste until end of turn.
//	 Sacrifice a creature: Add {R}."
//
// {S} is snow mana (Icebind Pillar). The second is a mana ability whose cost
// sacrifices a creature, this one included (Phyrexian Altar's cost), so it
// resolves at once and can't be responded to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c886fa01-bcee-4c73-9c46-2328711fc4f0",
		Name:         "Thermopod",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{S}: This creature gains haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{S}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("haste", "Thermopod — haste until end of turn"),
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{SacrificeOther: effects.SacrificeACreature().SacrificeOther},
			Produced: "{R}",
			Label:    "Sacrifice a creature: Add {R}",
			Answers:  game.AnswerSacOutlet,
		}},
	})
}
