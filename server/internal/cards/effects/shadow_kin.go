package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shadowKinGrant names the bundle a copy of another card keeps: the
// upkeep ability itself ("except it has this ability", CR 707.9a).
const shadowKinGrant = "shadow-kin/this-ability"

// shadowKinMilled is the cards milled so far this resolution.
const shadowKinMilled = "shadow-kin/milled"

const shadowKinLabel = "Shadow Kin — each player mills three cards; you may exile a creature card from among them; if you do, this creature becomes a copy of it"

// Shadow Kin — Creature — Shapeshifter {3}{U}, 2/2 (Midnight Hunt):
//
//	"Flash
//	 At the beginning of your upkeep, each player mills three cards. You
//	 may exile a creature card from among the cards milled this way. If
//	 you do, this creature becomes a copy of that card, except it has this
//	 ability."
//
// Every player mills three in APNAP order (CR 101.4), waiting for each
// mill to settle, and the cards that landed are recorded (a card a
// replacement exiled instead of milling is still one of "the cards milled
// this way", CR 701.17c, but only a card in a graveyard can be exiled
// from there). The controller then picks a creature card among them or
// none; a pick names the cards of one graveyard, so each owner's are
// offered in turn until one is chosen or all are declined. The exile and the copy are one answer: the copy is of the card's
// copiable values as it was in the graveyard (CR 707.2), has no
// duration (CR 611.2a) and keeps this very ability through a grant
// bundle that names the trigger, so a copy of a copy still mills. A Shadow
// Kin that left or came back as a new object is left alone (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7ecc506e-bed0-4ae2-9d2f-5143ef63f62c",
		Name:            "Shadow Kin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Grants: []AbilityGrant{{
			Key:       shadowKinGrant,
			Text:      "At the beginning of your upkeep, each player mills three cards. You may exile a creature card from among the cards milled this way. If you do, this creature becomes a copy of that card, except it has this ability.",
			Triggered: []game.TriggeredAbility{shadowKinTrigger()},
		}},
		Triggered: []game.TriggeredAbility{shadowKinTrigger()},
	})
}

func shadowKinTrigger() game.TriggeredAbility {
	return AtYourUpkeep(shadowKinLabel, shadowKinMill)
}

// shadowKinMill starts the mills, in APNAP order.
func shadowKinMill(g *game.Game, item *game.StackItem) error {
	return shadowKinMillStep(NewContext(g, item), apnapPlayers(g), 0)
}

// shadowKinMillStep mills three cards for seats[i], or goes on to the pick
// once every seat has. The seats ride the mill's tail as its Carry and the
// index as its N.
func shadowKinMillStep(ctx *Context, seats []uuid.UUID, i int) error {
	if i >= len(seats) {
		return shadowKinPayoff.Run(ctx)
	}
	return ctx.Game.MillToZoneKeyedMovedForEffect(seats[i], 3, game.ZoneGraveyard,
		withN(withCarry(ctx.ItemTail(shadowKinMilledTail), seats...), i))
}

// shadowKinMilledTail records what one seat's mill moved and mills the
// next. Set in init, because the next mill names it again.
var shadowKinMilledTail game.TailRef

func init() {
	shadowKinMilledTail = CardTail("then/shadow-kin-mill-step", func(ctx *Context, o game.TailOutcome) error {
		milled := slices.Clone(ctx.StepRecord(shadowKinMilled).IDs)
		milled = append(milled, o.IDs...)
		ctx.RecordStep(shadowKinMilled, milled)
		return shadowKinMillStep(ctx, o.Tail.Carry, o.Tail.N+1)
	})
}

// shadowKinPayoff offers the creature cards among those milled that are
// in a graveyard now.
var shadowKinPayoff = RegisterThenSteps("then-steps/shadow-kin-payoff",
	Instruction(func(ctx *Context) error {
		var cards []uuid.UUID
		for _, id := range ctx.StepRecord(shadowKinMilled).IDs {
			z := ctx.Game.FindCardZoneForEffect(id)
			if z == nil || z.Kind != game.ZoneGraveyard {
				continue
			}
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
				cards = append(cards, id)
			}
		}
		return shadowKinOffer(ctx, cards)
	}),
)

// shadowKinOffer asks the controller about the creature cards of ONE
// graveyard's owner (a pick names the cards of a single zone), and
// carries the cards of the other owners on the prompt. Declining asks
// about the next owner's cards; choosing one ends it, so the controller
// exiles at most one card among all of them.
func shadowKinOffer(ctx *Context, cards []uuid.UUID) error {
	if len(cards) == 0 {
		return nil
	}
	owner := uuid.Nil
	var here, rest []uuid.UUID
	for _, id := range cards {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if owner == uuid.Nil {
			owner = c.Owner
		}
		if c.Owner == owner {
			here = append(here, id)
		} else {
			rest = append(rest, id)
		}
	}
	if len(here) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: owner,
		Source:     ctx.Source(),
		Question:   "Shadow Kin — you may exile a creature card from among the cards milled this way",
		Cards:      here,
		Min:        0,
		Max:        1,
		ThenKey:    shadowKinPick,
		Carry:      rest,
		LiveZone:   game.ZoneGraveyard,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// shadowKinPick exiles the chosen card and, if it reached exile, makes
// Shadow Kin a copy of it. An on-disk key.
var shadowKinPick game.ChooseCardsThen

func init() {
	shadowKinPick = askedPick("choose-cards/shadow-kin-exile", shadowKinPicked)
}

func shadowKinPicked(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		// Nothing chosen from that graveyard: the cards of the next
		// owner's are still on offer.
		return shadowKinOffer(ctx, r.Carry)
	}
	g := ctx.Game
	id := r.Picked[0]
	values, ok := g.CopiableValuesForEffect(id)
	if err := g.ExileCardForEffect(id); err != nil {
		return err
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile || !ok {
		return nil
	}
	return BecomeCopy{
		Targets:  []uuid.UUID{ctx.Source()},
		Of:       id,
		Values:   &values,
		Duration: CopyIndefinite,
		Except:   func(v *game.PrintedValues) { v.GrantAbility(shadowKinGrant) },
		Label:    "Shadow Kin — becomes a copy",
	}.Apply(ctx)
}
