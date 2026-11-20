package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dis_b1_helpers.go — the shared bodies of Dissension batch 1 (#5598):
// the guild "sacrifice it unless {C} was spent" creatures, the Eidolon
// cycle's return from the graveyard, hellbent's statics and the Simic
// graft creatures' "target creature with a +1/+1 counter on it gains
// <keyword>" rows.

// disSacrificeUnlessSpent is "When this creature enters, sacrifice it
// unless {C} was spent to cast it" (Azorius Herald, Plaxmanta, Patagia
// Viper). The question is about the spell this permanent was (CR
// 400.7d, ManaSpentToCastThis), so a creature put onto the battlefield
// without being cast had nothing spent and is sacrificed. A source that
// has left the battlefield, or left and come back as a new object, is
// not this permanent any more and is not touched (CR 400.7). The card
// declares ReadsManaSpent, so a waived payment asks what was paid.
func disSacrificeUnlessSpent(name, color string) game.TriggeredAbility {
	return WhenThisEnters(name+" — sacrifice it unless {"+color+"} was spent to cast it", func(g *game.Game, item *game.StackItem) error {
		if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
			return nil
		}
		ctx := NewContext(g, item)
		if ctx.ManaSpentToCastThis().Count(color) > 0 {
			return nil
		}
		return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
	})
}

// disEidolonReturn is the Eidolon cycle's "Whenever you cast a
// multicolored spell, you may return this card from your graveyard to
// your hand." It works only from the graveyard (InGraveyard), and the
// return finds only the very card that triggered (CR 400.7).
func disEidolonReturn(name string) game.TriggeredAbility {
	return InGraveyard(Optional(WheneverYouCast(Multicolored(),
		name+" — return it from your graveyard to your hand",
		returnThisCardFromYourGraveyardToHand),
		"Return "+name+" from your graveyard to your hand?"))
}

// disHellbent is hellbent's "you have no cards in hand" (CR 207.2c, an
// ability word) about the source's controller. A hand's size is public.
func disHellbent(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	return p != nil && (p.Hand == nil || len(p.Hand.Cards) == 0)
}

// disSelfPumpWhileHellbent is "Hellbent — This creature gets +p/+t as
// long as you have no cards in hand" (Demon's Jester): a layer 7c
// modification of the source alone. The static declares that it depends
// on hand sizes, so a draw, a discard or a card put into the hand
// re-reads it (Vampire Champion Fury's hellbent).
func disSelfPumpWhileHellbent(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:             "Hellbent — This creature gets " + signedPT(power, toughness) + " as long as you have no cards in hand.",
		Layer:             game.Layer7PT,
		SubLayer:          game.SubLayer7C_Modify,
		AppliesTo:         whileYouHave(disHellbent, selfCreature),
		Apply:             pumpBy(power, toughness),
		DependsOnHandSize: true,
	}
}

// disSelfKeywordWhileHellbent is "Hellbent — This creature has <keyword>
// as long as you have no cards in hand" (Rakdos Pit Dragon's double
// strike): a layer 6 grant on the same live condition.
func disSelfKeywordWhileHellbent(keyword string) game.StaticAbility {
	ab := KeywordGrant(whileYouHave(disHellbent, selfCreature), keyword)
	ab.Label = "Hellbent — This creature has " + keyword + " as long as you have no cards in hand."
	ab.DependsOnHandSize = true
	return ab
}

// disAttachedPumpWhileHellbent is "Hellbent — Enchanted creature gets an
// additional +p/+t as long as you have no cards in hand" (Taste for
// Mayhem): a layer 7c modification of the host, read off the Aura's
// controller's hand.
func disAttachedPumpWhileHellbent(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:             "Hellbent — Enchanted creature gets an additional " + signedPT(power, toughness) + " as long as you have no cards in hand.",
		Layer:             game.Layer7PT,
		SubLayer:          game.SubLayer7C_Modify,
		AppliesTo:         whileYouHave(disHellbent, AttachedToSource),
		Apply:             pumpBy(power, toughness),
		DependsOnHandSize: true,
	}
}

// disWithPlusOneCounter is "target creature with a +1/+1 counter on it",
// judged as the target is chosen and again as the ability resolves (CR
// 608.2b).
func disWithPlusOneCounter() *game.TargetSpec {
	return TargetCreature("target creature with a +1/+1 counter on it", Not(b13WithoutPlusOneCounter()))
}

// disGraftGrantRow is the whole activated row "<cost>: Target creature
// with a +1/+1 counter on it gains <keyword> until end of turn."
// (Aquastrand Spider's reach, Cytospawn Shambler's trample, Plaxcaster
// Frogling's shroud; Helium Squirter's flying is the catalogued sibling).
func disGraftGrantRow(cost, keyword, name string) ActivatedAbility {
	return ActivatedAbility{
		Label:   cost + ": Target creature with a +1/+1 counter on it gains " + keyword + " until end of turn.",
		Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, keyword)),
		Cost:    ManaCost(cost),
		Targets: disWithPlusOneCounter(),
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{keyword}, Label: name + " — " + keyword}.Apply(ctx)
			}
			return nil
		},
	}
}

// disGraft is graft N's two halves (CR 702.58a): the entry replacement
// that puts N +1/+1 counters on the creature, and the trigger that may
// move one onto another creature as it enters (cmdr2B23Graft).
func disGraft(name string, n int) ([]game.ReplacementEffect, []game.TriggeredAbility) {
	counters := "+1/+1 counters"
	if n == 1 {
		counters = "+1/+1 counter"
	}
	return []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, n, "Graft "+strconv.Itoa(n)+" — enters with "+numberWord(n)+" "+counters+" on it."),
		},
		[]game.TriggeredAbility{cmdr2B23Graft(name)}
}

// disIckspitterEffect is Rakdos Ickspitter's "This creature deals 1
// damage to target creature and that creature's controller loses 1
// life." The controller is read before the damage, which cannot change
// it, and the life loss follows whether or not the damage was prevented.
func disIckspitterEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		controller := c.Controller
		if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
			return err
		}
		return g.ChangePlayerLifeForEffect(item.SourceCardID, controller, -1)
	}
	return nil
}
