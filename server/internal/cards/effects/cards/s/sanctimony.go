package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctimony — Enchantment {1}{W} (Eighth Edition, #5607):
//
//	"Whenever an opponent taps a Mountain for mana, you may gain 1 life."
//
// Vorinclex's watch on a land tapped for mana by an opponent. It gains life, not
// mana, so it is no triggered mana ability and uses the stack (CR 605.5a). The
// land type is read through the Enchantment's text (CR 612.2); the gain is
// optional, asked as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fc7740a-dfa4-4279-8822-3f7c93e6b685",
		Name:         "Sanctimony",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventManaAbilityActivated, effects.OldB608AnOpponentTappedAMountain,
				"Sanctimony — you may gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
				"Sanctimony — gain 1 life?"),
		},
	})
}
