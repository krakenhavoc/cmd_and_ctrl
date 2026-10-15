package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// abhorrent_oculus_test.go — #3536: "exile six cards from your graveyard"
// as a SPELL's additional cost, and the manifest dread Oculus makes on each
// opponent's upkeep.

const (
	abhorrentOculusOracle = "b1a0e129-27f0-4a09-b00d-497e18777306"
	oculusCost            = "{2}{U}"
)

func castOculus(t *testing.T, g *game.Game, exile ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, "Abhorrent Oculus", "Creature — Eye", oculusCost, abhorrentOculusOracle,
		game.CastSpellParams{ExileIDs: exile})
}

func graveyardCreatures(p *game.Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = pushGraveyardCardForTest(p, "Fuel")
	}
	return ids
}

// Six named cards are exiled with the spell on the stack: Syr Konrad's
// "a creature card leaves your graveyard" triggers above Oculus and
// resolves first, and the six are recorded for ctx.Exiled().
func TestAbhorrentOculusExilesSixWithTheSpellOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Syr Konrad, the Grim", "Legendary Creature — Human Knight", b02SyrKonradOracle, false)
	fuel := graveyardCreatures(me, 7)
	before := opp.Life

	spell, err := castOculus(t, g, fuel[:6]...)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	for _, id := range fuel[:6] {
		if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
			t.Errorf("card %s was not exiled by the cost", id)
		}
	}
	if !me.Graveyard.Contains(fuel[6]) {
		t.Error("the seventh card, which was not named, left the graveyard")
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("Oculus is not on the stack")
	}
	var exiled []uuid.UUID
	g.WithWriteLock(func() { exiled = NewContext(g, g.StackMeta[spell]).Exiled() })
	if !slices.Equal(exiled, fuel[:6]) {
		t.Errorf("ctx.Exiled() = %v, want the six named cards in order", exiled)
	}

	// Konrad's six triggers are above the spell: Oculus is still on the
	// stack when the last of them has dealt its damage.
	for g.Stack.Contains(spell) {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if got := before - opp.Life; got != 6 {
		t.Errorf("Syr Konrad dealt %d to the opponent before Oculus resolved, want 6 (one per card)", got)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, spell); !ok {
		t.Fatal("Oculus did not resolve onto the battlefield")
	}
	if abilities := effectiveAbilities(t, g, spell); !slices.Contains(abilities, "flying") {
		t.Errorf("Oculus abilities = %v, want flying", abilities)
	}
}

// A countered Oculus does not give the cards back: they were a cost, not an
// effect (CR 601.2h).
func TestAbhorrentOculusCounteredStillExilesTheSix(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fuel := graveyardCreatures(me, 6)

	spell, err := castOculus(t, g, fuel...)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	counterNow(t, g, spell)
	passPriorityAroundTable(t, g)
	for _, id := range fuel {
		if !g.Exile.Contains(id) {
			t.Errorf("card %s left exile when the spell was countered", id)
		}
	}
	if me.Graveyard.Size() != 1 {
		t.Errorf("graveyard = %d, want only the countered Oculus", me.Graveyard.Size())
	}
}

// CR 118.3: five cards cannot pay six, whatever is named, and nothing moves.
func TestAbhorrentOculusIsRefusedWithFiveCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fuel := graveyardCreatures(me, 5)

	spell, err := castOculus(t, g, fuel...)
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("cast with five named = %v, want ErrInvalidParam", err)
	}
	if g.Stack.Contains(spell) || g.Exile.Size() != 0 || me.Graveyard.Size() != 5 {
		t.Errorf("a refused cast moved cards: stack %v, exile %d, graveyard %d",
			g.Stack.Contains(spell), g.Exile.Size(), me.Graveyard.Size())
	}
}

// A copy of the spell keeps the record of what was exiled (CR 707.10).
func TestACopyOfAbhorrentOculusKeepsTheExiledCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fuel := graveyardCreatures(me, 6)
	spell, err := castOculus(t, g, fuel...)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(spell, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
	copies := 0
	for id, item := range g.StackMeta {
		if id == spell || item == nil || item.Kind != game.StackItemSpell {
			continue
		}
		copies++
		if !slices.Equal(item.Paid.Exiled, fuel) {
			t.Errorf("the copy's Paid.Exiled = %v, want the original's six", item.Paid.Exiled)
		}
	}
	if copies != 1 {
		t.Fatalf("found %d copies on the stack, want 1", copies)
	}
}

// answerManifestDread settles whatever Oculus's trigger left: the stack
// first, then the manifest-dread pick (the first of the two looked-at
// cards), until nothing is waiting.
func answerManifestDread(t *testing.T, g *game.Game, who uuid.UUID) int {
	t.Helper()
	answered := 0
	for i := 0; i < 16; i++ {
		passPriorityAroundTable(t, g)
		p := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, who)
		if p == nil {
			return answered
		}
		if err := g.ResolveChooseCards(p.ID, who, p.ChooseCards[:1]); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		answered++
	}
	t.Fatal("the manifest dread prompts never ran out")
	return answered
}

// "At the beginning of each opponent's upkeep, manifest dread": each of the
// three opponents' upkeeps triggers it, the controller's own does not.
func TestAbhorrentOculusManifestsDreadOnEachOpponentsUpkeepOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	oculus := pushCatalogPermanent(g, me.ID, "Abhorrent Oculus", "Creature — Eye", abhorrentOculusOracle, false)

	// One full turn cycle: seats 1, 2 and 3 take their upkeeps, then mine.
	manifested := 0
	for _, seat := range []int{1, 2, 3} {
		advanceToUpkeepOf(t, g, seat)
		if triggerOnStack(g, oculus) == nil && len(g.PendingTriggers) == 0 {
			t.Fatalf("seat %d's upkeep did not trigger Oculus", seat)
		}
		manifested += answerManifestDread(t, g, me.ID)
	}
	if manifested != 3 {
		t.Errorf("manifest dread ran %d times across three opponents' upkeeps, want 3", manifested)
	}
	advanceToUpkeepOf(t, g, 0)
	if triggerOnStack(g, oculus) != nil || len(g.PendingTriggers) != 0 {
		t.Error("Oculus triggered on its controller's own upkeep")
	}
}

// kickerThatExiles is a well-formed optional cost (it is a Kicker) that also
// exiles two graveyard cards, so the only thing Register can object to is
// the exile.
func kickerThatExiles() game.AdditionalCost {
	k := Kicker("{1}")
	k.ExileCards = ExileCardsFromGraveyard(2, "two cards", nil)
	return k
}

// Register's guards (ADR 0189 §2): the hand form, a non-positive count, an
// optional cost and an either/or branch each compile and then charge
// something no card prints, so they fail at boot.
func TestRegisterRefusesMisplacedExileCardsCosts(t *testing.T) {
	for name, spec := range map[string]Spec{
		"the hand form": {OracleID: "exilecards-hand", Name: "EC Hand",
			AdditionalCost: &game.AdditionalCost{ExileCards: ExileCardsFromHand(1, "a card", nil), Label: "x"}},
		"a zero count": {OracleID: "exilecards-zero", Name: "EC Zero",
			AdditionalCost: &game.AdditionalCost{ExileCards: ExileCardsFromGraveyard(0, "no cards", nil), Label: "x"}},
		"a negative count": {OracleID: "exilecards-neg", Name: "EC Negative",
			AdditionalCost: ExileFromGraveyardCost(-2, "cards", nil)},
		"an optional cost": {OracleID: "exilecards-optional", Name: "EC Optional",
			OptionalCosts: []game.AdditionalCost{kickerThatExiles()}},
		"an either/or branch": {OracleID: "exilecards-branch", Name: "EC Branch",
			AdditionalCost: EitherCost(
				ExileFromGraveyardCost(2, "two cards", nil).Keyed("exile"),
				ManaAdditionalCost("{2}").Keyed("mana"))},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("Register accepted %s", name)
				}
			}()
			Register(spec)
		})
	}
}
