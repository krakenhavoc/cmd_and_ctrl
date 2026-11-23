package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dungeon Geists — Creature — Spirit {2}{U}{U}, 3/3:
//
//	"Flying
//	 When this creature enters, tap target creature an opponent
//	 controls. That creature doesn't untap during its controller's
//	 untap step for as long as you control this creature."
//
// #1313: an untap hold with a "for as long as you control ~" duration
// (ADR 0058's 2026-09-23 amendment). The card whose rulings pin the
// family (2019-07-12): leaving before the trigger resolves taps the
// creature with no hold, and losing control ends the hold even if
// control comes back.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ab5ebae2-cd77-4a7d-a93b-8042cd486429",
		Name:            "Dungeon Geists",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WhenThisEnters("Dungeon Geists — tap target creature; it doesn't untap while you control Dungeon Geists",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.TapAndHoldWhileYouControlThis(ctx, effects.HoldTargetIDs(ctx))
				}),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		)},
	})
}
