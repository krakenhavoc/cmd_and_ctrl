package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b05_helpers.go — shared pieces of S82 GAP2 batch 5 (#5948):
// spell-on-stack characteristics and grants, and spell and ability copies
// and retargeting.

// ItemHasATarget is "with one or more targets" over a stack item (Wyll's
// Reversal): at least one target slot, counted as ItemHasASingleTarget
// counts them (StackItemTargetCountForEffect).
func ItemHasATarget() StackItemPredicate {
	return func(g *game.Game, _ uuid.UUID, item *game.StackItem) bool {
		return g.StackItemTargetCountForEffect(item.ID) >= 1
	}
}

// Gap2B05MayChooseNewTargetsForTheTarget is "you may choose new targets for
// that spell or ability" over the resolving item's own stack target (CR
// 115.7): every slot may move and declining is allowed. A target that has
// left the stack does nothing.
func Gap2B05MayChooseNewTargetsForTheTarget(ctx *Context) error {
	return RetargetTheTargetedItem(ctx, "You may choose new targets for that spell or ability", game.RetargetChooseNew, true)
}

// Gap2B05CopyTheTargetedSpellOrAbility is "copy it. You may choose new
// targets for the copy." over the resolving item's own stack target, read
// as it now stands: a spell is copied as a spell, an activated or
// triggered ability as an ability (CR 707.10, 707.10c). The copy is the
// resolving item's controller's.
func Gap2B05CopyTheTargetedSpellOrAbility(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		item := ctx.Game.StackItemForEffect(t.ID)
		if item == nil {
			return nil
		}
		if item.Kind == game.StackItemSpell {
			return CopySpell{StackID: t.ID, Controller: ctx.Controller(), ChooseNewTargets: true}.Apply(ctx)
		}
		return CopyAbility{ItemID: t.ID, Controller: ctx.Controller(), ChooseNewTargets: true}.Apply(ctx)
	}
	return nil
}

// ItemTargetsAPermanentYouControl is "spell or ability that targets a
// permanent you control" (Not of This World): one of the item's target
// slots is a permanent on the battlefield the chooser controls. A target
// that has left the battlefield is not a permanent any more.
func ItemTargetsAPermanentYouControl(match CardPredicate) StackItemPredicate {
	return func(g *game.Game, you uuid.UUID, item *game.StackItem) bool {
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			z := g.FindCardZoneForEffect(t.ID)
			if z == nil || z.Kind != game.ZoneBattlefield {
				continue
			}
			c, ok := g.LookupCardForEffect(t.ID)
			if ok && c.Controller == you && (match == nil || match(g, you, c)) {
				return true
			}
		}
		return false
	}
}

// ItTargetsAnItemThat is the cost twin of ItTargets for a spell whose
// target is a stack item (Not of This World's "if it targets a spell or
// ability that targets a creature you control with power 7 or greater"):
// one of the spell's announced targets is a spell or ability on the stack
// that `pred` admits, read as the cost is determined (CR 601.2f, after the
// targets of 601.2c). Use it on a modifier that sets ReadsTargets.
func ItTargetsAnItemThat(pred StackItemPredicate) CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		for _, t := range q.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			if item := q.Game.StackItemForEffect(t.ID); item != nil && pred(q.Game, q.Controller, item) {
				return true
			}
		}
		return false
	}
}

// gap2b05NeeraOffer is Neera, Wild Mage's keyed free-cast pick, the offer
// outside an open resolution (RegisterFreeLibraryCastOffer). An on-disk key.
var gap2b05NeeraOffer = RegisterFreeLibraryCastOffer("choose-cards/neera-wild-mage-cast", "Neera, Wild Mage")

// Gap2B05NeeraRevealAndMayCastFree is Neera, Wild Mage's "reveal cards from
// the top of your library until you reveal a nonland card. You may cast that
// card without paying its mana cost. Then put all revealed cards not cast
// this way on the bottom of your library in a random order." The cast is
// made during the resolution (CR 608.2g) from where the card sits in the
// library, through the resolution-cast prompt (Sunbird's Invocation's
// shape); a library with no nonland card reveals it all and bottoms it.
func Gap2B05NeeraRevealAndMayCastFree(ctx *Context) error {
	g, controller, source := ctx.Game, ctx.Controller(), ctx.Source()
	name := gap2b05NeeraOffer.name
	run, hit := RevealUntil(ctx, controller, func(c game.Card) bool { return !c.IsLand() },
		name+" — reveal cards until a nonland card")
	if len(run) == 0 {
		return nil
	}
	perm := g.FreeCastFromLibraryPermission(controller, source, name)
	var eligible []uuid.UUID
	if hit != uuid.Nil {
		eligible = FreeCastCandidates(g, &perm, []uuid.UUID{hit})
	}
	if len(eligible) == 0 {
		return finishFreeCastFromLooked(g, controller, source, name, run, uuid.Nil)
	}
	perm.LapseOnPass = game.LapseStaysInHand
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       name,
		Cards:        eligible,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put them on the bottom",
		Then:         LookedAtRestToBottom,
		Carry:        run,
	}) {
		return nil
	}
	eligible = keyedFreeCastCandidates(g, &perm, eligible)
	if len(eligible) == 0 {
		return finishFreeCastFromLooked(g, controller, source, name, run, uuid.Nil)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:     controller,
		Source:      source,
		Question:    name + " — cast the revealed nonland card without paying its mana cost?",
		Cards:       eligible,
		Min:         0,
		Max:         1,
		Destination: game.ChooseCastFreeFromLibrary,
		ThenKey:     gap2b05NeeraOffer.then,
		Carry:       run,
	})
	return nil
}

// Gap2B05TargetBecomesColor is "target spell or permanent becomes <colour>
// until end of turn" over the resolving item's first legal target (CR
// 613.1e): a spell on the stack through the stack step's layer 5
// (game.SetSpellColorsForEffect, #6088), a permanent through an ordinary
// setColors record. A target that has left does nothing (CR 608.2b).
func Gap2B05TargetBecomesColor(ctx *Context, colors []string, label string) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if ctx.Game.StackItemForEffect(t.ID) != nil {
			ctx.Game.SetSpellColorsForEffect(ctx.Source(), t.ID, colors, DurationUntilEndOfTurn(ctx), label)
			return nil
		}
		return ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.SetColorsMod(colors...)},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    label,
		}.Apply(ctx)
	}
	return nil
}

// SpentSpellGainsUntilYourNextTurn is SpentSpellGains for "it gains
// <keywords> until your next turn" (Jade Orb of Dragonkind's hexproof): the
// grant is made on the spell at the spend, the stack keyword pass applies
// it, and the permanent the spell becomes keeps it until its caster's next
// turn begins (CR 400.7a, 611.2b).
func SpentSpellGainsUntilYourNextTurn(keywords []string, when ...string) game.ManaSpendRider {
	return game.ManaSpendRider{
		Kind:              game.ManaRiderSpellGains,
		Keywords:          append([]string(nil), keywords...),
		UntilYourNextTurn: true,
		When:              when,
	}
}

// CounterTheFirstLegalStackTarget counters the resolving item's first legal
// target that is a spell or ability on the stack (Not of This World, Ooze
// Spill). A target countered or resolved in response does nothing (CR
// 608.2b).
func CounterTheFirstLegalStackTarget(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return CounterTarget{StackID: t.ID}.Apply(ctx)
		}
	}
	return nil
}

// TheFirstLegalTargetPlayerDrawsACard is "target player draws a card" over
// the resolving item's first legal player target (Manhole Cover, the
// ability Thief of Existence gains).
func TheFirstLegalTargetPlayerDrawsACard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return DrawCards{Player: t.ID, N: 1}.Apply(ctx)
		}
	}
	return nil
}

// SacrificeThisPermanentAtYourEndStep is the body of a granted "At the
// beginning of your end step, sacrifice this permanent." (Terminal
// Velocity, Mizzix, Replica Rider): the permanent the ability is on, while
// it is still on the battlefield under the trigger's controller.
func SacrificeThisPermanentAtYourEndStep(g *game.Game, item *game.StackItem) error {
	if !OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || c.Controller != item.Controller {
		return nil
	}
	return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
}
