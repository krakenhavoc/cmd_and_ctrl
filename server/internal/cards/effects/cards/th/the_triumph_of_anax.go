package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Triumph of Anax — Enchantment — Saga {2}{R} (Theros Beyond Death, #5514):
//
//	"I, II, III — Until end of turn, target creature gains trample and gets +X/+0,
//	 where X is the number of lore counters on this Saga.
//	 IV — Target creature you control fights up to one target creature you don't
//	 control."
//
// X is read as the chapter resolves (CR 608.2h) off the Saga's lore counters, or
// off its last-known counters if it has already left. Chapter IV is two target
// clauses, the second optional: the fight happens only while both are still legal
// (CR 701.14b), and with no second target nothing fights.
//
// No simplification.
func init() {
	const name = "The Triumph of Anax"
	pump := func(n int) game.TriggeredAbility {
		return effects.ChapterTriggerTargeting(n, effects.SagaChapterLabel(name, n, "until end of turn, target creature gains trample and gets +X/+0, where X is the number of lore counters on this Saga"),
			effects.TargetCreature("target creature"), triumphOfAnaxPump)
	}
	effects.Register(effects.Spec{
		OracleID:     "eea6ba94-8755-45c0-94ea-cde68e4a1e7c",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			pump(1), pump(2), pump(3),
			effects.ChapterTriggerTargeting(4, effects.SagaChapterLabel(name, 4, "target creature you control fights up to one target creature you don't control"),
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("up to one target creature you don't control", effects.OpponentControls()).WithCount(0, 1)),
				effects.SlotZeroFightsSlotOne),
		},
	})
}

func triumphOfAnaxPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target := effects.FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	x := 0
	if saga, ok := g.LookupCardForEffect(item.SourceCardID); ok {
		x = saga.Counters[game.CounterLore]
	}
	if err := (effects.BoostUntilEOT{Target: target, Power: x, Label: "The Triumph of Anax — +X/+0"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"trample"}, Label: "The Triumph of Anax — trample"}.Apply(ctx)
}
