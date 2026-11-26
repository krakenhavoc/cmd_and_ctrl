package cardsg

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gaea's Revenge — Creature — Elemental {5}{G}{G}, 8/5 (Magic Origins, S82
// older sets):
//
//	"This spell can't be countered.
//	 Haste
//	 This creature can't be the target of nongreen spells or abilities from
//	 nongreen sources."
//
// Thrun, Breaker of Silence's shape without its opponent clause: a
// TargetingRestriction read at the engine's announce and resolution checks (CR
// 601.2c, 608.2b) that refuses a spell or an ability whose source is not green,
// whoever controls it, the controller's own included. The source's colour is
// judged as it is now or as it last existed (CR 608.2h); a colourless source is
// nongreen. Uncounterable through Spec.CantBeCountered, haste through
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "26f6cc1e-d85e-45a3-a1ed-7f09d7cfc7df",
		Name:            "Gaea's Revenge",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"haste"},
		TargetingRestrictions: []game.TargetingRestriction{{
			Label: "Gaea's Revenge can't be the target of nongreen spells or abilities from nongreen sources.",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			Forbids: func(q game.TargetingQuery) bool {
				if q.Card.InstanceID != q.Source.InstanceID || q.Controller == uuid.Nil {
					return false
				}
				ch := q.By.Characteristics()
				if ch == nil {
					return true
				}
				for _, c := range ch.Colors {
					if strings.EqualFold(c, q.Source.TextColor("G")) || strings.EqualFold(c, "green") {
						return false
					}
				}
				return true
			},
		}},
	})
}
