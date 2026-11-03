package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// socB26DestroyTheTargetAndGainLife is "Destroy target <permanent>. You gain
// n life." as a spell's whole body (Terminal Criticism, Grapple with Death):
// the target is re-checked as the spell resolves (CR 608.2b), and the life is
// its own sentence, so it is gained whether or not the destroy did anything.
func socB26DestroyTheTargetAndGainLife(n int) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
			return err
		}
		return GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
	}
}

// socB26DestroyThenItsControllerCreates is "Destroy target creature. Its
// controller creates <token>" as a spell's whole body (Harsh Annotation): the
// controller is read BEFORE the destroy, so the token goes to whoever held
// the creature when the spell resolved, and it is made whether or not the
// destroy did anything (the creature was indestructible, say).
func socB26DestroyThenItsControllerCreates(token string) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		target := item.Targets[0].ID
		controller, ok := controllerOfTarget(ctx, target)
		if !ok {
			return nil
		}
		if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
			return err
		}
		return CreateToken{Controller: controller, Template: TokenCard(token), N: 1}.Apply(ctx)
	}
}

// socB26HeatedArgumentDamage is Heated Argument's first sentence: 6 damage to
// the announced creature, if it is still a legal target.
func socB26HeatedArgumentDamage(ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: id, Amount: 6}.Apply(ctx)
}

// socB26HeatedArgumentOffer is "You may exile a card from your graveyard": a
// pick of up to one card from the caster's graveyard, carrying the creature's
// controller (read here, while the creature is still the one that was dealt
// damage) for the "if you do" half. With an empty graveyard nothing is asked.
func socB26HeatedArgumentOffer(ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	controller, ok := controllerOfTarget(ctx, id)
	if !ok {
		return nil
	}
	var cards []uuid.UUID
	if p := ctx.PlayerByID(ctx.Controller()); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			cards = append(cards, c.InstanceID)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Item.Controller,
		Source:   ctx.Item.SourceCardID,
		Question: "Heated Argument — you may exile a card from your graveyard (if you do, 2 damage to that creature's controller)",
		Cards:    cards,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  heatedArgumentExile,
		Carry:    []uuid.UUID{controller},
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// heatedArgumentExile is the exile and, if a card was exiled, the 2 damage to
// the creature's controller. An on-disk key: never renamed, never reused.
var heatedArgumentExile = askedPick("choose-cards/heated-argument-exile",
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 1 || len(r.Carry) != 1 {
			return nil
		}
		if err := (ExileTarget{Target: r.Picked[0]}).Apply(ctx); err != nil {
			return err
		}
		return DealDamage{Source: ctx.Source(), Target: r.Carry[0], Amount: 2}.Apply(ctx)
	})

// soc_b26_helpers.go — shared pieces of Secrets of Strixhaven batch 26
// (#4456).

// socB26CardsLeaveYourGraveyard is "whenever one or more cards leave your
// graveyard" with the printed effect: a batch trigger (CR 603.2c) over cards
// of the controller's that leave a graveyard, by a zone move or by being cast
// from it (leftGraveyardEvents, whenCardsLeaveYourGraveyard). Cards that
// leave as one event trigger it once; separate events trigger it separately.
// Garrison Excavator and Hardened Academic.
func socB26CardsLeaveYourGraveyard(label string, effect Effect) game.TriggeredAbility {
	return OncePerBatch(OnAny(leftGraveyardEvents, whenCardsLeaveYourGraveyard, label, effect))
}

// socB26YouCastASpellWithXInItsManaCost is "whenever you cast a spell with
// {X} in its mana cost": the source's controller cast a spell whose own
// printed mana cost has an {X}. An {X} in an additional or alternative cost
// does not count (the rulings on Unbound Flourishing and Nev read it the
// same way).
func socB26YouCastASpellWithXInItsManaCost(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller && socB9HasXInManaCost(g, ev.CardID)
}
