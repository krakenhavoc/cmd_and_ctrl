package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Origin of Spider-Man — Enchantment — Saga {1}{W} (Marvel's Spider-Man
// promos, #4460):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Create a 2/1 green Spider creature token with reach.
//	 II — Put a +1/+1 counter on target creature you control. It becomes a
//	      legendary Spider Hero in addition to its other types.
//	 III — Target creature you control gains double strike until end of
//	       turn."
//
// Chapter II's type change has no duration, so it lasts as long as the
// creature is that object (CR 611.2a, 400.7): one record in layer 4 adding
// the Legendary supertype (CR 205.4b) and the Spider and Hero subtypes
// (CR 205.1b), each word read through the Saga's text (CR 612.2). Chapter
// III is the ordinary until-end-of-turn keyword grant. A target that is no
// longer legal as a chapter resolves is left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fbe9056-190f-40a1-bee9-59ebd6f981f5",
		Name:         "Origin of Spider-Man",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel("Origin of Spider-Man", 1, "create a 2/1 green Spider creature token with reach"),
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 green Spider with reach"), N: 1})),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel("Origin of Spider-Man", 2, "a +1/+1 counter on target creature you control; it becomes a legendary Spider Hero"),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				originOfSpiderManChapterTwo),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel("Origin of Spider-Man", 3, "target creature you control gains double strike until end of turn"),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Target:   target,
						Keywords: []string{"double strike"},
						Label:    "Origin of Spider-Man — double strike until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}

// originOfSpiderManChapterTwo is the counter, then "it becomes a legendary
// Spider Hero in addition to its other types" for as long as it is that
// creature.
func originOfSpiderManChapterTwo(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScopedEffectFor{
		Target: target,
		Mods: []game.Mod{
			game.AddSupertypesMod("Legendary"),
			game.AddSubtypesMod(ctx.TextSubtype("Spider"), ctx.TextSubtype("Hero")),
		},
		Duration: g.PinnedTo(game.IndefiniteDuration(), target),
		Label:    "Origin of Spider-Man — a legendary Spider Hero in addition to its other types",
	}.Apply(ctx)
}
