package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icefall Regent — Creature — Dragon {3}{U}{U}, 4/3 (Dragons of Tarkir, #5551):
//
//	"Flying
//	 When this creature enters, tap target creature an opponent controls. That creature doesn't untap during its controller's untap step for as long as you control this creature.
//	 Spells your opponents cast that target this creature cost {2} more to cast."
//
// Dungeon Geists' tap-and-hold (the hold is recorded only while the Regent is
// still on the battlefield under your control, CR 611.2b) and Boreal
// Elemental's tax on spells that target it.
//
// No simplification.

func init() {
	tax := effects.CostsMore(2,
		"Spells your opponents cast that target this creature cost {2} more to cast.",
		effects.OpponentsSpell(), effects.ThatTargetThisPermanent())
	tax.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:        "0cf18002-1bbe-40b3-ae15-fa51e61e9391",
		Name:            "Icefall Regent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers:   []game.CostModifier{tax},
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WhenThisEnters("Icefall Regent — tap target creature an opponent controls; it doesn't untap while you control Icefall Regent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.TapAndHoldWhileYouControlThis(ctx, effects.HoldTargetIDs(ctx))
				}),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		)},
	})
}
