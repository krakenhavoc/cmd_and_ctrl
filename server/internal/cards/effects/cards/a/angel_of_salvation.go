package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angel of Salvation — Creature — Angel {6}{W}{W}, 5/5:
//
//	"Flash; convoke (Your creatures can help cast this spell. Each creature
//	 you tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Flying
//	 When this creature enters, prevent the next 5 damage that would be
//	 dealt this turn to any number of targets, divided as you choose."
//
// Convoke is the shared tap cost. The enter trigger is Bogardan Hellkite's
// divided "any number of targets" with a shield for each share instead of
// damage: every target gets at least 1 (CR 601.2d), and the shields last
// the turn (CR 615.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "27b945ad-92a9-4615-87ec-790682e8c3e7",
		Name:            "Angel of Salvation",
		Completeness:    effects.CompletenessFull,
		TapCost:         effects.Convoke(),
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Targets:   effects.TargetAny().WithCount(0, 0).Dividing(effects.Divide(5)),
			Key:       "Angel of Salvation — prevent the next 5 damage that would be dealt this turn to any number of targets, divided as you choose",
			Effect:    effects.RprB04PreventDividedDamage("Angel of Salvation — prevent the next damage this turn"),
		}},
	})
}
