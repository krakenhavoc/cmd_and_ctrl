package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jadzi, Oracle of Arcavios // Journey to the Oracle — a modal double-faced
// card (Strixhaven: School of Mages, #5502), one file for both faces.
//
// Front, Legendary Creature — Human Wizard {6}{U}{U}, 5/5:
//
//	"Discard a card: Return Jadzi to its owner's hand.
//	 Magecraft — Whenever you cast or copy an instant or sorcery spell, reveal
//	 the top card of your library. If it's a nonland card, you may cast it by
//	 paying {1} rather than paying its mana cost. If it's a land card, put it
//	 onto the battlefield."
//
// Back, Sorcery {2}{G}{G} (key "<oracle>#1", the MDFC face keyspace):
//
//	"You may put any number of land cards from your hand onto the
//	 battlefield. Then if you control eight or more lands, you may discard a
//	 card. If you do, return Journey to the Oracle to its owner's hand."
//
// The bounce is Bramble Familiar's, guarded so a Jadzi that left and came
// back is not returned (CR 400.7). The magecraft reveals the top card
// (public); a land goes onto the battlefield (Coiling Oracle's put, not a
// land play), and a nonland card is offered as a cast during the
// resolution (CR 608.2g) through a cast permission whose price is {1}
// instead of its mana cost: X is 0 and nothing else replaces the cost. A card
// not cast stays on top of the library, as printed.
//
// Journey puts the lands as one batch (Wrenn and Seven's pick), and once they
// have landed it counts your lands, so the ones just put count. A discard
// you do not make leaves the sorcery where it resolves, the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     jadziOracleID,
		Name:         "Jadzi, Oracle of Arcavios",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: Return Jadzi to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.DiscardACard(),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
		Triggered: []game.TriggeredAbility{
			effects.Magecraft(jadziMagecraftLabel, jadziMagecraft),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     jadziOracleID + "#1",
		Name:         "Journey to the Oracle",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PutFromHandOntoBattlefield{
				Match:    effects.Land(),
				Optional: true,
				Max:      1000,
				Label:    "Journey to the Oracle — you may put any number of land cards from your hand onto the battlefield",
				ThenTail: jadziJourneyLands,
			}.Apply(ctx)
		},
	})
}

const jadziOracleID = "ca79cd3f-13ba-4d11-b885-48a504ae69aa"

const jadziMagecraftLabel = "Jadzi, Oracle of Arcavios — reveal the top card; cast it for {1} if it's nonland, put it onto the battlefield if it's a land (magecraft)"

// jadziMagecraft is the magecraft body: reveal, then the land or the offer.
func jadziMagecraft(g *game.Game, item *game.StackItem) error {
	you, source := item.Controller, item.SourceCardID
	revealed := g.RevealTopOfLibraryForEffect(you, source, 1, jadziMagecraftLabel)
	if len(revealed) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(revealed[0])
	if !ok {
		return nil
	}
	if c.IsLand() {
		_, err := g.PutFromLibraryOntoBattlefieldForEffect(revealed[0], game.LibraryEntryOptions{Controller: you})
		return err
	}
	perm := effects.AnySpell.On(g.FreeCastFromLibraryPermission(you, source, "Jadzi, Oracle of Arcavios"))
	perm.Cost = "{1}"
	perm.LapseOnPass = game.LapseStaysInHand
	perm.Label = "Jadzi, Oracle of Arcavios — cast it by paying {1} rather than paying its mana cost"
	if len(effects.FreeCastCandidates(g, &perm, revealed)) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       source,
		Reason:       "Jadzi, Oracle of Arcavios",
		Cards:        revealed,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordPaidCast,
		AcceptLabel:  "Cast it, paying {1}",
		DeclineLabel: "Leave it on top of your library",
	})
	return nil
}

// jadziJourneyLands is what follows the land put: "if you control eight or
// more lands, you may discard a card. If you do, return Journey to the Oracle
// to its owner's hand." Keyed (#4253).
var jadziJourneyLands = effects.PutFromHandTail("put-from-hand/jadzi-journey-lands", func(ctx *effects.Context, res effects.PutFromHandResult) error {
	g := ctx.Game
	lands := effects.CountControlled(g, res.Player, func(c game.Card) bool { return c.IsLand() })
	if lands < 8 {
		return nil
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   res.Player,
		Source:   res.Source,
		N:        1,
		UpTo:     true,
		Question: "Journey to the Oracle — you may discard a card; if you do, return Journey to the Oracle to its owner's hand",
	}, jadziJourneyReturn, nil)
})

// jadziJourneyReturn is "if you do, return Journey to the Oracle to its
// owner's hand": the spell is the discard's source, still on the stack as it
// resolves. An on-disk key, never renamed.
var jadziJourneyReturn = game.RegisterDiscardThen("choose-cards/jadzi-journey-return",
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Discarded.Count() == 0 {
			return nil
		}
		return g.BounceToHandForEffect(r.Source)
	})
