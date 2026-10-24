package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_misc_a_helpers.go — shared bodies for the
// fra-misc-a slice of the Reality Fracture set (surveil cards and lands).
// Every package-level name carries the rfMiscA prefix.

// rfMiscAFinalityExile is the second half of a finality counter (CR
// 122.1b): "if a creature with a finality counter on it would die,
// exile it instead". Grim Repriser spells the same replacement out
// inline; this is the reusable body, written for a card's own
// Replacements slot.
func rfMiscAFinalityExile(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove && ev.OldZone == game.ZoneBattlefield &&
				ev.NewZone == game.ZoneGraveyard && src != nil && ev.CardID == src.InstanceID &&
				src.Counters[rfCreatureBFinalityCounter] > 0
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.NewZone = game.ZoneExile
			ev.NewZoneOwner = uuid.Nil
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// rfMiscAEyeOfJaceUpkeep is Eye of Jace's upkeep trigger: surveil 1,
// then — once the surveil has settled, because the card may have gone
// to the graveyard — if the controller's graveyard holds seven or more
// cards, sacrifice the Eye, deal 2 damage to each opponent and gain 2
// life. The damage and the life are not conditional on the sacrifice
// (CR 608.2c: "it" is the Eye by last-known information), so an Eye
// removed in response to the trigger still pays out.
func rfMiscAEyeOfJaceUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return Surveil{N: 1, ThenTail: ctx.ItemTail(rfMiscAEyeOfJaceCheck)}.Apply(ctx)
}

// rfMiscAEyeOfJaceCheck is the seven-card graveyard check and its payout,
// once the surveil has settled; the item rides the tail. An on-disk key.
var rfMiscAEyeOfJaceCheck = CardTail("eye-of-jace/graveyard-check", func(ctx *Context, o game.TailOutcome) error {
	g, controller, eye, it := ctx.Game, o.Tail.Player, o.Tail.Source, ctx.Item
	if b31GraveyardSize(g, controller) < 7 {
		return nil
	}
	if onBattlefield(g, eye) && !sourceIsNewObject(g, it) {
		if err := (SacrificePermanent{Target: eye}).Apply(ctx); err != nil {
			return err
		}
	}
	if err := damageToEachOpponent(g, it, 2); err != nil {
		return err
	}
	return GainLife{Player: controller, Amount: 2}.Apply(ctx)
})

// rfMiscAConfidantEndStep is Enlightened Confidant's end-step trigger.
// The intervening "if you gained life this turn" is asked again as it
// resolves (CR 603.4). The amount is read once, then the surveil runs;
// if the card that was on top went to the graveyard THIS WAY and its
// mana value is no greater than the life gained, it goes to hand.
func rfMiscAConfidantEndStep(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	gained := b15LifeGainedThisTurn(g, controller)
	if gained <= 0 {
		return nil
	}
	p := g.PlayerByIDForEffect(controller)
	if p == nil {
		return nil
	}
	var top uuid.UUID
	if c, err := p.Library.Top(); err == nil {
		top = c.InstanceID
	}
	ctx := NewContext(g, item)
	tail := ctx.ItemTail(rfMiscAConfidantReturn)
	tail.N = gained
	if top != uuid.Nil {
		tail.Carry = []uuid.UUID{top}
	}
	return Surveil{N: 1, ThenTail: tail}.Apply(ctx)
}

// rfMiscAConfidantReturn returns the card that was on top to its owner's
// hand if the surveil put it into the graveyard and its mana value is no
// greater than the life gained (Tail.N); the card rides Tail.Carry. An
// on-disk key.
var rfMiscAConfidantReturn = CardTail("enlightened-confidant/return-top-card", func(ctx *Context, o game.TailOutcome) error {
	g, controller := ctx.Game, o.Tail.Player
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	top := o.Tail.Carry[0]
	p := g.PlayerByIDForEffect(controller)
	if p == nil || !p.Graveyard.Contains(top) {
		return nil
	}
	c, ok := g.LookupCardForEffect(top)
	if !ok {
		return nil
	}
	if mv, ok := g.ManaValueForEffect(c); !ok || mv > o.Tail.N {
		return nil
	}
	return ReturnFromGraveyard{Target: top, Dest: game.ZoneHand}.Apply(ctx)
})

// rfMiscAForestsYouControlOtherThan counts the Forests `controller`
// controls apart from `except` (post-layer land types). `forest` is the
// land type as the counting object's text now reads it (CR 612.2).
func rfMiscAForestsYouControlOtherThan(g *game.Game, controller, except uuid.UUID, forest string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.InstanceID != except && c.HasSubtype(forest) {
			n++
		}
	}
	return n
}

// rfMiscAForestEnteredWithFiveOthers is Roiling Canopy's trigger
// condition, intervening-if included: a Forest you control entered and
// you control at least five other Forests.
func rfMiscAForestEnteredWithFiveOthers(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	forest := source.TextSubtype("Forest")
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok || !c.HasSubtype(forest) {
		return false
	}
	return rfMiscAForestsYouControlOtherThan(g, source.Controller, c.InstanceID, forest) >= 5
}
