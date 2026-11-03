package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Herald of Amity — Creature — Griffin {3}{W}, 2/2 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Flying
//	 When this creature enters, exile the top eight cards of your library.
//	 You may cast an Aura spell from among them without paying its mana
//	 cost. Then put the rest on the bottom of your library in a random
//	 order.
//	 Whenever this creature attacks, it gets +X/+X until end of turn,
//	 where X is the number of Auras you control."
//
// The exile is one run off the top (a short library exiles what it has),
// face up. The cast is Scarlet Witch's: ONE free, cast-only permission
// over the exiled cards, spent by the first cast, its filter "an Aura
// spell" (PermissionFilter.Subtype) judged on the face being cast, and
// offered during the resolution (CR 608.2g), so the rulings' "you can't
// wait to cast it later in the turn" holds and an Aura with {X} is cast
// with X = 0. Once the offer is answered, by the cast or the decline,
// every exiled card still in exile goes to the bottom of the library in a
// random order from the game's keyed stream.
//
// The attack trigger counts the Auras you control as it resolves, as the
// rulings say, and pumps the creature until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6c25c6b5-d98c-4eb4-aacd-b6960a8779a9",
		Name:            "Herald of Amity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Herald of Amity — exile the top eight cards; you may cast an Aura spell from among them free",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{
						Player:   item.Controller,
						N:        8,
						To:       game.ZoneExile,
						ThenTail: ctx.ItemTail(heraldOfAmityExiled),
					}.Apply(ctx)
				}),
			WheneverThisAttacks("Herald of Amity — it gets +X/+X until end of turn, where X is the number of Auras you control",
				func(g *game.Game, item *game.StackItem) error {
					n := countControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype("Aura") })
					return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n,
						Label: "Herald of Amity — +X/+X"}.Apply(NewContext(g, item))
				}),
		},
	})
}

// heraldOfAmityExiled offers the free Aura cast over the exiled cards and
// bottoms the rest once it is answered. An on-disk key: never renamed.
var heraldOfAmityExiled = CardTail("herald-of-amity/after-exile", func(ctx *Context, o game.TailOutcome) error {
	g, controller, source := ctx.Game, ctx.Controller(), ctx.Source()
	var exiled []uuid.UUID
	for _, id := range o.IDs {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			exiled = append(exiled, id)
		}
	}
	if len(exiled) == 0 {
		return nil
	}
	perm := game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Filter:      game.PermissionFilter{Subtype: "Aura"},
		Source:      source,
		SourceName:  "Herald of Amity",
		Label:       "Herald of Amity — cast an Aura spell from among them without paying its mana cost",
	}
	candidates := freeCastCandidates(g, &perm, exiled)
	if len(candidates) > 0 && g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "Herald of Amity",
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put them on the bottom",
		Then:         heraldOfAmityRestToBottom,
		Carry:        exiled,
	}) {
		return nil
	}
	return heraldOfAmityBottomTheRest(g, controller, exiled)
})

// heraldOfAmityRestToBottom is "Then put the rest on the bottom of your
// library in a random order", once the offer is answered: Carry is the
// exiled cards. An on-disk key, never renamed.
var heraldOfAmityRestToBottom = game.RegisterChooseCardsThen("resolution-cast/herald-of-amity-rest-to-bottom",
	func(g *game.Game, r game.CardsChosen) error {
		return heraldOfAmityBottomTheRest(g, r.Chooser, r.Carry)
	})

// heraldOfAmityBottomTheRest puts every one of the cards still in exile on
// the bottom of its owner's library in a random order (the one cast is on
// the stack, not in exile).
func heraldOfAmityBottomTheRest(g *game.Game, controller uuid.UUID, ids []uuid.UUID) error {
	var rest []uuid.UUID
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			rest = append(rest, id)
		}
	}
	return g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneExile, rest)
}
