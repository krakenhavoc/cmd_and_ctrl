package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lieutenant Nyota Uhura — {2}{U} Legendary Creature — Human Officer 2/3
// (Star Trek Commander, #4974):
//
//	"Strange new worlds — {4}, {T}: Target opponent exiles the top X
//	 cards of their library, where X is the number of differently named
//	 lands you control. You may cast an instant or sorcery spell from
//	 among them without paying its mana cost. Then that player puts the
//	 rest on the bottom of their library in a random order."
//
// "Strange new worlds" is a flavor word with no rules meaning
// (CR 207.2d). X is counted once, as the ability resolves (CR 608.2h).
// The exile is one face-up run off the top of the opponent's library (a
// short library exiles what it has), and the cast is Herald of Amity's:
// ONE free, cast-only permission over the exiled cards for Uhura's
// controller, spent by the first cast, its filter "an instant or sorcery
// spell" judged on the face being cast, offered during the resolution
// (CR 608.2g). Once the offer is answered, every exiled card still in
// exile goes to the bottom of its owner's library in a random order,
// the opponent putting them there. Its {T} is a cost, so it cannot be
// activated the turn Uhura comes under your control (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d55f0074-2772-48b1-8d60-ca2da52d4c9d",
		Name:         "Lieutenant Nyota Uhura",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Strange new worlds — {4}, {T}: Target opponent exiles the top X cards of their library, where X is the number of differently named lands you control. You may cast an instant or sorcery spell from among them without paying its mana cost. Then that player puts the rest on the bottom of their library in a random order.",
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect:  uhuraStrangeNewWorlds,
		}},
	})
}

// uhuraStrangeNewWorlds exiles the top X cards of the still-legal target
// opponent's library.
func uhuraStrangeNewWorlds(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return MillToZone{
			Player:   t.ID,
			N:        trekB8DifferentlyNamedLands(g, item.Controller),
			To:       game.ZoneExile,
			ThenTail: ctx.ItemTail(uhuraExiled),
		}.Apply(ctx)
	}
	return nil
}

// uhuraExiled offers the free instant or sorcery cast over the exiled
// cards and bottoms the rest once it is answered. An on-disk key: never
// renamed.
var uhuraExiled = CardTail("lieutenant-nyota-uhura/after-exile", func(ctx *Context, o game.TailOutcome) error {
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
		Filter:      game.PermissionFilter{InstantOrSorceryOnly: true},
		Source:      source,
		SourceName:  "Lieutenant Nyota Uhura",
		Label:       "Lieutenant Nyota Uhura — cast an instant or sorcery spell from among them without paying its mana cost",
	}
	candidates := freeCastCandidates(g, &perm, exiled)
	if len(candidates) > 0 && g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "Lieutenant Nyota Uhura",
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put them on the bottom",
		Then:         uhuraRestToBottom,
		Carry:        exiled,
	}) {
		return nil
	}
	return uhuraBottomTheRest(g, exiled)
})

// uhuraRestToBottom is "Then that player puts the rest on the bottom of
// their library in a random order", once the offer is answered: Carry is
// the exiled cards. An on-disk key, never renamed.
var uhuraRestToBottom = game.RegisterChooseCardsThen("resolution-cast/lieutenant-nyota-uhura-rest-to-bottom",
	func(g *game.Game, r game.CardsChosen) error {
		return uhuraBottomTheRest(g, r.Carry)
	})

// uhuraBottomTheRest is Herald of Amity's bottoming with the cards'
// owner, the target opponent, as the player who puts them there.
func uhuraBottomTheRest(g *game.Game, ids []uuid.UUID) error {
	for _, id := range ids {
		if c, ok := g.LookupCardForEffect(id); ok {
			return heraldOfAmityBottomTheRest(g, c.Owner, ids)
		}
	}
	return nil
}
