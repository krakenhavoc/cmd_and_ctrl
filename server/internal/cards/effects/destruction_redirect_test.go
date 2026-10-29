package effects

import (
	"testing"

	"github.com/google/uuid"

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
	{"Liesa, Forgotten Archangel", "Legendary Creature — Angel", b19LiesaOracle},
	{"Rest in Peace", "Enchantment", restInPeaceOracle},
}

// optionFrom is the ID of the ordering prompt's option whose source is
// `src`.
func optionFrom(t *testing.T, g *game.Game, order *game.PendingChoice, src uuid.UUID) game.ReplacementEffectID {
	t.Helper()
	for _, id := range order.ReplacementEffectIDs {
		var got uuid.UUID
		g.ReadSnapshot(func() { _, got = g.ReplacementOptionMetaForEffect(id) })
		if got == src {
			return id
		}
	}
	t.Fatalf("no option on the prompt comes from %s: %v", src, order.ReplacementEffectIDs)
	return 0
}

// shieldedBear puts a 2/2 Bear under the opponent's control and gives it
// one regeneration shield; it returns the Bear.
func shieldedBear(g *game.Game, owner uuid.UUID) uuid.UUID {
	bear := b12Creature(g, owner, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.RegenerateForEffect(bear) })
	return bear
}

// destroyAndOrder destroys `victim` and returns the opponent's CR 616
// ordering prompt: the replacer against the victim's protection.
func destroyAndOrder(t *testing.T, g *game.Game, victim uuid.UUID, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	order := b06ReplacementOrderFor(g, chooser)
	if order == nil || len(order.ReplacementEffectIDs) != 2 {
		t.Fatalf("want the victim's controller asked to order two replacements (CR 616.1), got %+v", g.PendingChoices)
	}
	return order
}

func pushExileReplacer(g *game.Game, controller uuid.UUID, r exileReplacer) uuid.UUID {
	if r.typeLine == "Enchantment" {
		return pushCatalogPermanent(g, controller, r.name, r.typeLine, r.oracle, false)
	}
	return b12Push(g, controller, r.name, r.typeLine, r.oracle, 4, 5)
}

func answerOrder(t *testing.T, g *game.Game, order *game.PendingChoice, chooser uuid.UUID, first game.ReplacementEffectID) {
	t.Helper()
	if err := g.ResolveReplacementOrder(order.ID, chooser, orderWith(order, first)); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if left := b06ReplacementOrderFor(g, chooser); left != nil {
		t.Errorf("a second ordering prompt is open: %+v", left)
	}
}

// The exile first: the shield no longer applies, so it neither saves the
// creature nor is spent on the way.
func TestDestroyRedirectExileFirstLeavesARegenerationShieldNothingToReplace(t *testing.T) {
	for _, r := range exileReplacers {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pushExileReplacer(g, me.ID, r)
		bear := shieldedBear(g, opp.ID)
		order := destroyAndOrder(t, g, bear, opp.ID)
		answerOrder(t, g, order, opp.ID, optionFrom(t, g, order, src))
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
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pushExileReplacer(g, me.ID, r)
		bear := shieldedBear(g, opp.ID)
		order := destroyAndOrder(t, g, bear, opp.ID)
		answerOrder(t, g, order, opp.ID, orderWith(order, optionFrom(t, g, order, src))[1])
		if !g.Battlefield.Contains(bear) || g.Exile.Contains(bear) || opp.Graveyard.Contains(bear) {
			t.Errorf("%s second: the Bear did not stay on the battlefield", r.name)
		}
		if !isTapped(g, bear) {
			t.Errorf("%s second: a regenerated creature is tapped (CR 701.19a)", r.name)
		}
		if got := regenShieldsOn(t, g, bear); got != 0 {
			t.Errorf("%s second: shields = %d, want 0 (the shield was used)", r.name, got)
		}
	}
}

// Umbra armor behaves the same way.
func TestDestroyRedirectExileFirstLeavesUmbraArmorNothingToReplace(t *testing.T) {
	for _, r := range exileReplacers {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pushExileReplacer(g, me.ID, r)
		bear, _ := umbraBear(t, g, opp.ID, 2)
		order := destroyAndOrder(t, g, bear, opp.ID)
		answerOrder(t, g, order, opp.ID, optionFrom(t, g, order, src))
		if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) {
			t.Errorf("%s first: the Bear is not in exile", r.name)
		}
	}
}

func TestDestroyRedirectUmbraArmorFirstSavesTheCreatureFromTheExile(t *testing.T) {
	for _, r := range exileReplacers {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pushExileReplacer(g, me.ID, r)
		bear, umbra := umbraBear(t, g, opp.ID, 2)
		order := destroyAndOrder(t, g, bear, opp.ID)
		answerOrder(t, g, order, opp.ID, orderWith(order, optionFrom(t, g, order, src))[1])
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	r := exileReplacers[0]
	src := pushExileReplacer(g, me.ID, r)
	bear := shieldedBear(g, opp.ID)
	destroyAndOrder(t, g, bear, opp.ID)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMove)
	opp = seatByID(g, opp.ID)
	order := b06ReplacementOrderFor(g, opp.ID)
	if order == nil || len(order.ReplacementEffectIDs) != 2 {
		t.Fatalf("the restored prompt lost an option: %+v", g.PendingChoices)
	}
	answerOrder(t, g, order, opp.ID, optionFrom(t, g, order, src))
	if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) {
		t.Error("the restored answer did not exile the Bear")
	}
}
