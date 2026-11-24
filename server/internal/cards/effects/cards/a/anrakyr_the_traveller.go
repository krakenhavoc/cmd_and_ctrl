package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anrakyr the Traveller — Legendary Artifact Creature — Necron {4}{B},
// 4/4 (Warhammer 40,000 Commander, #5948 GAP2 b07):
//
//	"Lord of the Pyrrhian Legions — Whenever Anrakyr the Traveller
//	 attacks, you may cast an artifact spell from your hand or graveyard
//	 by paying life equal to its mana value rather than paying its mana
//	 cost."
//
// The cast is made during the trigger's resolution (CR 608.2g), through
// the resolution_cast prompt (#6091). A permission opens one zone, so the
// offer is two permissions, one over the artifact cards in the hand and
// one over those in the graveyard, tied into ONE budget (BudgetGroup, a
// count of one): "an artifact spell" is one cast from either zone. Each is
// Bolas's Citadel's price, life equal to the spell's mana value as an
// alternative cost (CR 118.9, 119.4, 202.3), so additional costs are still
// paid and a player without the life is not offered it. "Artifact spell"
// is judged on the face being cast (PermissionFilter.CardType), so an
// artifact card's non-artifact Adventure is not offered. Any timing, cast
// only; declining leaves every card where it is.
//
// No simplification.
func init() {
	const name = "Anrakyr the Traveller"
	effects.Register(effects.Spec{
		OracleID:     "32476743-c9d4-49ca-bec2-0669c215841b",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Lord of the Pyrrhian Legions — you may cast an artifact spell from your hand or graveyard by paying life equal to its mana value",
				anrakyrOfferCast),
		},
	})
}

func anrakyrOfferCast(g *game.Game, item *game.StackItem) error {
	const name = "Anrakyr the Traveller"
	ctx := effects.NewContext(g, item)
	you := item.Controller
	base := game.CastPermission{
		Player:               you,
		Duration:             g.UntilEndOfTurnDuration(),
		AltCostKey:           "anrakyr_the_traveller",
		LifeEqualToManaValue: true,
		Filter:               game.PermissionFilter{CardType: "artifact"},
		Timing:               game.TimingFlash,
		CastOnly:             true,
		CastsLeft:            1,
		BudgetGroup:          uuid.New(),
		LapseOnPass:          game.LapseStaysInHand,
		Source:               ctx.Source(),
		SourceName:           name,
		Label:                "Lord of the Pyrrhian Legions — cast an artifact spell by paying life equal to its mana value",
	}
	var offered []uuid.UUID
	for _, zone := range []game.ZoneKind{game.ZoneHand, game.ZoneGraveyard} {
		perm := base
		perm.Zone = zone
		var cards []game.Card
		for _, c := range anrakyrZoneCards(g, you, zone) {
			if perm.OffersASpell(c) {
				cards = append(cards, c)
				offered = append(offered, c.InstanceID)
			}
		}
		g.GrantCastPermissionToCardsForEffect(perm, cards)
	}
	if len(offered) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       ctx.Source(),
		Reason:       name,
		Cards:        offered,
		Keyword:      game.MayCastKeywordPaidCast,
		AcceptLabel:  "Cast it, paying life",
		DeclineLabel: effects.DeclineLeaveWhereItIs,
	})
	return nil
}

// anrakyrZoneCards is the cards in `player`'s hand or graveyard.
func anrakyrZoneCards(g *game.Game, player uuid.UUID, zone game.ZoneKind) []game.Card {
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return nil
	}
	z := p.Hand
	if zone == game.ZoneGraveyard {
		z = p.Graveyard
	}
	if z == nil {
		return nil
	}
	return append([]game.Card(nil), z.Cards...)
}
