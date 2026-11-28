package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hand to Hand — Enchantment {2}{R} (Tempest, S82 OLD batch 765, #5628):
//
//	"During combat, players can't cast instant spells or activate abilities
//	 that aren't mana abilities."
//
// Two statics that bind every player, the controller included (CR 101.2): a
// cast restriction on instant spells and an activation restriction on every
// activated ability that is not a mana ability, from any zone (a cycling
// ability in a hand is an activated ability too). "During combat" is the
// combat phase's steps, beginning of combat through end of combat
// (InCombatStep). A flash creature is not an instant spell and a triggered
// ability is not activated, so both are untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf45f579-539c-4ffe-89df-4896a394158f",
		Name:         "Hand to Hand",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{{
			Label: "During combat, players can't cast instant spells.",
			Forbids: func(q game.CastQuery) bool {
				return effects.InCombatStep(q.Game.Turn.Step) && q.Card.IsInstant()
			},
		}},
		ActivationRestrictions: []game.ActivationRestriction{{
			Label: "During combat, players can't activate abilities that aren't mana abilities.",
			Forbids: func(q game.ActivationQuery) bool {
				return effects.InCombatStep(q.Game.Turn.Step) && !q.Ability.Mana
			},
		}},
	})
}
