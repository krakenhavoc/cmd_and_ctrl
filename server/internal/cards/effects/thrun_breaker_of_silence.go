package effects

import (
	"strings"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thrun, Breaker of Silence — Legendary Creature — Troll Shaman {3}{G}{G},
// 5/5 (ONE):
//
//	"This spell can't be countered.
//	 Trample
//	 Thrun can't be the target of nongreen spells your opponents control or
//	 abilities from nongreen sources your opponents control.
//	 During your turn, Thrun has indestructible."
//
// Uncounterable through Spec.CantBeCountered, trample through
// PrintedKeywords. The targeting clause is a TargetingRestriction read at
// the engine's two targeting choke points (CR 601.2c announce, CR 608.2b
// resolution): it refuses a spell or an ability whose controller is an
// opponent of Thrun's controller and whose source is not green, the
// source judged as it is now or as it last existed (CR 608.2h). "During
// your turn" is a layer 6 grant whose predicate reads the turn, so it
// switches off at the turn boundary.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "789b7af5-ac15-40b6-b5b7-f3fcdcfb52e1",
		Name:            "Thrun, Breaker of Silence",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"trample"},
		TargetingRestrictions: []game.TargetingRestriction{{
			Label: "Thrun can't be the target of nongreen spells your opponents control or abilities from nongreen sources your opponents control.",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			Forbids: func(q game.TargetingQuery) bool {
				if q.Card.InstanceID != q.Source.InstanceID || q.Controller == q.Card.Controller || q.Controller == uuid.Nil {
					return false
				}
				ch := q.By.Characteristics()
				if ch == nil {
					return false
				}
				for _, c := range ch.Colors {
					if strings.EqualFold(c, "G") || strings.EqualFold(c, "green") {
						return false
					}
				}
				return true
			},
		}},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && isActivePlayer(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = append(c.Abilities, "indestructible")
			},
		}},
	})
}
