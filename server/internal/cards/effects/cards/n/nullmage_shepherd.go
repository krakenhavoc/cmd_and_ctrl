package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nullmage Shepherd — Creature — Elf Shaman {3}{G}, 2/4 (Commander Starter
// Decks):
//
//	"Tap four untapped creatures you control: Destroy target artifact or
//	 enchantment."
//
// Baylen's tap-others cost with a count of four over any untapped creatures
// you control, the Shepherd itself among them when it is untapped (there is
// no "another", and the cost is not the {T} symbol, so a Shepherd that came
// under your control this turn can pay it, CR 302.6). The destroy is the
// shared one on the ability's target; nothing happens if it has gone.
//
// No simplification.
func init() {
	const label = "four untapped creatures you control"
	effects.Register(effects.Spec{
		OracleID:     "4dcfd397-9d29-420c-8fd8-c7c2a99de122",
		Name:         "Nullmage Shepherd",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap four untapped creatures you control: Destroy target artifact or enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  4,
				Filter: effects.TargetPermanent(label, effects.Creature()),
				Label:  label,
			}},
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
