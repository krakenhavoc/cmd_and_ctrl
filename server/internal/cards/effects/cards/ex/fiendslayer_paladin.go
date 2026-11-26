package cardsex

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fiendslayer Paladin — Creature — Human Knight {1}{W}{W}, 2/2 (Magic 2014,
// S82 old sets):
//
//	"First strike
//	 Lifelink
//	 This creature can't be the target of black or red spells your opponents
//	 control."
//
// Thrun's targeting restriction, narrowed to spells: a spell whose controller
// is an opponent of the Paladin's controller and which is black or red (or
// both), judged as it is now or as it last existed (CR 608.2h). An ability from
// a black or red source is not a spell and can still target it, and a spell of
// either colour you control can.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3eb02993-b2fc-4185-aa67-08c8f133ce42",
		Name:            "Fiendslayer Paladin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "lifelink"},
		TargetingRestrictions: []game.TargetingRestriction{{
			Label: "Fiendslayer Paladin can't be the target of black or red spells your opponents control.",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			Forbids: func(q game.TargetingQuery) bool {
				if q.Card.InstanceID != q.Source.InstanceID || q.Controller == q.Card.Controller || q.Controller == uuid.Nil || !q.By.IsSpell() {
					return false
				}
				ch := q.By.Characteristics()
				if ch == nil {
					return false
				}
				for _, c := range ch.Colors {
					if strings.EqualFold(c, "B") || strings.EqualFold(c, "R") || strings.EqualFold(c, "black") || strings.EqualFold(c, "red") {
						return true
					}
				}
				return false
			},
		}},
	})
}
