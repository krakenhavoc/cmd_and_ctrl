package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b07_helpers.go — shared bodies for Innistrad: Crimson Vow,
// batch 7 (#5495): the creature-type words of the Wolf and Vampire cards.

// vowB07WolvesAndWerewolves counts the permanents `player` controls that
// are a Wolf or a Werewolf, after the layers (a changeling counts). The two
// words are words of the card's text, passed in already read (CR 612.2), so a changed word counts
// the changed type. Leaving `except` out of the count is "another": pass
// the source's own id, or uuid.Nil to count every one.
func VowB07WolvesAndWerewolves(g *game.Game, player uuid.UUID, wolf, werewolf string, except uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || c.InstanceID == except {
			continue
		}
		if c.HasSubtypeWord(wolf) || c.HasSubtypeWord(werewolf) {
			n++
		}
	}
	return n
}

// vowB07AnotherWolfOrWerewolf is "you control another Wolf or Werewolf",
// read for the source's controller.
func VowB07AnotherWolfOrWerewolf(g *game.Game, source *game.Card) bool {
	return VowB07WolvesAndWerewolves(g, source.Controller, source.TextSubtype("Wolf"), source.TextSubtype("Werewolf"), source.InstanceID) > 0
}

// vowB07SelfWhileAnotherWolf is the AppliesTo of "as long as you control
// another Wolf or Werewolf, this creature gets/has ...": the source alone.
func VowB07SelfWhileAnotherWolf(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID && VowB07AnotherWolfOrWerewolf(g, source)
}

// vowB07OneIfYouControlManaValueFourOrMore is the count of Ascendant
// Packleader's "enters with a +1/+1 counter on it if you control a
// permanent with mana value 4 or greater", for b19EntersWithCountersCounted:
// one when its controller controls such a permanent as it enters (it is
// still the stack object, so it is never its own permanent; a token has
// mana value 0).
func VowB07OneIfYouControlManaValueFourOrMore(g *game.Game, src *game.Card) int {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == src.Controller && c.ManaValue() >= 4 {
			return 1
		}
	}
	return 0
}

// vowB07TwoPerOtherNontokenHuman is the count of Hamlet Vanguard's "enters
// with two +1/+1 counters on it for each other nontoken Human you control",
// for b19EntersWithCountersCounted: two per Human its controller controls
// as it enters, through the layers (a changeling counts), tokens left out.
// The Vanguard is still the stack object, so it is never counted itself.
func VowB07TwoPerOtherNontokenHuman(g *game.Game, src *game.Card) int {
	human := src.TextSubtype("Human")
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == src.Controller && c.InstanceID != src.InstanceID && !c.IsToken() && c.HasSubtypeWord(human) {
			n++
		}
	}
	return 2 * n
}

// vowB07OpponentExilesTwoFromHand is the mode body of "target opponent
// exiles two cards from their hand" (Aim for the Head): the opponent picks
// their own two cards (all of them with fewer than two, none from an empty
// hand) and each is exiled through Ruthless Negotiation's keyed answer.
func VowB07OpponentExilesTwoFromHand(_ *game.StackItem, ctx *Context, occ int) error {
	for _, t := range ctx.ModeTargets(occ) {
		if t.Kind != game.TargetPlayer || !ctx.IsTargetLegal(t) {
			continue
		}
		p := ctx.Game.PlayerByIDForEffect(t.ID)
		if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			return nil
		}
		cards := make([]uuid.UUID, 0, p.Hand.Size())
		for _, c := range p.Hand.Cards {
			cards = append(cards, c.InstanceID)
		}
		n := min(2, len(cards))
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    t.ID,
			FromPlayer: t.ID,
			Source:     ctx.Item.SourceCardID,
			Question:   "Aim for the Head — exile two cards from your hand",
			Cards:      cards,
			Min:        n,
			Max:        n,
			ThenKey:    RuthlessNegotiationExiled,
			LiveZone:   game.ZoneHand,
		})
		return nil
	}
	return nil
}

// vowB07VampireBitesAnother is the mode body of "target Vampire you control
// deals damage equal to its power to another target creature" (Markov
// Retribution): Bite Down's one-sided bite over this occurrence's two
// clauses. Nothing is dealt unless both picks are still legal (CR 608.2b).
func VowB07VampireBitesAnother(_ *game.StackItem, ctx *Context, occ int) error {
	var biter, victim game.TargetRef
	var haveBiter, haveVictim bool
	for _, t := range ctx.ModeTargets(occ) {
		if !ctx.IsTargetLegal(t) || t.Kind != game.TargetCard {
			continue
		}
		switch t.Slot {
		case 0:
			biter, haveBiter = t, true
		case 1:
			victim, haveVictim = t, true
		}
	}
	if !haveBiter || !haveVictim {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(biter.ID)
	if !ok {
		return nil
	}
	return DealDamage{Source: biter.ID, Target: victim.ID, Amount: c.CurrentPower()}.Apply(ctx)
}
