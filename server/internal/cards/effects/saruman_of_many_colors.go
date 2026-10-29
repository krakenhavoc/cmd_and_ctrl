package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saruman of Many Colors — Legendary Creature — Avatar Wizard {3}{W}{U}{B},
// 5/4 (The Lord of the Rings: Tales of Middle-earth):
//
//	"Ward—Discard an enchantment, instant, or sorcery card.
//	 Whenever you cast your second spell each turn, each opponent mills two
//	 cards. When one or more cards are milled this way, exile target
//	 enchantment, instant, or sorcery card with equal or lesser mana value
//	 than that spell from an opponent's graveyard. Copy the exiled card. You
//	 may cast the copy without paying its mana cost."
//
// The ward is Graveyard Trespasser's "Ward—Discard a card" narrowed to a kind
// of card (WardDiscardKind): the spell's controller must hold at least one
// enchantment, instant or sorcery card to pay, discards one of their choice,
// and is refused any other card; with none in hand the spell is countered at
// once (CR 118.3).
//
// The trigger is Breeches' second-spell condition (the engine's per-turn
// cast tally). "that spell's mana value" is read when the trigger resolves,
// while the spell is normally still on the stack (CR 202.3e, so a spell cast
// with X counts it), and rides the mill's tail. The mills are Dread Summons'
// chain (millEachPlayerInTurn), one opponent at a time in turn order, and
// the tail is told every card that reached a graveyard: with none milled the
// trigger ends there (CR 603.12), since "when one or more cards are milled
// this way" needs one. Otherwise the reflexive trigger goes on the stack once
// above its parent, and chooses its target then (CR 603.12a): an enchantment,
// instant or sorcery card in an opponent's graveyard (any, not only a milled
// card) with mana value at most the carried one.
//
// The rest is Kaervek's: the target is exiled (the cards that really reached
// exile are the continuation's), CopyAndCast copies it and offers the free
// cast while the trigger is still resolving (CR 707.12, 608.2g), and a copy
// not cast ceases to exist. The copy is an exact copy, so the catalog
// resolves it as the original.
//
// No simplification.
func init() {
	const label = "Saruman of Many Colors — each opponent mills two cards"
	Register(Spec{
		OracleID:     "c2d04138-cc17-4dd0-9a4c-9c87df72ed01",
		Name:         "Saruman of Many Colors",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(sarumanWardKind.Cost(), "Saruman of Many Colors — ward, discard an enchantment, instant, or sorcery card"),
			On(game.EventCast, YouCastYourSecondSpellEachTurn, label, sarumanMill),
		},
	})
}

// sarumanWardKind is the card the ward asks for.
var sarumanWardKind = NewWardDiscardKind("enchantment-instant-sorcery", "an enchantment, instant, or sorcery card",
	func(c game.Card) bool { return c.IsEnchantment() || c.IsInstant() || c.IsSorcery() })

// sarumanMill mills two cards from each opponent, carrying the second
// spell's mana value to the tail.
func sarumanMill(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	tail := ctx.ItemTail(sarumanMilled)
	tail.N = spellManaValueForEffect(g, item.Trigger.Event.CardID)
	return millEachPlayerInTurn(ctx, ctx.Opponents(), constantMill(2), tail)
}

// sarumanMilled is the clause after the mills: with at least one card milled,
// the reflexive trigger is raised, its ceiling the spell's mana value.
var sarumanMilled = CardTail("saruman-of-many-colors/after-mill", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ReflexiveTrigger{
		Label:  "Saruman of Many Colors — exile target enchantment, instant, or sorcery card from an opponent's graveyard, copy it and cast the copy",
		Body:   sarumanExileBody,
		Params: game.EffectParams{Amount: o.Tail.N},
	}.Apply(ctx)
})

// sarumanExileBody is the reflexive trigger's registration: the target is
// built as the trigger goes on the stack, from the carried mana value.
var sarumanExileBody = game.ReflexiveBody("saruman-of-many-colors/exile-and-copy", simpleBody(sarumanExile),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetCardInGraveyard("target enchantment, instant, or sorcery card with mana value "+strconv.Itoa(p.Amount)+" or less from an opponent's graveyard",
			Not(YouOwn()), Or(Enchantment(), Instant(), Sorcery()), ManaValueLE(p.Amount))
	})

// sarumanExile exiles the chosen card, then copies it and offers the cast.
func sarumanExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return g.ExileCardsKeyedForEffect([]uuid.UUID{t.ID}, ctx.ItemTail(sarumanCopyExiled))
	}
	return nil
}

// sarumanCopyExiled copies what reached exile and offers the free cast.
var sarumanCopyExiled = CardTail("saruman-of-many-colors/copy-exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return CopyAndCast{Cards: o.IDs, Free: true}.Apply(ctx)
})
