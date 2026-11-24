package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moderation — Enchantment {1}{W}{U} (Modern Horizons 2, #5501):
//
//	"You can't cast more than one spell each turn.
//	 Whenever you cast a spell, draw a card."
//
// The restriction binds the controller only (Rakdos's "you" reading, not
// Rule of Law's every seat), counted off the per-turn cast tally. The draw
// is a cast trigger over any spell, so it goes on the stack above the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57f94fcb-a224-410b-a38e-3ed8cca7d0c7",
		Name:         "Moderation",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{{
			Label: "Moderation — you can't cast more than one spell each turn.",
			Forbids: func(q game.CastQuery) bool {
				return q.Source.Controller == q.Controller && q.Game.CastTallyFor(q.Controller).Total >= 1
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(nil, "Moderation — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
