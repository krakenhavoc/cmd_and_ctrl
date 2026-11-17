package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invasion of Alara // Awaken the Maelstrom — Battle — Siege, defense 7, for
// {W}{U}{B}{R}{G} (March of the Machine, S82 batch 23):
//
//	"When this Siege enters, exile cards from the top of your library until
//	 you exile two nonland cards with mana value 4 or less. You may cast one
//	 of those two cards without paying its mana cost. Put one of them into
//	 your hand. Then put the other cards exiled this way on the bottom of
//	 your library in a random order."
//
// Cascade's walk run for two hits (game.ExileUntilHitsForEffect): every
// card passed over, lands and expensive spells included, lands in exile face
// up. A library that runs dry stops the walk with fewer than two hits, and
// the clause works on what it found (CR 608.2c).
//
// The free cast is one offer over both hits, a {0} flash-timed permission
// spent whole by the first cast (CR 608.2g, CR 118.9): "one of those two
// cards". A card that can't be cast (a split card with only a land half is
// not an issue here) just isn't offered. Once it is answered:
//
//   - one was cast: the other hit is put into the hand;
//   - none was cast and both hits are still in exile: the controller
//     chooses which of the two goes into their hand;
//   - none was cast and one hit was found: it goes into the hand.
//
// Everything else exiled this way, the passed-over cards and the hit not
// put into the hand, goes to the bottom of the library in a random order
// (the controller's own random-order stream).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfAlaraOracleID,
		Name:         "Invasion of Alara",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 7,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Alara — exile cards from the top of your library until you exile two nonland cards with mana value 4 or less", invasionOfAlaraEnter),
			DefeatedTrigger("Invasion of Alara — defeated: exile it, then cast Awaken the Maelstrom", SiegeDefeated()),
		},
	})
}

// invasionOfAlaraOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfAlaraOracleID = "8faa4330-07bb-4b65-810a-bbce29d3cd42"

// invasionOfAlaraHit is "a nonland card with mana value 4 or less". An
// unreadable cost is not a hit (cascade's reading).
func invasionOfAlaraHit(c game.Card) bool {
	if c.IsLand() {
		return false
	}
	mv, ok := c.ParsedManaValue()
	return ok && mv <= 4
}

// invasionOfAlaraEnter is the enters trigger: the walk, then the free-cast
// offer, then the hand and the bottom.
func invasionOfAlaraEnter(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	pile, hits, err := g.ExileUntilHitsForEffect(controller, 2, invasionOfAlaraHit)
	if err != nil {
		return err
	}
	all := append(append([]uuid.UUID(nil), hits...), pile...)
	if len(hits) == 0 {
		return g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneExile, all)
	}
	perm := game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      source,
		SourceName:  "Invasion of Alara",
		Label:       "Invasion of Alara — cast it without paying its mana cost",
	}
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "Invasion of Alara",
		Cards:        hits,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Don't cast either",
		Then:         invasionOfAlaraAfterCast,
		Carry:        all,
	}) {
		return nil
	}
	return invasionOfAlaraFinish(g, controller, source, all, false)
}

// invasionOfAlaraAfterCast runs once the free-cast offer is answered: Picked
// is the card cast (none when declined), Carry everything exiled. An on-disk
// key.
var invasionOfAlaraAfterCast = game.RegisterChooseCardsThen("resolution-cast/invasion-of-alara-after-cast",
	func(g *game.Game, r game.CardsChosen) error {
		return invasionOfAlaraFinish(g, r.Chooser, r.Source, r.Carry, len(r.Picked) > 0)
	})

// invasionOfAlaraStillExiled is the cards of ids still in exile.
func invasionOfAlaraStillExiled(g *game.Game, ids []uuid.UUID, hitsOnly bool) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range ids {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if hitsOnly {
			if c, ok := g.LookupCardForEffect(id); !ok || !invasionOfAlaraHit(c) {
				continue
			}
		}
		out = append(out, id)
	}
	return out
}

// invasionOfAlaraFinish is "put one of them into your hand. Then put the
// other cards exiled this way on the bottom": `all` is every card the walk
// exiled, `cast` whether one of the two was cast.
func invasionOfAlaraFinish(g *game.Game, controller, source uuid.UUID, all []uuid.UUID, cast bool) error {
	left := invasionOfAlaraStillExiled(g, all, true)
	if !cast && len(left) > 1 {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  controller,
			Source:   source,
			Question: "Invasion of Alara — put one of them into your hand",
			Cards:    left,
			Min:      1,
			Max:      1,
			ThenKey:  invasionOfAlaraPutOne,
			Carry:    all,
			LiveZone: game.ZoneExile,
		})
		return nil
	}
	for _, id := range left {
		if _, err := g.ReturnFromAnyZoneToHandForEffect(controller, source, id); err != nil {
			return err
		}
	}
	return g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneExile, invasionOfAlaraStillExiled(g, all, false))
}

// invasionOfAlaraPutOne is the choice of which of the two uncast cards goes
// into the hand; the rest go to the bottom. An on-disk key.
var invasionOfAlaraPutOne = game.RegisterChooseCardsThen("choose-cards/invasion-of-alara-put-one",
	func(g *game.Game, r game.CardsChosen) error {
		picked := invasionOfAlaraStillExiled(g, r.Picked, false)
		if len(picked) > 1 {
			picked = picked[:1]
		}
		for _, id := range picked {
			if _, err := g.ReturnFromAnyZoneToHandForEffect(r.Chooser, r.Source, id); err != nil {
				return err
			}
		}
		return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneExile, invasionOfAlaraStillExiled(g, r.Carry, false))
	})
