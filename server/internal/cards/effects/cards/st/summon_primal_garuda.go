package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summon: Primal Garuda — Enchantment Creature — Saga Harpy {3}{W},
// 3/3 (Final Fantasy, #4462):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Aerial Blast — This creature deals 4 damage to target tapped
//	 creature an opponent controls.
//	 II, III — Slipstream — Another target creature you control gets
//	 +1/+0 and gains flying until end of turn.
//	 Flying"
//
// A Saga creature. Chapter I is Momo's target (a tapped creature an
// opponent controls) with the Saga dealing the 4 damage. Chapters II
// and III are one until-end-of-turn effect at one timestamp, +1/+0 and
// flying, on another creature you control (Another excludes this
// object).
//
// No simplification.
func init() {
	const name = "Summon: Primal Garuda"
	slipstream := func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return effects.UntilEndOfTurn(ctx, id, nil, name+" — Slipstream: +1/+0 and flying",
			game.ModifyPTMod(1, 0), game.AddKeywordsMod("flying"))
	}
	another := func() *game.TargetSpec {
		return effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))
	}
	effects.Register(effects.Spec{
		OracleID:        "aaf9e711-94e1-4869-81b0-baa6380ee873",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "Aerial Blast — 4 damage to target tapped creature an opponent controls"),
				effects.TargetCreature("target tapped creature an opponent controls", effects.OpponentControls(), effects.B751Tapped()),
				effects.FourDamageToFirstLegalTarget),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "Slipstream — another target creature you control gets +1/+0 and gains flying until end of turn"),
				another(), slipstream),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "Slipstream — another target creature you control gets +1/+0 and gains flying until end of turn"),
				another(), slipstream),
		},
	})
}
