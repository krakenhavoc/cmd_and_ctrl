package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogis's Marauder — Creature — Human Berserker {2}{B}, 2/2 (Theros):
//
//	"When this creature enters, up to X target creatures each gain intimidate
//	 and haste until end of turn, where X is your devotion to black. (Each {B}
//	 in the mana costs of permanents you control counts toward your devotion to
//	 black.)"
//
// X is counted as the trigger goes on the stack (CR 700.5), the Marauder's own
// {B} included, and bounds how many targets are chosen; with no devotion
// nothing is asked. A target gone by resolution gains nothing (CR 608.2b).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Mogis's Marauder — up to X target creatures each gain intimidate and haste until end of turn",
		effects.OldB395GrantTargetsUntilEOT("Mogis's Marauder", "intimidate", "haste"))
	enters.TargetsFrom = func(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
		x := effects.DevotionTo(g, source.Controller, source.TextColor("B"))
		if x <= 0 {
			return nil
		}
		return effects.TargetCreature("up to X target creatures").WithCount(0, x)
	}
	effects.Register(effects.Spec{
		OracleID:     "7a8bfc6b-0a24-42ad-a2c5-ac8d29d4c1b9",
		Name:         "Mogis's Marauder",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
