package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Devout Chaplain — Creature — Human Cleric {2}{W}, 2/2 (Commander 2020, S82
// OLD, #5513):
//
//	"{T}, Tap two untapped Humans you control: Exile target artifact or
//	 enchantment."
//
// Skirsdag High Priest's cost shape: the {T} is the symbol (summoning
// sickness applies, CR 302.6) and the two Humans are a separate tap-others
// cost that is not {T}, so a Human that just arrived may pay it; the Chaplain
// cannot pay for itself (CR 118.3). The exile is Strip Mine's shape, judged
// again on resolution (CR 608.2b).
//
// No simplification.
func init() {
	const tapTwo = "two untapped Humans you control"
	effects.Register(effects.Spec{
		OracleID:     "1e575a82-e128-4637-87cd-f075654bc41b",
		Name:         "Devout Chaplain",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Tap two untapped Humans you control: Exile target artifact or enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost: game.AbilityCost{
				Tap: true,
				TapOthers: &game.TapOthersCost{
					Count:         2,
					Filter:        effects.TargetPermanent(tapTwo, effects.Subtype("Human")),
					ExcludeSource: true,
					Label:         tapTwo,
				},
			},
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.B27ExileChosenTarget,
		}},
	})
}
