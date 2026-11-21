package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Befriending the Moths // Imperial Moth — a transforming Saga (Kamigawa:
// Neon Dynasty, #5493, S82).
//
// Front face, Enchantment — Saga {3}{W}:
//
//	"I, II — Target creature you control gets +1/+1 and gains flying until
//	 end of turn.
//	 III — Exile this Saga, then return it to the battlefield transformed
//	       under your control."
//
// Back face, Enchantment Creature — Insect, 2/4: "Flying".
//
// The target is chosen as the chapter goes on the stack and re-checked as
// it resolves (CR 608.2b); a chapter with no creature to target is removed
// (CR 603.3d).
//
// No simplification.
const BefriendingTheMothsOracleID = "2ee5f5ad-2f16-40d9-831a-2aefece31b36"

func init() {
	const name = "Befriending the Moths"
	boost := func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		target := effects.FirstLegalBattlefieldTarget(ctx)
		if target == [16]byte{} {
			return nil
		}
		if err := (effects.BoostUntilEOT{Target: target, Power: 1, Toughness: 1, Label: name + " — +1/+1"}).Apply(ctx); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"flying"}, Label: name + " — flying"}.Apply(ctx)
	}
	chapter := func(n int) game.TriggeredAbility {
		return effects.ChapterTriggerTargeting(n, effects.SagaChapterLabel(name, n, "target creature you control gets +1/+1 and flying"),
			effects.TargetCreature("target creature you control", effects.YouControl()), boost)
	}
	effects.Register(effects.Spec{
		OracleID:     BefriendingTheMothsOracleID,
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chapter(1), chapter(2),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        BefriendingTheMothsOracleID + "#1",
		Name:            "Imperial Moth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
