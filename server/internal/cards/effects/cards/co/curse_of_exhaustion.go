package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Exhaustion — Enchantment — Aura Curse {2}{W}{W} (Dark Ascension,
// S82 OLD, #5571):
//
//	"Enchant player
//	 Enchanted player can't cast more than one spell each turn."
//
// Rule of Law's cast limit (the per-turn tally, bumped as each cast succeeds)
// narrowed to the player this Curse is attached to: the restriction reads its
// source's attachment as each cast is announced, so it follows the Curse and
// binds nobody else. Spells cast by someone else's effect still count against
// the caster, as ever.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6659fae5-09d9-4095-8de3-643afd6ad3be",
		Name:         "Curse of Exhaustion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		CastRestrictions: []game.CastRestriction{{
			Label: "Curse of Exhaustion — enchanted player can't cast more than one spell each turn.",
			Forbids: func(q game.CastQuery) bool {
				return q.Source.AttachedTo.Kind == game.TargetPlayer && q.Source.AttachedTo.ID == q.Controller &&
					q.Game.CastTallyFor(q.Controller).Total >= 1
			},
		}},
	})
}
