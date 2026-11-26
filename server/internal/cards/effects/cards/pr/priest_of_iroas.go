package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Priest of Iroas — Creature — Human Cleric {R}, 1/1 (Theros, S82 OLD b392, #5561):
//
//	"{3}{W}, Sacrifice this creature: Destroy target enchantment."
//
// Ronom Unicorn's sacrifice ability with a mana cost beside it: both are
// paid at announce, so the Priest is gone before the destruction resolves;
// the target is re-checked at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "706f7d56-f4da-48ba-ac94-fc09580beec7",
		Name:         "Priest of Iroas",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}, Sacrifice this creature: Destroy target enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{3}{W}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
