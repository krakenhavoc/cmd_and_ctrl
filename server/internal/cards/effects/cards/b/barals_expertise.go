package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baral's Expertise — Sorcery {3}{U}{U} (Tarkir: Dragonstorm Commander):
//
//	"Return up to three target artifacts and/or creatures to their
//	 owners' hands.
//	 You may cast a spell with mana value 4 or less from your hand
//	 without paying its mana cost."
//
// Calamitous Tide's bounce with the free cast as its keyed tail: the cast
// is offered after the cards have reached hands, so a returned artifact
// or creature is a legal pick (the printed ruling), and a bounce that
// pauses (a commander's CR 903.9 question) holds the cast back until it
// is answered. The free cast is Rishkar's Expertise's, with the ceiling
// at four: judged on the face being cast (#4509), pass-closed, ignoring
// timing (CR 608.2g). A spell whose targets have all gone does not
// resolve, so there is no cast (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1cdcb8c-1e12-4d39-aa4a-db678a80df9c",
		Name:         "Baral's Expertise",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("up to three target artifacts and/or creatures", effects.Or(effects.Artifact(), effects.Creature())).WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.BounceCardsToHandKeyedForEffect(effects.LegalTargetIDs(ctx), ctx.ItemTail(baralsExpertiseFreeCastTail))
		},
	})
}

// baralsExpertiseFreeCastTail is "You may cast a spell with mana value 4
// or less from your hand without paying its mana cost." after the return,
// keyed (#4253): the key is an on-disk identity.
var baralsExpertiseFreeCastTail = effects.CardTail("then/barals-expertise-free-cast", func(ctx *effects.Context, _ game.TailOutcome) error {
	effects.OfferFreeHandCast(ctx, ctx.Controller(), baralsExpertiseFreeCast,
		"a spell with mana value 4 or less", effects.SpellWithManaValueAtMost(4))
	return nil
})

// baralsExpertiseFreeCast is the offer's keyed continuation (#4246).
var baralsExpertiseFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/barals-expertise-cast", "Baral's Expertise")
