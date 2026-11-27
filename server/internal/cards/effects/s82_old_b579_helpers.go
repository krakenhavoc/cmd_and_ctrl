package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b579_helpers.go — shared pieces of Ravnica: City of Guilds batch 579
// (#5600).

// OldB579CounterTheSpellUnlessDecline is Blood Funnel's "or else": the payer
// declined, or could not sacrifice, so the spell the prompt carries (Carry[0],
// its stack ID) is countered if it is still on the stack. Keyed: on disk, never
// renamed.
var OldB579CounterTheSpellUnlessDecline = PayUnlessElse("pay-unless/old-b579-blood-funnel-counter",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 || ctx.Game.StackItemForEffect(r.Carry[0]) == nil {
			return nil
		}
		return CounterTarget{StackID: r.Carry[0]}.Apply(ctx)
	})

// OldB579CounterUnlessYouSacrificeACreature is the resolution of "counter that
// spell unless you sacrifice a creature": the triggering spell is the cast
// event's, read off the trigger; its caster, the ability's controller, is asked
// whether to sacrifice a creature of their choice, and a "no", or a "yes" with no
// creature to sacrifice (CR 118.3), counters it. A spell that has left the stack
// asks nothing.
func OldB579CounterUnlessYouSacrificeACreature(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		spell := ctx.Trigger().Event.CardID
		if spell == uuid.Nil || g.StackItemForEffect(spell) == nil {
			return nil
		}
		words, action := SacrificePayment(1, "creature", "creatures", QueryType("creature")).Prompt()
		a := *action
		return UpkeepPayUnless{
			Chooser:  item.Controller,
			Cost:     words,
			Action:   &a,
			Question: label + " — sacrifice a creature, or counter the spell?",
			Else:     OldB579CounterTheSpellUnlessDecline,
			Carry:    []uuid.UUID{spell},
		}.Apply(ctx)
	}
}

// OldB579HandToBottomThenDraw is the resolution of "put the cards in your hand on
// the bottom of your library in any order, then draw that many cards" (Mindmoil,
// Arjun's text): the controller's whole hand is the pile, ordered by them; the
// draw is as many as were really put, once placed (a keyed continuation, so the
// order prompt is a restore point). An empty hand does nothing.
func OldB579HandToBottomThenDraw(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		hand := AllHandCardIDs(g, item.Controller)
		if len(hand) == 0 {
			return nil
		}
		return PutInLibraryInAnyOrder{
			Cards:     hand,
			From:      game.ZoneHand,
			Placement: game.LibraryPlaceBottom,
			Label:     label + " — put the cards in your hand on the bottom of your library in any order",
			ThenTail:  ctx.TailN(oldB579DrawThatMany, len(hand)),
		}.Apply(ctx)
	}
}

// oldB579DrawThatMany draws as many cards as the tail's N. An on-disk key.
var oldB579DrawThatMany = CardTail("old-b579/draw-that-many", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	return DrawCards{N: o.Tail.N}.Apply(ctx)
})

// OldB579DrawEqualToSacrificedPower is Twisted Justice's tail: the caster draws
// cards equal to the power of the creature the target player sacrificed, as it
// last existed (CR 608.2h). The carry is the caster, then the victim. On disk:
// never renamed.
var OldB579DrawEqualToSacrificedPower = game.RegisterPromptRunThen("prompt-run/old-b579-draw-sacrificed-power", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) < 2 {
		return nil
	}
	me, victim := r.Carry[0], r.Carry[1]
	ids := r.Sacrifices().By(victim)
	if len(ids) == 0 {
		return nil
	}
	power := DepartedCreaturePower(g, ids[0])
	if power <= 0 {
		return nil
	}
	return DrawCards{Player: me, N: power}.Apply(KeyedContext(g, me, r.Source))
})

// OldB579EyeOfTheStormResolve is the resolution of "exile it. Then that player
// copies each instant or sorcery card exiled with this enchantment. For each copy,
// the player may cast the copy without paying its mana cost" (Eye of the Storm).
// The spell the trigger names (the cast event's card) is exiled from the stack
// with the enchantment's "exiled with" link (CR 607.2a), then, once it has really
// landed, every instant and sorcery card linked to this enchantment is copied
// (CR 707.12) and offered to the CASTER, not the enchantment's controller. A spell
// that has left the stack is not exiled, and the copying still happens.
func OldB579EyeOfTheStormResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spell := ctx.Trigger().Event.CardID
	if spell != uuid.Nil && g.StackItemForEffect(spell) != nil {
		if link, ok := ctx.SourceRef(); ok {
			return g.ExileLinkedKeyedForEffect(link, []uuid.UUID{spell}, ctx.ItemTail(oldB579EyeOfTheStormCopies))
		}
	}
	return oldB579EyeOfTheStormCopyAll(ctx)
}

// oldB579EyeOfTheStormCopies is the "then" of the exile: copy and offer. An
// on-disk key: never renamed, never reused.
var oldB579EyeOfTheStormCopies = CardTail("old-b579/eye-of-the-storm-copies", func(ctx *Context, _ game.TailOutcome) error {
	return oldB579EyeOfTheStormCopyAll(ctx)
})

func oldB579EyeOfTheStormCopyAll(ctx *Context) error {
	link, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range ctx.Game.ExiledWithForEffect(game.PermissionCardRef(link)) {
		if c.IsInstant() || c.IsSorcery() {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return CopyAndCast{Cards: ids, Player: ctx.Trigger().Event.Actor, Free: true}.Apply(ctx)
}
