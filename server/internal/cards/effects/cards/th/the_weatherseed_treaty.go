package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Weatherseed Treaty — Enchantment — Saga {2}{G} (Dominaria United):
//
//	"Read ahead (Choose a chapter and start with that many lore
//	 counters. Add one after your draw step. Skipped chapters don't
//	 trigger. Sacrifice after III.)
//	 I — Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle.
//	 II — Create a 1/1 green Saproling creature token.
//	 III — Domain — Target creature you control gets +X/+X and gains
//	 trample until end of turn, where X is the number of basic land types
//	 among lands you control."
//
// Read ahead is the engine's (CR 702.155, game/read_ahead.go). X is
// counted as chapter III resolves (CR 608.2h), so a land that left or
// entered while the chapter was on the stack counts; the target is
// re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	const name = "The Weatherseed Treaty"
	fetch := effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "search for a basic land, put it onto the battlefield tapped"),
		func(g *game.Game, item *game.StackItem) error {
			return effects.SearchLibrary{
				Player:        item.Controller,
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        name + " — a basic land",
			}.Apply(effects.NewContext(g, item))
		})
	fetch.Purpose = game.Purpose{Lands: 1}
	token := effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "create a 1/1 green Saproling"),
		func(g *game.Game, item *game.StackItem) error {
			return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 green Saproling"), N: 1}.Apply(effects.NewContext(g, item))
		})
	token.Purpose = game.Purpose{Tokens: 1, CreatureTokens: 1}
	pump := effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "target creature you control gets +X/+X and trample"),
		effects.TargetCreature("target creature you control", effects.YouControl()), weatherseedPump)
	effects.Register(effects.Spec{
		OracleID:        "7d723a6f-854f-43df-9401-a4eb0c0f639f",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordReadAhead},
		Triggered:       []game.TriggeredAbility{fetch, token, pump},
	})
}

// weatherseedPump is chapter III.
func weatherseedPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := effects.DomainCount(g, item.Controller)
	label := "The Weatherseed Treaty — +X/+X and trample until end of turn"
	if err := (effects.BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: label}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: label}.Apply(ctx)
}
