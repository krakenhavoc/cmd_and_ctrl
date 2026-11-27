package cardsj

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karplusan Strider — Creature — Yeti {3}{G}, 3/4 (Coldsnap, #5597):
//
//	"This creature can't be the target of blue or black spells."
//
// A targeting restriction (CR 115.2) on this permanent alone, read off the
// spell that would choose it: only a SPELL that is blue or black is refused, so
// an ability from a blue or black source, or a colorless or multicolor spell
// with neither color, may still target it. Whoever controls the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2074c37-6e76-4460-9666-010c7f38ae66",
		Name:         "Karplusan Strider",
		Completeness: effects.CompletenessFull,
		TargetingRestrictions: []game.TargetingRestriction{{
			Label: "This creature can't be the target of blue or black spells.",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			Forbids: func(q game.TargetingQuery) bool {
				if q.Card.InstanceID != q.Source.InstanceID || !q.By.IsSpell() {
					return false
				}
				ch := q.By.Characteristics()
				if ch == nil {
					return false
				}
				for _, c := range ch.Colors {
					if strings.EqualFold(c, "U") || strings.EqualFold(c, "blue") ||
						strings.EqualFold(c, "B") || strings.EqualFold(c, "black") {
						return true
					}
				}
				return false
			},
		}},
	})
}
