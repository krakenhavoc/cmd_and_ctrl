package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Tale of Tamiyo — Legendary Enchantment — Saga {2}{U} (Duskmourn):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, II, III — Mill two cards. If two cards that share a card type were
//	 milled this way, draw a card and repeat this process.
//	 IV — Exile any number of target instant, sorcery, and/or Tamiyo
//	 planeswalker cards from your graveyard. Copy them. You may cast any
//	 number of the copies."
//
// Chapters I to III are one loop: a mill of two through the CR 614 window,
// whose keyed tail is told the cards that really landed in the graveyard
// ("milled this way", CR 400.7). Two cards that share a card type draw a
// card and start the mill again from the same tail; fewer than two milled,
// or two with no type in common, ends the story, as the ruling says.
//
// Chapter IV exiles the still-legal picks (any number, none included), copies
// the cards that really reached exile and offers a cast of each copy through
// the resolution cast prompt (CopyAndCast): the printed mana cost is paid
// (an X is chosen as normal), copies not cast cease to exist, and a copy is
// not cast "from exile" in any way that matters (CR 707.12). Chapter IV and
// the Saga's sacrifice (CR 714.4) are the engine's.
func init() {
	const mill = "mill two cards; if two cards that share a card type were milled this way, draw a card and repeat"
	millLoop := func(g *game.Game, item *game.StackItem) error {
		return taleOfTamiyoMill(g, effects.NewContext(g, item))
	}
	targets := effects.TargetCardInGraveyard("any number of target instant, sorcery, and/or Tamiyo planeswalker cards from your graveyard",
		effects.YouOwn(), effects.Or(effects.OfCardType("instant"), effects.OfCardType("sorcery"), effects.And(effects.OfCardType("planeswalker"), effects.HasSubtype("Tamiyo")))).WithCount(0, 0)
	effects.Register(effects.Spec{
		OracleID:     "98a13f7b-ebe7-4a90-a0b9-95dccfff7fa6",
		Name:         "The Tale of Tamiyo",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel("The Tale of Tamiyo", 1, mill), millLoop),
			effects.ChapterTrigger(2, effects.SagaChapterLabel("The Tale of Tamiyo", 2, mill), millLoop),
			effects.ChapterTrigger(3, effects.SagaChapterLabel("The Tale of Tamiyo", 3, mill), millLoop),
			effects.ChapterTriggerTargeting(4, effects.SagaChapterLabel("The Tale of Tamiyo", 4, "exile any number of target instant, sorcery, and/or Tamiyo planeswalker cards from your graveyard, copy them, and cast any number of the copies"),
				targets, taleOfTamiyoExile),
		},
	})
}

// taleOfTamiyoCardTypes are the card types two milled cards may share.
var taleOfTamiyoCardTypes = []string{"artifact", "battle", "creature", "enchantment", "instant", "kindred", "land", "planeswalker", "sorcery"}

// taleOfTamiyoMill mills two cards and hands the landed cards to the loop's
// tail.
func taleOfTamiyoMill(g *game.Game, ctx *effects.Context) error {
	return g.MillToZoneKeyedForEffect(ctx.Controller(), 2, game.ZoneGraveyard, game.MillUntil{}, ctx.ItemTail(taleOfTamiyoMilled))
}

// taleOfTamiyoMilled is "if two cards that share a card type were milled
// this way, draw a card and repeat this process". An on-disk key: never
// renamed, never reused.
//
// It is assigned in init: the tail mills again, and the mill names the tail,
// which a package-level initializer cannot say.
var taleOfTamiyoMilled game.TailRef

func init() {
	taleOfTamiyoMilled = effects.CardTail("then/the-tale-of-tamiyo-milled", func(ctx *effects.Context, o game.TailOutcome) error {
		if len(o.IDs) != 2 || !taleOfTamiyoShareAType(ctx.Game, o.IDs[0], o.IDs[1]) {
			return nil
		}
		if err := ctx.Game.DrawNForEffect(ctx.Controller(), 1); err != nil {
			return err
		}
		return taleOfTamiyoMill(ctx.Game, ctx)
	})
}

// taleOfTamiyoShareAType reports whether two cards share a card type.
func taleOfTamiyoShareAType(g *game.Game, a, b uuid.UUID) bool {
	ca, ok := g.LookupCardForEffect(a)
	if !ok {
		return false
	}
	cb, ok := g.LookupCardForEffect(b)
	if !ok {
		return false
	}
	for _, t := range taleOfTamiyoCardTypes {
		if ca.HasCardType(t) && cb.HasCardType(t) {
			return true
		}
	}
	return false
}

// taleOfTamiyoExile exiles the still-legal picks, then copies what reached
// exile.
func taleOfTamiyoExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(taleOfTamiyoCopy))
}

// taleOfTamiyoCopy copies the exiled cards and offers a cast of each copy,
// paying its mana cost. An on-disk key: never renamed, never reused.
var taleOfTamiyoCopy = effects.CardTail("then/the-tale-of-tamiyo-copy-and-cast", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.CopyAndCast{Cards: o.IDs}.Apply(ctx)
})
