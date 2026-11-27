package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b587_helpers.go — Saviors of Kamigawa, batch 587 (#5601): the
// graveyard recursion of the ki cycle, Death Denied, Footsteps of the Goryo and
// Iname as One.

// MoreCardsInHandThanEachOpponent is "you have more cards in hand than
// each opponent": strictly more than every opponent still in the game.
func MoreCardsInHandThanEachOpponent(g *game.Game, you uuid.UUID) bool {
	me := g.PlayerByIDForEffect(you)
	if me == nil || me.Hand == nil {
		return false
	}
	mine := len(me.Hand.Cards)
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you {
			continue
		}
		if p.Hand != nil && len(p.Hand.Cards) >= mine {
			return false
		}
	}
	return true
}

// OldB587ReturnThisCardToHandUpkeep is the ki cycle's upkeep ability, "At the
// beginning of your upkeep, if you have more cards in hand than each opponent,
// you may return this card from your graveyard to your hand", for a card whose
// other text is a spell. It works from the graveyard (CR 113.6) and its "if" is
// an intervening if (CR 603.4): asked as the upkeep begins and again as it
// resolves, before the "you may".
func OldB587ReturnThisCardToHandUpkeep() game.TriggeredAbility {
	const label = "return this card from your graveyard to your hand"
	return InGraveyard(Optional(InterveningIfYou("you have more cards in hand than each opponent", MoreCardsInHandThanEachOpponent,
		AtYourUpkeep(label, ReturnThisCardFromYourGraveyardToYourHand)), label+"?"))
}

// --- Akuta, Born of Ash ------------------------------------------------

// OldB587AkutaSacrificeASwamp is Akuta's upkeep effect: "you may sacrifice a
// Swamp. If you do, return Akuta from your graveyard to the battlefield." The
// pick is the "may" (zero or one Swamp, CR 701.21); the return happens only if a
// Swamp was sacrificed and Akuta is still in the graveyard.
func OldB587AkutaSacrificeASwamp(g *game.Game, item *game.StackItem) error {
	return ChoosePermanents{
		Question: "Akuta, Born of Ash — you may sacrifice a Swamp to return Akuta to the battlefield",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			return PermanentsControlledByMatching(g, of, HasSubtype("Swamp")), 0, 1
		},
		Keyed: oldB587AkutaPick,
	}.Apply(NewContext(g, item))
}

var oldB587AkutaPick = RegisterPermanentsPick("prompt-run/old-b587-akuta-sacrifice-a-swamp", PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		if picked.Count() == 0 {
			return nil
		}
		id := ctx.Source()
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
	},
})

// --- Iname as One ------------------------------------------------------

// OldB587InameSearch is Iname as One's enters trigger: if it was cast from
// hand, you may search your library for a Spirit permanent card, put it onto the
// battlefield, then shuffle.
func OldB587InameSearch(g *game.Game, item *game.StackItem) error {
	if !B30CastFromHand(g, item.SourceCardID) {
		return nil
	}
	spirit := g.TextReaderForEffect(item).TextSubtype("Spirit")
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: func(c game.Card) bool { return IsPermanentCard(c) && c.HasSubtypeWord(spirit) },
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Optional:  true,
		Source:    item.SourceCardID,
		Reason:    "Iname as One — search for a Spirit permanent card",
	}.Apply(NewContext(g, item))
}

// OldB587InameDies is Iname's dies trigger body: "you may exile it. If you do,
// return target Spirit permanent card from your graveyard to the battlefield."
// The "may" is the trigger's optional prompt; the return is the exile's tail, so
// it runs only if Iname actually reached exile (a replacement may stop it), and
// a target of Iname itself, which has just left the graveyard, is gone.
func OldB587InameDies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ExileTarget{Target: item.SourceCardID, ThenTail: ctx.ItemTail(oldB587InameExiled)}.Apply(ctx)
}

// oldB587InameExiled is the "if you do" of Iname's dies trigger. An on-disk
// key: never renamed, never reused.
var oldB587InameExiled = CardTail("old-b587/iname-exiled-return-a-spirit", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
})

// --- Death Denied -------------------------------------------------------

// OldB587ReturnTargetsToHand returns every still-legal card target to its
// owner's hand ("Return X target creature cards from your graveyard to your
// hand").
func OldB587ReturnTargetsToHand(_ *game.StackItem, ctx *Context) error {
	for _, id := range LegalTargetCardIDs(ctx) {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// --- Footsteps of the Goryo ---------------------------------------------

// OldB587FootstepsReturn returns the target creature card to the battlefield and
// schedules its sacrifice at the beginning of the next end step.
func OldB587FootstepsReturn(_ *game.StackItem, ctx *Context) error {
	return ReturnFromGraveyardTogether{
		Targets:    LegalTargetCardIDs(ctx),
		Controller: ctx.Controller(),
		ThenTail:   oldB587FootstepsSacrificeTail,
	}.Apply(ctx)
}

// oldB587FootstepsSacrificeTail is "Sacrifice that creature at the beginning of
// the next end step", a delayed trigger naming the object that entered (CR
// 603.7, CR 400.7). An on-disk key: never renamed, never reused.
var oldB587FootstepsSacrificeTail = TogetherTail("old-b587/sacrifice-that-creature-at-the-next-end-step",
	func(ctx *Context, entered []uuid.UUID, _ int) error {
		for _, id := range entered {
			c, ok := BattlefieldCardForEffect(ctx.Game, id)
			if !ok {
				continue
			}
			if err := (ScheduleDelayedTrigger{
				At:         game.StepEnd,
				Controller: ctx.Item.Controller,
				Label:      "Footsteps of the Goryo — sacrifice that creature",
				Body:       Cmdr2S12SoulstokeSacrifice,
				Params:     game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

// OldB587SpiritPermanentCardTarget is "Spirit permanent card" as a target
// predicate: a card with a permanent type, not a token, that is a Spirit.
func OldB587SpiritPermanentCardTarget() CardPredicate {
	return And(func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return IsPermanentCard(c) }, Subtype("Spirit"))
}

// OldB587TargetPlayerSacrificesASmallCreature is "target player sacrifices a
// creature of their choice with mana value 3 or less" (Exile into Darkness).
func OldB587TargetPlayerSacrificesASmallCreature(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		for _, t := range NewContext(g, item).LegalTargets() {
			if t.Kind == game.TargetPlayer {
				g.PlayerSacrificesForEffect(item.SourceCardID, t.ID, SacrificeSpec("a creature with mana value 3 or less", Creature(), ManaValueLE(3)), label)
				return nil
			}
		}
		return nil
	}
}
