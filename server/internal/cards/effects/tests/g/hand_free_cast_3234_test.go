package testsg

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsbu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/bu"
	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	cardsri "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ri"
	cardsst "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/st"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3234: a "cast it without paying its mana cost" grant over a HAND card
// applies only to a cast that claims the alternative cost it carries
// (CastPermission.ForClaim). The grants were keyless, so the cast paid the
// printed cost. GrantCastPermissionOverCardForEffect now gives a hand grant
// that names a price the shared key.

// freeCastFixture seeds a Thrill of Possibility ({1}{R}, MV 2, additional
// cost: discard a card) and a card to discard into me's hand.
func freeCastFixture(g *game.Game, me *game.Player) (thrill, fodder uuid.UUID) {
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	thrill = effectstest.HandCardFull(me, "Thrill of Possibility", "Instant", "{1}{R}", effectstest.ThrillOfPossibilityOracle, []string{"R"})
	fodder = effectstest.HandCardFull(me, "Fodder", "Sorcery", "{4}", "", nil)
	return thrill, fodder
}

// answerFreeCastPick checks the free-cast offer names id. Inside a
// resolution (#4250) the offer is the resolution_cast prompt and the cast
// itself is its answer, so there is nothing to pick; outside one it is the
// keyed pick (#4246), answered with id.
func answerFreeCastPick(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, id)
		return
	}
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no free-cast pick")
	}
	if !effectstest.HasID(pick.ChooseCards, id) {
		t.Fatalf("card not offered: %v", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{id}); err != nil {
		t.Fatalf("choosing the free cast: %v", err)
	}
}

// assertCastsFree casts id with an empty pool: refused at the printed
// cost, accepted when the caster claims the granted offer, with the
// additional cost (discard) still paid. Inside a resolution (#4250) the
// offer's prompt holds a hand card's cast to the granted claim, so the
// unclaimed cast IS the free one.
func assertCastsFree(t *testing.T, g *game.Game, me *game.Player, id, fodder uuid.UUID) {
	t.Helper()
	params := game.CastSpellParams{Strict: true, DiscardIDs: []uuid.UUID{fodder}}
	if effectstest.ResolutionCastOfferFor(g, me.ID) == nil {
		err := g.CastSpell(me.ID, id, params)
		if err == nil || !strings.Contains(err.Error(), "insufficient mana") {
			t.Fatalf("an unclaimed cast must pay the printed cost, got %v", err)
		}
		params.AlternativeCost = game.AltCostKeyGrantedCost
	}
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	if me.Hand.Contains(fodder) {
		t.Error("the additional cost (discard) was not paid")
	}
	if me.Hand.Contains(id) {
		t.Error("the spell did not leave the hand")
	}
}

func TestKariZevsExpertiseGrantedCardCastsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	thrill, fodder := freeCastFixture(g, me)
	big := effectstest.HandCardFull(me, "Big", "Instant", "{3}", "", nil)

	effectstest.CastCatalogSpell(t, g, "Kari Zev's Expertise", "Sorcery", effectstest.KariZevsExpertiseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	answerFreeCastPick(t, g, me, thrill)
	assertCastsFree(t, g, me, thrill, fodder)
	// The grant never covered the card it was not made over.
	err := g.CastSpell(me.ID, big, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost})
	if err == nil {
		t.Error("an ungranted card cast free")
	}
}

func TestKariZevsExpertiseDoesNotOfferManaValueThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.HandCardFull(me, "Big", "Instant", "{3}", "", nil)
	effectstest.OfferFrom(t, g, me, func(ctx *effects.Context) error { return cardsj.KariZevsExpertiseOfferFreeCast(ctx, me.ID) })
	if pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); pick != nil {
		t.Errorf("mana value 3 offered for free: %v", pick.ChooseCards)
	}
}

func TestRishkarsExpertiseGrantedCardCastsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thrill, fodder := freeCastFixture(g, me)
	effectstest.OfferFrom(t, g, me, func(ctx *effects.Context) error { return cardsri.RishkarsExpertiseOfferFreeCast(ctx, me.ID) })
	answerFreeCastPick(t, g, me, thrill)
	assertCastsFree(t, g, me, thrill, fodder)
}

func TestRishkarsExpertiseDoesNotOfferManaValueSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.HandCardFull(me, "Huge", "Instant", "{6}", "", nil)
	effectstest.OfferFrom(t, g, me, func(ctx *effects.Context) error { return cardsri.RishkarsExpertiseOfferFreeCast(ctx, me.ID) })
	if pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); pick != nil {
		t.Errorf("mana value 6 offered for free: %v", pick.ChooseCards)
	}
}

func TestBusterSwordGrantedCardCastsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thrill, fodder := freeCastFixture(g, me)
	effectstest.OfferFrom(t, g, me, func(ctx *effects.Context) error { return cardsbu.BusterSwordOfferFreeCast(ctx, me.ID, 2) })
	answerFreeCastPick(t, g, me, thrill)
	assertCastsFree(t, g, me, thrill, fodder)
}

func TestSurtlandElementalistGrantedCardCastsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thrill, fodder := freeCastFixture(g, me)
	effectstest.OfferFrom(t, g, me, func(ctx *effects.Context) error { return cardsst.SurtlandElementalistOfferFreeCast(ctx, me.ID) })
	answerFreeCastPick(t, g, me, thrill)
	assertCastsFree(t, g, me, thrill, fodder)
}

func TestGreatIntelligencesPlanGrantedCardCastsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.VillainFillLibrary(g, me, 6)
	effectstest.VillainEmptyHand(me)
	thrill, fodder := freeCastFixture(g, me)
	effectstest.VillainCastPlan(t, g, opp)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 1)
	answerFreeCastPick(t, g, me, thrill)
	assertCastsFree(t, g, me, thrill, fodder)
}
