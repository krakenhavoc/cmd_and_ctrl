package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Face of Boe — Legendary Creature — Alien Advisor {1}{U}{R}{W}, 0/4
// (Doctor Who, #5948 GAP2 b07):
//
//	"{T}: You may cast a spell with suspend from your hand. If you do, pay
//	 its suspend cost rather than its mana cost. Activate only as a
//	 sorcery."
//
// The cast is made during the ability's resolution (CR 608.2g), through
// the resolution_cast prompt over the cards in the hand that have suspend
// (CR 702.62a, the card's printed Suspend N—cost). Each card's price is
// its own suspend cost, an alternative cost (CR 118.9), so each is its own
// permission with that Cost; they are tied into ONE budget (BudgetGroup, a
// count of one), since it is one spell (#6091). Additional costs, cost
// increases and reductions still apply (CR 118.9d). A card with suspend
// and no mana cost (Ancestral Vision) is castable this way. Casting it is
// not casting it through suspend: it gains no haste, and it gets no time
// counters. Declining keeps the cards in the hand.
//
// No simplification.
func init() {
	const name = "The Face of Boe"
	effects.Register(effects.Spec{
		OracleID:     "acef7666-53b1-43c5-8107-236bea22a8b3",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: You may cast a spell with suspend from your hand. If you do, pay its suspend cost rather than its mana cost. Activate only as a sorcery.",
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Effect:       faceOfBoeOffer,
		}},
	})
}

// FaceOfBoeAltCostKey is the alternative cost The Face of Boe's offer
// claims. An on-disk value.
const FaceOfBoeAltCostKey = "the_face_of_boe_suspend_cost"

func faceOfBoeOffer(g *game.Game, item *game.StackItem) error {
	const name = "The Face of Boe"
	you := item.Controller
	p := g.PlayerByIDForEffect(you)
	if p == nil || p.Hand == nil {
		return nil
	}
	ref, _ := g.SourceObjectForEffect(item)
	budget := uuid.New()
	var offered []uuid.UUID
	for _, c := range append([]game.Card(nil), p.Hand.Cards...) {
		cost, ok := faceOfBoeSuspendCost(c)
		if !ok {
			continue
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:      you,
			Zone:        game.ZoneHand,
			Duration:    g.UntilEndOfTurnDuration(),
			AltCostKey:  FaceOfBoeAltCostKey,
			Cost:        cost,
			Timing:      game.TimingFlash,
			CastOnly:    true,
			CastsLeft:   1,
			BudgetGroup: budget,
			LapseOnPass: game.LapseStaysInHand,
			Source:      ref.ID,
			SourceName:  name,
			Label:       name + " — cast it for its suspend cost " + cost,
		}, []game.Card{c})
		offered = append(offered, c.InstanceID)
	}
	if len(offered) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       ref.ID,
		Reason:       name,
		Cards:        offered,
		Keyword:      game.MayCastKeywordPaidCast,
		AcceptLabel:  "Cast it for its suspend cost",
		DeclineLabel: "Keep it in your hand",
	})
	return nil
}

// faceOfBoeSuspendCost is the printed suspend cost of a card with suspend.
func faceOfBoeSuspendCost(c game.Card) (string, bool) {
	for _, sa := range game.SpecialActionsFor(game.CatalogKey(c)) {
		if sa.Kind == game.SpecialActionSuspend {
			return sa.Cost, true
		}
	}
	return "", false
}
