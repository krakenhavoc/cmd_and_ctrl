package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Founding the Third Path — Enchantment — Saga {1}{U} (Dominaria United):
//
//	"Read ahead (Choose a chapter and start with that many lore
//	 counters. Add one after your draw step. Skipped chapters don't
//	 trigger. Sacrifice after III.)
//	 I — You may cast an instant or sorcery spell with mana value 1 or 2
//	 from your hand without paying its mana cost.
//	 II — Target player mills four cards.
//	 III — Exile target instant or sorcery card from your graveyard. Copy
//	 it. You may cast the copy."
//
// Read ahead is the engine's (CR 702.155). Chapter I is a free cast from
// the hand as the chapter resolves (the resolution_cast offer Kari Zev's
// Expertise uses), narrowed to instants and sorceries with mana value 1
// or 2; mana value 0 is not offered. Chapter III exiles the still-legal
// target and, for whatever really reached exile, copies the card and offers
// a cast of the copy paying its mana cost (CR 707.12): a copy not cast
// ceases to exist, and the exiled original stays in exile.
//
// No simplification.
func init() {
	const name = "Founding the Third Path"
	free := ChapterTrigger(1, SagaChapterLabel(name, 1, "you may cast an instant or sorcery spell with mana value 1 or 2 from your hand without paying its mana cost"),
		func(g *game.Game, item *game.StackItem) error {
			offerFreeHandCast(NewContext(g, item), item.Controller, foundingThirdPathFreeCast,
				"an instant or sorcery spell with mana value 1 or 2", foundingThirdPathSpell())
			return nil
		})
	mill := ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "target player mills four cards"),
		TargetPlayer("target player"),
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					return MillCards{Player: t.ID, N: 4}.Apply(ctx)
				}
			}
			return nil
		})
	copyCast := ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "exile target instant or sorcery card from your graveyard, copy it, and you may cast the copy"),
		TargetCardInGraveyard("target instant or sorcery card in your graveyard", YouOwn(), Or(Instant(), Sorcery())),
		foundingThirdPathExileAndCopy)
	Register(Spec{
		OracleID:        "0eff910a-712d-4652-8b3b-9eddf63eb891",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordReadAhead},
		Triggered:       []game.TriggeredAbility{free, mill, copyCast},
	})
}

var foundingThirdPathFreeCast = registerFreeHandCastOffer("choose-cards/founding-the-third-path-cast", "Founding the Third Path")

// foundingThirdPathSpell is "an instant or sorcery spell with mana value 1
// or 2": at most 2 on the permission, and 0 refused by the card filter.
func foundingThirdPathSpell() freeCastSpell {
	return freeInstantOrSorcery().withManaValueAtMost(2).withManaValueAtLeast(1)
}

// foundingThirdPathExileAndCopy exiles the still-legal target and offers a
// cast of a copy of whatever really reached exile.
func foundingThirdPathExileAndCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(foundingThirdPathCopyExiled))
}

// foundingThirdPathCopyExiled copies what reached exile and offers the cast.
var foundingThirdPathCopyExiled = CardTail("founding-the-third-path/copy-exiled", func(ctx *Context, o game.TailOutcome) error {
	return CopyAndCast{Cards: o.IDs}.Apply(NewContext(ctx.Game, ctx.Item))
})
