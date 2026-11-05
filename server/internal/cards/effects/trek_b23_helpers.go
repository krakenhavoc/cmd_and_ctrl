package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b23_helpers.go — shared bodies for Star Trek batch 23 (#4973).

// --- tokens -----------------------------------------------------------------

// printedTrekB23CantBlockWormToken is Cleveland Booker, Empath's token:
// a 2/2 red and green Worm creature token with "This token can't
// block." (ttrk 18). A catalog template and not a tokens_table row
// because the can't-block clause is a static the token always prints,
// and the plain "1/1 black and green Worm" row is a different token.
func printedTrekB23CantBlockWormToken() tokenTemplate {
	return tokenTemplate{
		Slug: "worm-cant-block",
		Card: game.Card{
			Name:      "Worm",
			TypeLine:  "Token Creature — Worm",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"R", "G"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Text:   "This token can't block.",
	}
}

// trekB23CantBlockWormToken is the Worm's card value for CreateToken.
func trekB23CantBlockWormToken() game.Card {
	return tokenFromCatalog(printedTrekB23CantBlockWormToken)
}

// printedTrekB23CantPhaseInHumanToken is The City on the Edge of
// Forever's token: a 0/1 white Human creature token with "Permanents
// can't phase in." (ttrk 1). The line is the PermanentsCantPhaseIn world
// rule carried on the token itself; game.phaseInLocked refuses every
// phase-in while the token is on the battlefield.
func printedTrekB23CantPhaseInHumanToken() tokenTemplate {
	return tokenTemplate{
		Slug: "human-permanents-cant-phase-in",
		Card: game.Card{
			Name:           "Human",
			TypeLine:       "Token Creature — Human",
			Power:          0,
			Toughness:      1,
			Colors:         []string{"W"},
			PrintedPTKnown: true,
		},
		Static: []game.StaticAbility{RestrictSelf(game.PermanentsCantPhaseIn)},
		Text:   "Permanents can't phase in.",
	}
}

// trekB23CantPhaseInHumanToken is the Human's card value for CreateToken.
func trekB23CantPhaseInHumanToken() game.Card {
	return tokenFromCatalog(printedTrekB23CantPhaseInHumanToken)
}

// --- All Good Things... -------------------------------------------------------

// trekB23DrawTwoThenExileFromHand is All Good Things...'s chapters I, II
// and III: "Draw two cards, then exile a card from your hand." The draw
// comes first, then the controller picks one card of the hand they hold
// now. The exile is linked to the Saga object (CR 607.2a, Card.ExiledWith)
// so chapter IV can read "cards exiled with this Saga". An empty hand
// exiles nothing.
func trekB23DrawTwoThenExileFromHand(g *game.Game, item *game.StackItem) error {
	if err := (DrawCards{N: 2}).Apply(NewContext(g, item)); err != nil {
		return err
	}
	var hand []uuid.UUID
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
		for _, c := range p.Hand.Cards {
			hand = append(hand, c.InstanceID)
		}
	}
	if len(hand) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "All Good Things... — exile a card from your hand",
		Cards:    hand,
		Min:      1,
		Max:      1,
		ThenKey:  trekB23AllGoodThingsExiled,
		LiveZone: game.ZoneHand,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// trekB23AllGoodThingsExiled is the keyed answer: the pick is exiled,
// linked to the asking chapter's source object. An on-disk key: never
// renamed, never reused.
var trekB23AllGoodThingsExiled = game.RegisterChooseCardsThen("choose-cards/all-good-things-exile-from-hand",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) == 0 || r.Asked == nil {
			return nil
		}
		return ExileTarget{Target: r.Picked[0], Linked: true}.Apply(NewContext(g, r.Asked))
	})

// trekB23AllGoodThingsFinale is chapter IV: "If all cards exiled with
// this Saga share a card type, you gain 10 life." The cards are the ones
// in exile linked to this Saga object (a card that has left exile is a
// new object, CR 400.7). They share a card type when one card type is
// common to every one of them; with no card exiled there is nothing to
// share and no life is gained.
func trekB23AllGoodThingsFinale(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	exiled := g.ExiledWithForEffect(game.PermissionCardRef(ref))
	if !trekB23ShareACardType(exiled) {
		return nil
	}
	return GainLife{Player: item.Controller, Amount: 10}.Apply(ctx)
}

// trekB23ShareACardType reports whether some card type is common to
// every card in cards. False for no cards.
func trekB23ShareACardType(cards []*game.Card) bool {
	if len(cards) == 0 {
		return false
	}
	for _, t := range semblanceCardTypes {
		all := true
		for _, c := range cards {
			if !c.HasCardType(t) {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}
	return false
}

// --- The City on the Edge of Forever --------------------------------------------

// trekB23CityChapterOne is "All artifacts and creatures phase out. Create
// a 0/1 white Human creature token with 'Permanents can't phase in.'"
// Every artifact and creature on the battlefield phases out in one
// simultaneous event (CR 702.26a; attachments go with their hosts,
// CR 702.26g), then the token is created, so it stays in.
func trekB23CityChapterOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsArtifact() || c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) > 0 {
		if err := (PhaseOut{Targets: ids}).Apply(ctx); err != nil {
			return err
		}
	}
	return CreateToken{Template: trekB23CantPhaseInHumanToken(), N: 1}.Apply(ctx)
}

// eachPlayerLosesTwoLife is "each player loses 2 life": every live
// player, the controller included, in seat order. Life loss, not damage.
// Crushing Disappointment and The City on the Edge of Forever's chapter II.
func eachPlayerLosesTwoLife(ctx *Context) error {
	for _, id := range tablePlayers(ctx) {
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -2); err != nil {
			return err
		}
	}
	return nil
}

// trekB23CityChapterThree is "Sacrifice a creature token. If you do, you
// gain 5 life." The controller picks one creature token they control as
// the chapter resolves; with none there is no sacrifice and no life.
func trekB23CityChapterThree(g *game.Game, item *game.StackItem) error {
	return mustSacrificeOne(g, item, "The City on the Edge of Forever — sacrifice a creature token",
		And(Creature(), IsTokenPredicate()), trekB23CityGainFive)
}

// trekB23CityGainFive is chapter III's "if you do, you gain 5 life". An
// on-disk key.
var trekB23CityGainFive = RegisterSacrificedOne("the-city-on-the-edge-of-forever/gain-five",
	func(ctx *Context, _ uuid.UUID) error {
		return GainLife{Player: ctx.Controller(), Amount: 5}.Apply(ctx)
	})

// --- Constable Odo ----------------------------------------------------------------

// trekB23TapAndTwoStunTheModesTarget is "Tap target creature an opponent
// controls. Put two stun counters on it." as a modal bullet's body: the
// bullet's target, if it is still legal, is tapped and gets two stun
// counters (CR 122.1d).
func trekB23TapAndTwoStunTheModesTarget(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: t.ID, Kind: game.CounterStun, N: 2}.Apply(ctx)
}
