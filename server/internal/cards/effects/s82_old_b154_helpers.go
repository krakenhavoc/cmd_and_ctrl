package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b154_helpers.go — shared bodies for the Core Set 2020 cards of S82
// set OLD, batch 154 (#5518).

// OldB154CavalierDiscardDraw is Cavalier of Flame's enter trigger: "discard any
// number of cards, then draw that many cards". An empty hand asks nothing.
func OldB154CavalierDiscardDraw(g *game.Game, item *game.StackItem) error {
	hand := len(AllHandCardIDs(g, item.Controller))
	if hand == 0 {
		return nil
	}
	return B39MayDiscardThenDraw(hand, false,
		"Cavalier of Flame — discard any number of cards, then draw that many",
		DrawAsManyAsDiscarded)(NewContext(g, item))
}

// OldB154CavalierDeath is Cavalier of Flame's dies trigger: X damage to each
// opponent and each planeswalker they control, X the land cards in its
// controller's graveyard as the trigger resolves.
func OldB154CavalierDeath(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := B17LandCardsInGraveyard(g, ctx.Controller())
	if x <= 0 {
		return nil
	}
	opponents := ctx.Opponents()
	var walkers []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsPlaneswalker() && slices.Contains(opponents, c.Controller) {
			walkers = append(walkers, c)
		}
	}
	return g.DamageInstanceForEffect(func() error {
		for _, opp := range opponents {
			if err := (DealDamage{Source: item.SourceCardID, Target: opp, Amount: x}).Apply(ctx); err != nil {
				return err
			}
		}
		for _, c := range walkers {
			if err := (DealDamage{Source: item.SourceCardID, Target: c.InstanceID, Amount: x}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// oldB154AnotherNamed reports whether `player` controls a creature named
// `name` other than `except`.
func oldB154AnotherNamed(g *game.Game, player, except uuid.UUID, name string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.InstanceID != except && c.IsCreature() && c.Name == name {
			return true
		}
	}
	return false
}

// OldB154IfYouControlAnotherNamed is the intervening "if you control another
// creature named <name>" (CR 603.4) on an enter trigger of that same card: it
// triggers only while the condition holds and does nothing if it has stopped
// holding as the ability resolves.
func OldB154IfYouControlAnotherNamed(name string, t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && oldB154AnotherNamed(g, source.Controller, source.InstanceID, name) && (when == nil || when(ev, source, lki, g))
	}
	return InterveningIf(CondBoard{Name: "you control another creature named " + name, Test: func(ctx *Context) bool {
		return oldB154AnotherNamed(ctx.Game, ctx.Controller(), ctx.Source(), name)
	}}, t)
}

// YouActivateAChandraLoyaltyAbility is "you activate a loyalty ability
// of a Chandra planeswalker" (CR 606.2): the activation event's loyalty bit
// and the activated planeswalker's Chandra subtype through the layered view.
func YouActivateAChandraLoyaltyAbility(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventActivateAbility || !ev.Loyalty || ev.Actor != source.Controller {
		return false
	}
	pw, ok := g.LookupCardForEffect(ev.CardID)
	return ok && pw.IsPlaneswalker() && pw.HasSubtype("Chandra")
}

// oldB154RegulatorPaid is the paid half of Chandra's Regulator's trigger: copy
// the activation named by the triggering event, still on the stack beneath the
// trigger, with the new-targets offer (CR 707.10c). An on-disk key: never
// renamed, never reused.
var oldB154RegulatorPaid = PayUnlessBranches("pay-unless/chandras-regulator-copy-ability",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CopyAbility{
			ItemID:           ctx.Trigger().Event.StackItemID,
			Controller:       ctx.Controller(),
			ChooseNewTargets: true,
		}.Apply(ctx)
	}, nil)

// OldB154RegulatorOffer asks for Chandra's Regulator's {1} as the trigger
// resolves ("you may pay {1}. If you do, copy that ability").
func OldB154RegulatorOffer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if ctx.Trigger().Event.StackItemID == uuid.Nil {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Chandra's Regulator — pay {1} to copy that ability?",
		Then:     oldB154RegulatorPaid,
	}.Apply(ctx)
}

// OldB154RedOrMountainCard is "a Mountain card or a red card" for a discard.
func OldB154RedOrMountainCard(c game.Card) bool {
	return c.HasColorWord("R") || (c.IsLand() && c.HasSubtypeWord("Mountain"))
}

// OldB154AtemsisRevealForSix is Atemsis, All-Seeing's damage trigger as it
// resolves: reveal your hand, and if cards with at least six different mana
// values were revealed, the damaged player loses the game.
func OldB154AtemsisRevealForSix(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if ctx.Trigger().Event.Target == uuid.Nil {
		return nil
	}
	g.RevealHandForEffect(ctx.Controller())
	seen := map[int]bool{}
	for _, id := range AllHandCardIDs(g, ctx.Controller()) {
		if c, ok := g.LookupCardForEffect(id); ok {
			seen[c.ManaValue()] = true
		}
	}
	if len(seen) < 6 {
		return nil
	}
	return LoseTheGame{Player: ctx.Trigger().Event.Target}.Apply(ctx)
}

// OldB154PatternMatcherSearch is Pattern Matcher's enter trigger: you may
// search your library for a card with the same name as another creature you
// control, reveal it, put it into your hand, then shuffle. The names are read
// as the trigger resolves.
func OldB154PatternMatcherSearch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	names := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.InstanceID != ctx.Source() && c.IsCreature() && c.Name != "" {
			names[c.Name] = true
		}
	}
	return SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: func(c game.Card) bool { return names[c.Name] },
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Optional:  true,
		Reason:    "Pattern Matcher — a card with the same name as another creature you control",
	}.Apply(ctx)
}

// oldB154LockboxNamed is Diviner's Lockbox once the name (the tail's Word) is
// known: reveal the top card of your library; if it has the chosen name,
// sacrifice the Lockbox and draw three cards. An on-disk key: never renamed,
// never reused.
var oldB154LockboxNamed = CardTail("then/diviners-lockbox-named", func(ctx *Context, o game.TailOutcome) error {
	revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 1, "Diviner's Lockbox — reveal the top card of your library")
	if len(revealed) == 0 {
		return nil
	}
	top, ok := ctx.Game.LookupCardForEffect(revealed[0])
	if !ok || !game.CardNameMatches(top, o.Tail.Word) {
		return nil
	}
	if OnBattlefield(ctx.Game, ctx.Source()) && !ctx.IsNewSourceObject(ctx.Source()) {
		if err := ctx.Game.SacrificePermanentForEffect(ctx.Source()); err != nil {
			return err
		}
	}
	return ctx.Game.DrawNForEffect(ctx.Controller(), 3)
})

// OldB154LockboxChooseName is Diviner's Lockbox's ability: choose a card name,
// then reveal the top card of your library.
func OldB154LockboxChooseName(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	g.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
		"Diviner's Lockbox — choose a card name", ctx.ItemTail(oldB154LockboxNamed))
	return nil
}
