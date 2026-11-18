package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invoke Calamity — Instant {1}{R}{R}{R}{R} (Kamigawa: Neon Dynasty):
//
//	"You may cast up to two instant and/or sorcery spells with total mana
//	 value 6 or less from your graveyard and/or hand without paying their
//	 mana costs. If those spells would be put into your graveyard, exile
//	 them instead. Exile Invoke Calamity."
//
// The controller picks up to two instant and sorcery cards from their hand
// and graveyard together as the spell resolves, and the set is held to a
// total mana value of 6 (neoTotalSpellManaValueAtMost: X counts as 0,
// CR 202.3e; a card with two instant or sorcery faces counts its larger).
// Each picked card gets its own free, any-timing cast permission
// (ExileInsteadOfGraveyard is the "exile them instead" rider, so it holds
// when the spell resolves, is countered or fizzles), and the casts are made
// during the resolution (CR 608.2g) through the resolution_cast prompt; a
// card not cast stays where it was. Invoke Calamity itself is exiled only
// after the casts are made, whether or not any were (neoCalamityExile), and
// is exiled rather than put into the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b3f041e-ad4a-47ea-bdc4-1be2353f2e18",
		Name:         "Invoke Calamity",
		Completeness: CompletenessFull,
		OnResolve:    invokeCalamityResolve,
	})
}

const neoCalamityName = "Invoke Calamity"

// neoCalamityBudget is "total mana value 6 or less".
const neoCalamityBudget = 6

// neoCalamityPermission is one card's free cast: the any-timing {0} cast of
// an instant or sorcery spell of mana value 6 or less, in the hand or the
// graveyard the card is in, exiled rather than put into the graveyard. A
// pass closes the window with the card left where it is.
func neoCalamityPermission(g *game.Game, controller, source uuid.UUID) game.CastPermission {
	perm := freeInstantOrSorcery().withManaValueAtMost(neoCalamityBudget).on(g.FreeCastFromHandPermission(controller, source, neoCalamityName))
	perm.Zone = ""
	perm.AltCostKey = game.AltCostKeyGrantedCost
	perm.ExileInsteadOfGraveyard = true
	perm.ExileInsteadOnlyInstantOrSorcery = false
	perm.LapseOnPass = game.LapseStaysInHand
	return perm
}

func invokeCalamityResolve(item *game.StackItem, ctx *Context) error {
	g, you := ctx.Game, ctx.Controller()
	perm := neoCalamityPermission(g, you, ctx.Source())
	var cands []uuid.UUID
	for _, id := range allHandCardIDs(g, you) {
		if c, ok := g.LookupCardForEffect(id); ok && perm.OffersASpell(c) {
			cands = append(cands, id)
		}
	}
	for _, id := range graveyardCardIDs(ctx, you, perm.OffersASpell) {
		cands = append(cands, id)
	}
	if len(cands) == 0 {
		if sourceIsNewObject(g, item) {
			return nil
		}
		return g.ExileCardForEffect(item.SourceCardID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  you,
		Source:   ctx.Source(),
		Question: neoCalamityName + " — choose up to two instant and/or sorcery cards with total mana value 6 or less from your graveyard and/or hand to cast without paying their mana costs",
		Cards:    cands,
		Min:      0,
		Max:      2,
		ThenKey:  neoCalamityPicked,
	})
	return nil
}

// neoCalamityPicked grants each picked card its free cast and offers the
// casts; Invoke Calamity is exiled once they are made (or at once when
// nothing was picked or can be cast).
var neoCalamityPicked = game.RegisterChooseCardsThenRule("choose-cards/neo-invoke-calamity-picked",
	neoTotalSpellManaValueAtMost(neoCalamityBudget),
	func(g *game.Game, r game.CardsChosen) error {
		var granted []uuid.UUID
		if r.Answered {
			for _, id := range r.Picked {
				if g.GrantCastPermissionOverCardForEffect(id, neoCalamityPermission(g, r.Chooser, r.Source)) {
					granted = append(granted, id)
				}
			}
		}
		if len(granted) > 0 && g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
			Chooser:      r.Chooser,
			Source:       r.Source,
			Reason:       neoCalamityName,
			Cards:        granted,
			Keyword:      game.MayCastKeywordFreeCast,
			AcceptLabel:  "Cast it free",
			DeclineLabel: "Leave it where it is",
			Then:         neoCalamityExile,
		}) {
			return nil
		}
		return g.ExileCardForEffect(r.Source)
	})

// neoCalamityExile is "Exile Invoke Calamity", once the casts are made.
var neoCalamityExile = game.RegisterChooseCardsThen("choose-cards/neo-invoke-calamity-exile",
	func(g *game.Game, r game.CardsChosen) error {
		return g.ExileCardForEffect(r.Source)
	})
