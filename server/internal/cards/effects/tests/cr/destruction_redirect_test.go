package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// destruction_redirect_test.go — #4526: once a CR 616 replacement in an
// order has sent a destroyed permanent to exile, the event is no longer a
// destruction (CR 701.8a), so a regeneration shield (CR 701.19a) and umbra
// armor (CR 702.89a), which replace only a destruction, stop applying to
// it (CR 616.1f: the next choice takes into account only the replacements
// that would now be applicable). The other order is the mirror image: a
// regeneration or umbra armor that goes first cancels the event, and the
// "exile it instead" has nothing left to replace.

// exileReplacer is a card that exiles a creature of the opponent's that
// would die, controlled by `me`.
type exileReplacer struct {
	name     string
	typeLine string
	oracle   string
}

var exileReplacers = []exileReplacer{
	{"Liesa, Forgotten Archangel", "Legendary Creature — Angel", effectstest.B19LiesaOracle},
	{"Rest in Peace", "Enchantment", effectstest.RestInPeaceOracle},
}

func pushExileReplacer(g *game.Game, controller uuid.UUID, r exileReplacer) uuid.UUID {
	if r.typeLine == "Enchantment" {
		return effectstest.PushCatalogPermanent(g, controller, r.name, r.typeLine, r.oracle, false)
	}
	return effectstest.B12Push(g, controller, r.name, r.typeLine, r.oracle, 4, 5)
}

// The exile first: the shield no longer applies, so it neither saves the
// creature nor is spent on the way.
func TestDestroyRedirectExileFirstLeavesARegenerationShieldNothingToReplace(t *testing.T) {
	for _, r := range exileReplacers {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pushExileReplacer(g, me.ID, r)
		bear := effectstest.ShieldedBear(g, opp.ID)
		order := effectstest.DestroyAndOrder(t, g, bear, opp.ID)
		effectstest.AnswerOrder(t, g, order, opp.ID, effectstest.OptionFrom(t, g, order, src))
		if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) || opp.Graveyard.Contains(bear) {
			t.Errorf("%s first: the Bear is not in exile (battlefield %v, graveyard %v)",
				r.name, g.Battlefield.Contains(bear), opp.Graveyard.Contains(bear))
		}
	}
}

// The shield first: the Bear regenerates (tapped, damage gone), the
// destruction is replaced, and there is nothing to exile.
func TestDestroyRedirectRegenerationShieldFirstSavesTheCreatureFromTheExile(t *testing.T) {
	for _, r := range exileReplacers {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pushExileReplacer(g, me.ID, r)
		bear := effectstest.ShieldedBear(g, opp.ID)
		order := effectstest.DestroyAndOrder(t, g, bear, opp.ID)
		effectstest.AnswerOrder(t, g, order, opp.ID, effectstest.OrderWith(order, effectstest.OptionFrom(t, g, order, src))[1])
		if !g.Battlefield.Contains(bear) || g.Exile.Contains(bear) || opp.Graveyard.Contains(bear) {
			t.Errorf("%s second: the Bear did not stay on the battlefield", r.name)
		}
		if !effectstest.IsTapped(g, bear) {
			t.Errorf("%s second: a regenerated creature is tapped (CR 701.19a)", r.name)
		}
		if got := effectstest.RegenShieldsOn(t, g, bear); got != 0 {
			t.Errorf("%s second: shields = %d, want 0 (the shield was used)", r.name, got)
		}
	}
}

// Umbra armor behaves the same way.
func TestDestroyRedirectExileFirstLeavesUmbraArmorNothingToReplace(t *testing.T) {
	for _, r := range exileReplacers {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pushExileReplacer(g, me.ID, r)
		bear, _ := effectstest.UmbraBear(t, g, opp.ID, 2)
		order := effectstest.DestroyAndOrder(t, g, bear, opp.ID)
		effectstest.AnswerOrder(t, g, order, opp.ID, effectstest.OptionFrom(t, g, order, src))
		if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) {
			t.Errorf("%s first: the Bear is not in exile", r.name)
		}
	}
}

func TestDestroyRedirectUmbraArmorFirstSavesTheCreatureFromTheExile(t *testing.T) {
	for _, r := range exileReplacers {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pushExileReplacer(g, me.ID, r)
		bear, umbra := effectstest.UmbraBear(t, g, opp.ID, 2)
		order := effectstest.DestroyAndOrder(t, g, bear, opp.ID)
		effectstest.AnswerOrder(t, g, order, opp.ID, effectstest.OrderWith(order, effectstest.OptionFrom(t, g, order, src))[1])
		if !g.Battlefield.Contains(bear) || g.Exile.Contains(bear) {
			t.Errorf("%s second: the Bear did not stay on the battlefield", r.name)
		}
		// The Aura is destroyed instead (to the graveyard, or on to
		// exile under Rest in Peace).
		if g.Battlefield.Contains(umbra) {
			t.Errorf("%s second: umbra armor should have destroyed the Aura", r.name)
		}
	}
}

// A restore at the prompt keeps the order a real choice, and the answer
// after it ends the destruction the same way.
func TestDestroyRedirectOrderIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	r := exileReplacers[0]
	src := pushExileReplacer(g, me.ID, r)
	bear := effectstest.ShieldedBear(g, opp.ID)
	effectstest.DestroyAndOrder(t, g, bear, opp.ID)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMove)
	opp = effectstest.SeatByID(g, opp.ID)
	order := effectstest.B06ReplacementOrderFor(g, opp.ID)
	if order == nil || len(order.ReplacementEffectIDs) != 2 {
		t.Fatalf("the restored prompt lost an option: %+v", g.PendingChoices)
	}
	effectstest.AnswerOrder(t, g, order, opp.ID, effectstest.OptionFrom(t, g, order, src))
	if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) {
		t.Error("the restored answer did not exile the Bear")
	}
}
