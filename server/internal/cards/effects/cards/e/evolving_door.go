package cardse

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Evolving Door — Artifact {2}{G} (Streets of New Capenna, #5878):
//
//	"{1}, {T}, Sacrifice a creature: Count the colors of the sacrificed
//	 creature, then search your library for a creature card that's
//	 exactly that many colors plus one. Exile that card, then shuffle.
//	 You may cast the exiled card. Activate only as a sorcery."
//
// No new seam: the issue's three pieces were all on develop.
//
// The count is the sacrificed creature as it last existed on the
// battlefield (CR 608.2h, ctx.SacrificedPermanent), so a colour an effect
// gave it counts and a token that has ceased to exist is still read. The
// search is Bring to Light's: a stated quality, so the searcher may find
// nothing (CR 701.23b), the card goes to exile, then the library is
// shuffled, and the rest rides a keyed tail.
//
// "You may cast the exiled card" is cast during the resolution (CR
// 608.2g), so timing is no bar, but it is PAID: the card's own mana cost,
// as Boiling Rock Rioter's offer (an empty permission Cost is the printed
// cost). Declined, the card stays in exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30e4849e-c58c-4b65-9fb3-7aa40c944904",
		Name:         "Evolving Door",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, {T}, Sacrifice a creature: Search your library for a creature card that's exactly one more color than the sacrificed creature, exile it, then shuffle. You may cast the exiled card.",
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeACreature()),
			SorcerySpeed: true,
			Effect:       evolvingDoorSearch,
		}},
	})
}

const evolvingDoorName = "Evolving Door"

func evolvingDoorSearch(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	sacrificed, ok := ctx.SacrificedPermanent()
	if !ok {
		return nil
	}
	want := len(sacrificed.Characteristic.Colors) + 1
	return effects.SearchLibrary{
		Player: ctx.Controller(),
		Limit:  1,
		Dest:   game.ZoneExile,
		Predicate: func(c game.Card) bool {
			return c.IsCreature() && len(c.EffectiveColors()) == want
		},
		Shuffle:  true,
		Reason:   evolvingDoorName + " — a creature card that's exactly " + strconv.Itoa(want) + " colors",
		ThenTail: ctx.ItemTail(evolvingDoorCast),
	}.Apply(ctx)
}

// evolvingDoorCast is "you may cast the exiled card", paid, over the card
// the search exiled. An on-disk key: never renamed, never reused.
var evolvingDoorCast = effects.CardTail("evolving-door/cast", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	var cards []game.Card
	for _, id := range o.IDs {
		z := g.FindCardZoneForEffect(id)
		c, ok := g.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		cards = append(cards, c)
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      ctx.Controller(),
		Zone:        game.ZoneExile,
		Timing:      game.TimingFlash,
		CastOnly:    true,
		Duration:    g.UntilEndOfTurnDuration(),
		LapseOnPass: game.LapseStaysInExile,
		Source:      ctx.Source(),
		SourceName:  evolvingDoorName,
		Label:       evolvingDoorName + " — cast the exiled card",
	}, cards)
	effects.OfferResolutionCastOver(g, ctx.Controller(), ctx.Source(), evolvingDoorName,
		game.MayCastKeywordPaidCast, effects.DeclineLeaveInExile, cards[0].InstanceID)
	return nil
})
