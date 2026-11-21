package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Singer of Swift Rivers — Creature — Merfolk Shaman {1}{G}{U}, 3/2:
//
//	"Flash
//	 When this creature enters, put a shield counter on another target
//	 creature you control.
//	 You may cast Merfolk spells as though they had flash."
//
// The shield counter is CR 122.1c's engine-owned replacement and
// prevention (game/shield_counters.go). "Another" keeps the Singer out of
// its own target list. The flash grant is a cast-timing rule narrowed to
// Merfolk spells (Yeva's shape, with a creature-type filter).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7df43670-c948-464d-9bb9-db672e755f0c",
		Name:            "Singer of Swift Rivers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{CreatureType: "Merfolk"},
				"You may cast Merfolk spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WithTargetsFrom(
				effects.WhenThisEnters("Singer of Swift Rivers — put a shield counter on another target creature you control",
					effects.ShieldCounterOnTarget),
				effects.AnotherTarget(func(other effects.CardPredicate) *game.TargetSpec {
					return effects.TargetCreature("another target creature you control", effects.YouControl(), other)
				})),
		},
	})
}
