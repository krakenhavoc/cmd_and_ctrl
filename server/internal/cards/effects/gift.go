package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gift.go — the gift keyword (CR 702.174, ADR 0089), both halves.
//
// Its own file for the reason offspring.go is one: concurrent card
// batches collide on shared files, and gift is one mechanic whose cost
// half and gift half are meaningless apart.
//
// "Gift a [something]" is two abilities (CR 702.174a):
//
//   - a COST: "As an additional cost to cast this spell, you may choose
//     an opponent." It rides ADR 0073's optional-cost machinery as one
//     more game.AdditionalCost, with ChoosesOpponent set; the engine
//     announces it at CR 601.2b, validates the opponent, records it on
//     PaidCost.GiftOpponent, copies it (CR 707.10) and carries it onto
//     the permanent (CR 400.7d). The engine knows nothing else about
//     gift.
//   - the GIFT: on an instant or sorcery, "if this spell's gift cost was
//     paid, [effect]", which happens BEFORE the spell's other abilities
//     (CR 702.174j); on a permanent, "when this permanent enters, if its
//     gift cost was paid, [effect]" (CR 702.174b). The [effect] is fixed
//     by the [something] (CR 702.174d–i).
//
// A card declares gift with ONE field, and buildDef grows both halves
// from it — the optional cost, the pre-OnResolve gift for a spell and
// the entry trigger for a permanent — so a gift card cannot forget one
// or spell them two ways:
//
//	Gift: GiftACard(),                                                  // Dawn's Truce
//	Gift: GiftATappedFish().Instead(TargetPermanent(…, Nonland(), …)),  // Into the Flood Maw
//
// The card's own "if the gift was promised" branches read
// ctx.GiftPromised() at resolution, or card.GiftPromised() from a
// permanent's own trigger.

// Gift is one printed "Gift a [something]" declaration. Build it with
// the GiftA… constructors below, never by hand: the constructor is what
// ties the printed noun to the effect CR 702.174 defines for it.
type Gift struct {
	// Label is the keyword as printed — "Gift a card", "Gift a tapped
	// Fish" — shown on the client's toggle and in the bot's move label.
	Label string

	// give is the CR 702.174d–i effect, applied for the chosen
	// opponent `to`. Never nil on a constructed Gift.
	give func(ctx *Context, to uuid.UUID) error

	// Targets is the spell's WHOLE target clause when the gift was
	// promised (CR 702.174m) — set with Instead. Nil when the promise
	// leaves the printed clause alone.
	Targets *game.TargetSpec
}

// Instead declares the target clause a promised gift gives the spell,
// replacing the printed one: Long River's Pull's "Counter target
// creature spell. If the gift was promised, instead counter target
// spell" is
//
//	Targets: TargetSpell("target creature spell", Creature()),
//	Gift:    GiftACard().Instead(TargetSpell("target spell")),
//
// A clause that only exists when the gift was promised (Mind Spiral's
// "if the gift was promised, tap target creature an opponent
// controls") is the printed list with that clause appended
// (TargetPlayer(…).Then(TargetCreature(…))) — CR 702.174m's "chooses
// those targets only if the gift was promised".
//
// The engine announces the clause at CR 601.2c after the promise at
// 601.2b, checks it again at resolution (CR 608.2b), and stamps its
// legal set on the offer so the client's picker and the bot see the
// clause the cast will be judged under.
func (g *Gift) Instead(t *game.TargetSpec) *Gift {
	out := *g
	out.Targets = t
	return &out
}

// GiftACard is CR 702.174e: "The chosen player draws a card."
func GiftACard() *Gift {
	return &Gift{Label: "Gift a card", give: func(ctx *Context, to uuid.UUID) error {
		return DrawCards{Player: to, N: 1}.Apply(ctx)
	}}
}

// GiftAFood is CR 702.174d: "The chosen player creates a Food token."
// The token is the catalog Food (ADR 0083), so its sacrifice ability
// works for the opponent who got it.
func GiftAFood() *Gift {
	return &Gift{Label: "Gift a Food", give: func(ctx *Context, to uuid.UUID) error {
		return CreateToken{Controller: to, Template: FoodToken(), N: 1}.Apply(ctx)
	}}
}

// GiftATappedFish is CR 702.174f: "The chosen player creates a tapped
// 1/1 blue Fish creature token." Tapped through the token entry
// options rather than tapped a beat later, so nothing watching for an
// untapped creature entering sees one.
func GiftATappedFish() *Gift {
	return &Gift{Label: "Gift a tapped Fish", give: func(ctx *Context, to uuid.UUID) error {
		return CreateTokenAdvanced{Controller: to, Spec: Token(TokenCard("1/1 blue Fish")).EntersTapped(), N: 1}.Apply(ctx)
	}}
}

// GiftATreasure is CR 702.174h: "The chosen player creates a Treasure
// token." The catalog Treasure (ADR 0083).
func GiftATreasure() *Gift {
	return &Gift{Label: "Gift a Treasure", give: func(ctx *Context, to uuid.UUID) error {
		return CreateToken{Controller: to, Template: TreasureToken(), N: 1}.Apply(ctx)
	}}
}

// giftCost is the cost half as the engine reads it: an optional
// additional cost keyed game.GiftKey whose payment is choosing an
// opponent. Appended LAST to the card's optional costs by buildDef, so
// a card that also prints a kicker keeps its kicker's index.
func (g *Gift) cost() game.AdditionalCost {
	return game.AdditionalCost{
		Optional:        true,
		Key:             game.GiftKey,
		ChoosesOpponent: true,
		Label:           g.Label,
		Targets:         g.Targets,
	}
}

// giveTo applies the gift to `to`. A recipient who has left the game
// since the promise (CR 800.4a) gets nothing, and that needs no check
// here: the draw and token-creation paths the constructors use already
// refuse a player who has left, which TestGiftGoesToAPlayerWhoLeft…
// pins. The gift was still promised, and the card's own "if the gift
// was promised" branches still apply — CR 702.174k is about the
// declaration, not the delivery.
func (g *Gift) giveTo(ctx *Context, to uuid.UUID) error {
	if to == uuid.Nil {
		return nil
	}
	return g.give(ctx, to)
}

// wrapResolve is the instant-and-sorcery half (CR 702.174b, j): when
// the gift was promised, the gift happens FIRST, then the card's own
// OnResolve. A permanent spell's gift is its entry trigger instead, so
// the wrapper stands aside for one — which is also why buildDef can
// install both halves on every gift card without knowing its type.
//
// A spell countered, or one that fizzles because every target left
// (CR 608.2b), never reaches this function, which is CR 702.174j's
// "the gift effect doesn't happen" for free.
func (g *Gift) wrapResolve(onResolve func(*game.StackItem, *Context) error) func(*game.Game, *game.StackItem) error {
	return func(gm *game.Game, item *game.StackItem) error {
		ctx := NewContext(gm, item)
		if item.Paid.GiftPromised() && !resolvingPermanentSpell(gm, item) {
			if err := g.giveTo(ctx, item.Paid.GiftOpponent); err != nil {
				return err
			}
		}
		if onResolve == nil {
			return nil
		}
		return onResolve(item, ctx)
	}
}

// resolvingPermanentSpell reports whether the spell resolving as `item`
// is a permanent spell — the one whose gift is an entry trigger rather
// than a spell ability.
func resolvingPermanentSpell(g *game.Game, item *game.StackItem) bool {
	c, ok := g.LookupCardForEffect(item.ID)
	return ok && c.IsPermanent()
}

// entryTrigger is the permanent half, CR 702.174b: "When this permanent
// enters, if its gift cost was paid, [effect]." The "if" is a CR 603.4
// intervening if on the permanent's own record, so a creature cast
// without the promise — or reanimated, or a token — puts nothing on
// the stack.
//
// The recipient is read in Build, off the permanent that just entered,
// and carried on the item as Params.Player: by resolution the permanent
// may be gone and its record with it, but the chosen player is still
// the chosen player. The Build only fills that in; the effect is the
// row's (ADR 0041 P9), so a gift waiting on the stack is a restore
// point.
func (g *Gift) entryTrigger(cardName string) game.TriggeredAbility {
	label := cardName + " — " + g.Label
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventETB},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID && source.GiftPromised()
		},
		Key: label,
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Player = source.Provenance.GiftOpponent
			return item
		},
		Effect: func(gm *game.Game, item *game.StackItem) error {
			return g.giveTo(NewContext(gm, item), item.Params.Player)
		},
	}
}

// GiftPromised is CR 702.174k's "if the gift was promised" for the
// resolving spell — the read every gift card's own text branches on:
//
//	if ctx.GiftPromised() { … the bonus … }
//
// False for an ability item and for any spell cast without the promise.
func (c *Context) GiftPromised() bool {
	return c.Item != nil && c.Item.Paid.GiftPromised()
}

// GiftOpponent is the opponent the resolving spell's gift was promised
// to, or uuid.Nil. Most cards never need it — the gift itself is
// delivered by the keyword — but "the chosen player" can appear in a
// card's own text too.
func (c *Context) GiftOpponent() uuid.UUID {
	if c.Item == nil {
		return uuid.Nil
	}
	return c.Item.Paid.GiftOpponent
}
