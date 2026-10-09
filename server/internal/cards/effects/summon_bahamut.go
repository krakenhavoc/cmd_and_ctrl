package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Summon: Bahamut — Enchantment Creature — Saga Dragon {9}, 9/9:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, II — Destroy up to one target nonland permanent.
//	 III — Draw two cards.
//	 IV — Mega Flare — This creature deals damage equal to the total
//	 mana value of other permanents you control to each opponent.
//	 Flying"
//
// A Saga creature, Summon: Alexander's shape (ChapterTrigger). Chapters
// I and II target "up to one", so a chapter with nothing worth hitting
// still resolves.
//
// Chapter IV counts as it resolves (CR 608.2h): every OTHER permanent
// you control, each at its mana value. A land is 0 (CR 202.3a), a
// token that is no copy is 0, and a face-down permanent is 0. The
// Saga is the source of the damage, and its last-known information
// stands in if it has already left (a Saga is sacrificed only after
// its last chapter ability has left the stack, CR 714.4).
//
// No simplification.
func init() {
	destroyUpToOne := func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	upToOneNonland := func() *game.TargetSpec {
		return TargetPermanent("up to one target nonland permanent", Nonland()).WithCount(0, 1)
	}
	Register(Spec{
		OracleID:        "fe9f6825-597b-4d80-ab7c-f4ee4e824b6f",
		Name:            "Summon: Bahamut",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel("Summon: Bahamut", 1, "destroy up to one target nonland permanent"), upToOneNonland(), destroyUpToOne),
			ChapterTriggerTargeting(2, SagaChapterLabel("Summon: Bahamut", 2, "destroy up to one target nonland permanent"), upToOneNonland(), destroyUpToOne),
			ChapterTrigger(3, SagaChapterLabel("Summon: Bahamut", 3, "draw two cards"), func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			}),
			ChapterTrigger(4, SagaChapterLabel("Summon: Bahamut", 4, "Mega Flare — damage equal to the total mana value of other permanents you control to each opponent"), summonBahamutMegaFlare),
		},
	})
}

// summonBahamutMegaFlare is chapter IV, counted as it resolves.
func summonBahamutMegaFlare(g *game.Game, item *game.StackItem) error {
	total := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == item.SourceCardID || c.Controller != item.Controller || c.FaceDown {
			continue
		}
		total += c.ManaValue()
	}
	if total <= 0 {
		return nil
	}
	return damageToEachOpponent(g, item, total)
}
