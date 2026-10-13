package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_a_krark_tersa_test.go — Krark, the Thumbless and Tersa
// Lightshatter (#3213).

// krarkCast casts a Divination with Krark out, calls the coin heads,
// and returns the game, the hand size before the flip and whether the
// flip was won. Each fresh game mints its own RNG key, so the call does
// not decide the outcome (ADR 0054).
func krarkCast(t *testing.T) (g *game.Game, handAfterCast int, divination uuid.UUID, won bool) {
	t.Helper()
	g = newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Krark, the Thumbless", krarkThumblessOrcl, "Legendary Creature — Goblin Wizard", 2, 2)
	divination = castCatalogSpell(t, g, "Divination", "Sorcery", spellControlDivinationOracle, nil)
	handAfterCast = me.Hand.Size()
	passPriorityAroundTable(t, g)
	flip := latestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if flip == nil {
		t.Fatal("Krark asked for no coin call")
	}
	if flip.Chooser != me.ID {
		t.Fatalf("the coin call went to %s, want Krark's controller", flip.Chooser)
	}
	if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	for _, ev := range randomEvents(g, game.EventFlipCoin) {
		won = ev.Won
	}
	passPriorityAroundTable(t, g)
	return g, handAfterCast, divination, won
}

func TestKrarkWinCopiesTheSpellAndLossReturnsItToHand(t *testing.T) {
	sawWin, sawLoss := false, false
	for i := 0; i < 64 && !(sawWin && sawLoss); i++ {
		g, before, div, won := krarkCast(t)
		me := g.Seats[0]
		if won {
			sawWin = true
			// Divination (draw 2) and its copy: four cards, and the
			// original is in the graveyard (a copy is not a card).
			if got := me.Hand.Size(); got != before+4 {
				t.Errorf("won flip: hand %d -> %d, want +4 (Divination and its copy)", before, got)
			}
			if !me.Graveyard.Contains(div) {
				t.Error("won flip: the original Divination should be in the graveyard")
			}
			if got := graveyardSize(g, me.ID); got != 1 {
				t.Errorf("won flip: graveyard = %d, want 1 — a copy is not a card", got)
			}
			continue
		}
		sawLoss = true
		if !me.Hand.Contains(div) {
			t.Error("lost flip: the spell should be back in its owner's hand")
		}
		if got := me.Hand.Size(); got != before+1 {
			t.Errorf("lost flip: hand %d -> %d, want +1 (just the returned spell, nothing drawn)", before, got)
		}
		if me.Graveyard.Contains(div) {
			t.Error("lost flip: a returned spell is not countered and must not be in the graveyard")
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("64 games never showed both outcomes (win=%v loss=%v)", sawWin, sawLoss)
	}
}

func TestKrarkIgnoresCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Krark, the Thumbless", krarkThumblessOrcl, "Legendary Creature — Goblin Wizard", 2, 2)
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoiceCoinCall) != nil {
		t.Error("Krark must not flip for a creature spell")
	}
}

// --- Tersa Lightshatter ---------------------------------------------

func tersaGame(t *testing.T, graveyard int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < graveyard; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	return g, pushDiesCreatureForTest(g, me.ID, "Tersa Lightshatter", tersaOracle, "Legendary Creature — Orc Wizard", 3, 3)
}

func TestTersaAttackWithSevenCardsExilesOneAtRandomAndYouMayPlayIt(t *testing.T) {
	g, tersa := tersaGame(t, 7)
	me, opp := g.Seats[0], g.Seats[1]
	attackWithoutDamage := func() {
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(tersa, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		lockInAttacks(t, g)
		passPriorityAroundTable(t, g)
	}
	attackWithoutDamage()
	if got := graveyardSize(g, me.ID); got != 6 {
		t.Fatalf("graveyard = %d, want 6 after exiling one", got)
	}
	var exiled *game.Card
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].Name == "Dead Card" {
			exiled = &g.Exile.Cards[i]
		}
	}
	if exiled == nil {
		t.Fatal("no card was exiled")
	}
	perm := g.CastPermissionOnCardByIDForEffect(exiled.InstanceID)
	if !perm.Granted() || perm.Player != me.ID {
		t.Fatalf("permission = %+v, want a grant to Tersa's controller", perm)
	}
	if perm.CastOnly {
		t.Error("\"play\" includes lands: the grant must not be cast-only")
	}
}

func TestTersaAttackWithSixCardsDoesNothing(t *testing.T) {
	g, tersa := tersaGame(t, 6)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(tersa, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := graveyardSize(g, me.ID); got != 6 {
		t.Errorf("graveyard = %d, want 6 untouched", got)
	}
	if len(g.Exile.Cards) != 0 {
		t.Errorf("exile = %d cards, want none", len(g.Exile.Cards))
	}
}

func TestTersaHasHaste(t *testing.T) {
	g, tersa := tersaGame(t, 0)
	if !hasEffectiveKeyword(t, g, tersa, "haste") {
		t.Error("Tersa Lightshatter has haste")
	}
}

func TestTersaEntersDiscardUpToTwoThenDrawThatMany(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castCatalogSpell(t, g, "Tersa Lightshatter", "Legendary Creature — Orc Wizard", tersaOracle, nil)
	_ = id
	passPriorityAroundTable(t, g)
	// The ETB offers a discard of up to two: take a prompt, discard none.
	ch := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if ch == nil {
		t.Fatalf("no discard prompt after entering: %+v", g.PendingChoices)
	}
	hand := me.Hand.Size()
	two := append([]uuid.UUID(nil), ch.ChooseCards[:2]...)
	if err := g.ResolveChooseCards(ch.ID, me.ID, two); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want unchanged: discard two, draw two", hand, got)
	}
	if got := graveyardSize(g, me.ID); got != 2 {
		t.Errorf("graveyard = %d, want the 2 discards", got)
	}
}
