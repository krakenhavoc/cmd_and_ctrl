package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// worlds_within_worlds_test.go — Worlds Within Worlds (#4455), end to end on
// the real catalog card: the sweep, one pick per player in APNAP order, the
// simultaneous entry, the exiled cards coming home and the spell exiling
// itself.

const worldsWithinWorldsOracle = "e25062ba-7176-4d59-ac3e-26d19b26f02a"

func TestWorldsWithinWorldsIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(worldsWithinWorldsOracle)
	if !ok {
		t.Fatal("Worlds Within Worlds is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Worlds Within Worlds: completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
}

// wwwTable is a four-player board for the card: creatures on three
// battlefields (one a token), creature cards in three hands.
type wwwTable struct {
	g                *game.Game
	a, b, c, d       *game.Player // in APNAP order: a is the active player
	ax, by, token    uuid.UUID    // creatures on the battlefield
	a1, a2, bolt     uuid.UUID    // a's hand
	b1               uuid.UUID    // b's hand
	c1, c2           uuid.UUID    // c's hand
	spell            uuid.UUID
	castWorldsWithin func(t *testing.T)
}

func newWWWTable(t *testing.T) *wwwTable {
	t.Helper()
	g := newCatalogGame(t)
	n := len(g.Seats)
	seat := func(i int) *game.Player { return g.Seats[(g.Turn.ActiveSeat+i)%n] }
	w := &wwwTable{g: g, a: seat(0), b: seat(1), c: seat(2), d: seat(3)}
	w.ax = b12Creature(g, w.a.ID, "Ax", "Creature — Ox", 2, 2)
	w.by = b12Creature(g, w.b.ID, "By", "Creature — Bear", 2, 2)
	w.token = b12Creature(g, w.b.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	w.a1 = handCardForTest(w.a, "Aaa", "Creature — Elk", "")
	w.a2 = handCardForTest(w.a, "Abb", "Creature — Elk", "")
	w.bolt = handCardForTest(w.a, "Bolt", "Instant", "")
	w.b1 = handCardForTest(w.b, "Bee", "Creature — Insect", "")
	w.c1 = handCardForTest(w.c, "Cee", "Creature — Cat", "")
	w.c2 = handCardForTest(w.c, "Cey", "Creature — Cat", "")
	w.castWorldsWithin = func(t *testing.T) {
		t.Helper()
		w.spell = castCatalogSpell(t, g, "Worlds Within Worlds", "Sorcery", worldsWithinWorldsOracle, nil)
		passPriorityAroundTable(t, g)
	}
	return w
}

// The whole card: the sweep, APNAP picks (each player is asked in turn, the
// next only after the last has answered), the simultaneous entry under each
// owner's control, the exiled cards home, the token gone, the spell in exile.
func TestWorldsWithinWorldsWholeCard(t *testing.T) {
	w := newWWWTable(t)
	g := w.g
	w.castWorldsWithin(t)

	// Exile all creatures: both battlefields are bare, the token is gone.
	if g.Battlefield.Contains(w.ax) || g.Battlefield.Contains(w.by) || g.Battlefield.Contains(w.token) {
		t.Fatal("a creature survived the sweep")
	}

	// APNAP (CR 101.4): the active player alone is asked first.
	pick := chooseCardsChoiceFor(g, w.a.ID)
	if pick == nil {
		t.Fatal("the active player was not asked")
	}
	for _, p := range []*game.Player{w.b, w.c, w.d} {
		if chooseCardsChoiceFor(g, p.ID) != nil {
			t.Fatalf("%s was asked before the active player answered", p.ID)
		}
	}
	// The hand is read after the sweep: only the two creature cards in it,
	// not the exiled creature and not the instant.
	if len(pick.ChooseCards) != 2 || !hasID(pick.ChooseCards, w.a1) || !hasID(pick.ChooseCards, w.a2) {
		t.Fatalf("candidates %v, want exactly the two creature cards in hand", pick.ChooseCards)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Errorf("bounds %d..%d, want 0..2 — \"any number\"", pick.ChooseMin, pick.ChooseMax)
	}
	answerChooseCards(t, g, w.a.ID, w.a1)
	// Nothing enters until every player has chosen (CR 101.4: then the
	// actions happen simultaneously).
	if g.Battlefield.Contains(w.a1) {
		t.Fatal("a pick entered before the later players chose")
	}

	pick = chooseCardsChoiceFor(g, w.b.ID)
	if pick == nil {
		t.Fatal("the next player was not asked")
	}
	answerChooseCards(t, g, w.b.ID) // none: "any number" includes zero

	// The fourth player holds no creature card: not asked.
	pick = chooseCardsChoiceFor(g, w.c.ID)
	if pick == nil {
		t.Fatal("the third player was not asked")
	}
	answerChooseCards(t, g, w.c.ID, w.c1, w.c2)
	if chooseCardsChoiceFor(g, w.d.ID) != nil {
		t.Error("a player with no creature card in hand was asked")
	}
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{w.a1, w.c1, w.c2} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("a picked creature card %v is not on the battlefield", id)
		}
	}
	if g.Battlefield.Contains(w.a2) || !w.a.Hand.Contains(w.a2) {
		t.Error("an unpicked creature card left the hand")
	}
	// Each entered under its owner's control.
	for _, c := range g.Battlefield.Cards {
		if (c.InstanceID == w.a1 && c.Controller != w.a.ID) || (c.InstanceID == w.c1 && c.Controller != w.c.ID) {
			t.Errorf("%s entered under %v, want its owner", c.Name, c.Controller)
		}
	}
	// Then all cards exiled this way go to their owners' hands.
	if !w.a.Hand.Contains(w.ax) || !w.b.Hand.Contains(w.by) {
		t.Error("an exiled creature card did not come home to its owner's hand")
	}
	if g.Exile.Contains(w.ax) || g.Exile.Contains(w.by) {
		t.Error("an exiled creature card is still in exile")
	}
	// A token is not a card: it ceased to exist and went to no hand.
	if g.Exile.Contains(w.token) || w.b.Hand.Contains(w.token) || g.Battlefield.Contains(w.token) {
		t.Error("the exiled token should have ceased to exist")
	}
	// Exile Worlds Within Worlds: not the graveyard.
	if !g.Exile.Contains(w.spell) {
		t.Error("Worlds Within Worlds is not in exile")
	}
	for _, p := range g.Seats {
		if p.Graveyard.Contains(w.spell) {
			t.Error("Worlds Within Worlds went to a graveyard")
		}
	}
}

// Every prompt is a restore point, and the answers carry on after it: the
// players still to ask, the picks so far and the exiled cards ride the
// prompt as data.
func TestWorldsWithinWorldsEveryPromptIsARestorePoint(t *testing.T) {
	w := newWWWTable(t)
	w.castWorldsWithin(t)
	g := w.g
	ids := map[string]uuid.UUID{"a1": w.a1, "c1": w.c1, "ax": w.ax, "by": w.by, "spell": w.spell}
	players := []uuid.UUID{w.a.ID, w.b.ID, w.c.ID}
	picks := [][]uuid.UUID{{w.a1}, nil, {w.c1}}
	for i, who := range players {
		g = restoreRoundTrip(t, g, true)
		answerChooseCards(t, g, who, picks[i]...)
	}
	g = restoreRoundTrip(t, g, true)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids["a1"]) || !g.Battlefield.Contains(ids["c1"]) {
		t.Error("a restored pick did not enter")
	}
	byHand := g.PlayerByIDForEffect(w.b.ID).Hand
	axHand := g.PlayerByIDForEffect(w.a.ID).Hand
	if !byHand.Contains(ids["by"]) || !axHand.Contains(ids["ax"]) {
		t.Error("a restored exile did not come home")
	}
	if !g.Exile.Contains(ids["spell"]) {
		t.Error("the spell is not in exile after the restore")
	}
}

// The answers are checked: only the asked player may answer, only a creature
// card from their hand is a legal pick.
func TestWorldsWithinWorldsRefusesIllegalPicks(t *testing.T) {
	w := newWWWTable(t)
	w.castWorldsWithin(t)
	g := w.g
	pick := chooseCardsChoiceFor(g, w.a.ID)
	if pick == nil {
		t.Fatal("the active player was not asked")
	}
	if err := g.ResolveChooseCards(pick.ID, w.b.ID, nil); err == nil {
		t.Error("another player answered the active player's prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, w.a.ID, []uuid.UUID{w.bolt}); err == nil {
		t.Error("an instant was accepted as a creature card")
	}
	if err := g.ResolveChooseCards(pick.ID, w.a.ID, []uuid.UUID{w.c1}); err == nil {
		t.Error("another player's creature card was accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, w.a.ID, []uuid.UUID{w.a1, w.a1}); err == nil {
		t.Error("the same card was accepted twice")
	}
	if chooseCardsChoiceFor(g, w.a.ID) == nil {
		t.Error("a refused answer closed the prompt")
	}
}

// With no creature card in any hand nobody is asked; the exiled creatures
// still come home and the spell still exiles itself.
func TestWorldsWithinWorldsWithNobodyToAsk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Ox", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	handCardForTest(me, "Bolt", "Instant", "")
	spell := castCatalogSpell(t, g, "Worlds Within Worlds", "Sorcery", worldsWithinWorldsOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if chooseCardsChoiceFor(g, p.ID) != nil {
			t.Fatalf("%s was asked with no creature card in hand", p.ID)
		}
	}
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("an exiled creature card did not come home")
	}
	if !g.Exile.Contains(spell) {
		t.Error("Worlds Within Worlds is not in exile")
	}
}

// The card costs {5}{G}{U}: a short pool is refused and takes nothing, the
// right pool casts it and leaves it exiled on resolution.
func TestWorldsWithinWorldsCostsSevenMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	card := handCardFull(me, "Worlds Within Worlds", "Sorcery", "{5}{G}{U}", worldsWithinWorldsOracle, []string{"G", "U"})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{G}{U}")
	err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("casting on six mana: %v, want ErrInsufficientMana", err)
	}
	if !me.Hand.Contains(card) {
		t.Fatal("a refused cast took the card out of the hand")
	}
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting on seven mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(card) {
		t.Error("Worlds Within Worlds is not in exile")
	}
}
