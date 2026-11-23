package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cankerbloom — Creature — Phyrexian Fungus {1}{G}, 3/2:
//
//	"{1}, Sacrifice this creature: Choose one —
//	 • Destroy target artifact.
//	 • Destroy target enchantment.
//	 • Proliferate."
//
// A modal activated ability (ADR 0065): the mode and its target are
// announced with the activation, and the sacrifice and the {1} are
// paid up front, so the Cankerbloom is gone whatever the answer.
// The two destroy bullets are Bant Charm's shared body; the
// proliferate bullet has no target.
//
// The proliferate bullet asks the player what to proliferate (#2525).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5b80895-621a-40df-bf48-6c7295658f21",
		Name:         "Cankerbloom",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Choose one — destroy target artifact; destroy target enchantment; or proliferate.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Modes: effects.ChooseOne(
				effects.ModeDoing("Destroy target artifact.",
					effects.TargetPermanent("target artifact", effects.Artifact()),
					effects.DestroyTheModesTarget),
				effects.ModeDoing("Destroy target enchantment.",
					effects.TargetPermanent("target enchantment", effects.Enchantment()),
					effects.DestroyTheModesTarget),
				effects.ModeDoing("Proliferate.", nil,
					func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.Proliferate{}.Apply(ctx)
					}),
			),
		}},
	})
}
