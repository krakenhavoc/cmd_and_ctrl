package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const creationOfAvacynOracle = "7f32344f-06d8-4862-bfeb-f90a3e1d7fc2"

// avacynTable casts The Creation of Avacyn for the active seat with one
// tutor target ("Wanted") in the library, resolves chapter I's trigger and
// stops at the search prompt.
func avacynTable(t *testing.T, wanted game.Card) (g *game.Game, me *game.Player, saga, target uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	wanted.Name = "Wanted"
	wanted.Owner, wanted.Controller = me.ID, me.ID
	g.WithWriteLock(func() {
		target = pushLibraryCardForTest(me, wanted)
	})
	saga = castCatalogSpell(t, g, "The Creation of Avacyn", "Enchantment — Saga", creationOfAvacynOracle, nil)
	passPriorityAroundTable(t, g)
	return g, me, saga, target
}

func avacynExileCard(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

func avacynNextChapter(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
}

func TestCreationOfAvacynChapterISearchesAndExilesFaceDownLinked(t *testing.T) {
	g, me, saga, target := avacynTable(t, game.Card{TypeLine: "Creature — Elf", ManaCost: "{2}{G}"})
	answerSearchByID(t, g, me.ID, target)

	c, ok := avacynExileCard(g, target)
	if !ok {
		t.Fatal("the found card is not in exile")
	}
	if !c.FaceDown {
		t.Error("the found card was exiled face up")
	}
	if c.ExiledWith.ID != saga {
		t.Errorf("exiled with %v, want the Saga %v", c.ExiledWith.ID, saga)
	}
	if !c.KnownBy[me.ID] {
		t.Error("the searcher cannot look at the card they exiled")
	}
	for _, p := range g.Seats {
		if p.ID != me.ID && c.KnownBy[p.ID] {
			t.Errorf("opponent %s knows a face-down exiled card", p.ID)
		}
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("a prompt is still open after the search: %d", len(g.PendingChoices))
	}
}

func TestCreationOfAvacynChapterIIFlipsAndChargesCreatureManaValue(t *testing.T) {
	g, me, _, target := avacynTable(t, game.Card{TypeLine: "Creature — Elf", ManaCost: "{2}{G}"})
	answerSearchByID(t, g, me.ID, target)
	life := me.Life

	avacynNextChapter(t, g, me)

	c, ok := avacynExileCard(g, target)
	if !ok {
		t.Fatal("the card left exile during chapter II")
	}
	if c.FaceDown {
		t.Error("chapter II left the card face down")
	}
	for _, p := range g.Seats {
		if !c.KnownBy[p.ID] {
			t.Errorf("seat %s cannot read the card chapter II turned face up", p.ID)
		}
	}
	if got := me.Life; got != life-3 {
		t.Errorf("life = %d, want %d (lost the creature's mana value 3)", got, life-3)
	}
}

func TestCreationOfAvacynChapterIINonCreatureCostsNothing(t *testing.T) {
	g, me, _, target := avacynTable(t, game.Card{TypeLine: "Sorcery", ManaCost: "{4}{B}"})
	answerSearchByID(t, g, me.ID, target)
	life := me.Life

	avacynNextChapter(t, g, me)

	if c, _ := avacynExileCard(g, target); c.FaceDown {
		t.Error("chapter II left a non-creature card face down")
	}
	if me.Life != life {
		t.Errorf("life = %d, want %d: only a creature card costs life", me.Life, life)
	}
}

func TestCreationOfAvacynChapterIIIAcceptPutsCreatureOntoBattlefield(t *testing.T) {
	g, me, saga, target := avacynTable(t, game.Card{TypeLine: "Creature — Elf", ManaCost: "{2}{G}", Power: 2, Toughness: 2})
	answerSearchByID(t, g, me.ID, target)
	avacynNextChapter(t, g, me)
	avacynNextChapter(t, g, me)

	c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if c == nil {
		t.Fatal("chapter III asked nothing about a creature card")
	}
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)

	found := false
	for _, bc := range g.Battlefield.Cards {
		if bc.Name == "Wanted" && bc.Controller == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("accepting did not put the creature onto the battlefield under the player's control")
	}
	if me.Hand.Contains(target) {
		t.Error("the card is in hand as well as on the battlefield")
	}
	if loreCountersOn(g, saga) >= 0 {
		t.Error("the Saga survived chapter III")
	}
}

func TestCreationOfAvacynChapterIIIDeclineSendsCreatureToHand(t *testing.T) {
	g, me, _, target := avacynTable(t, game.Card{TypeLine: "Creature — Elf", ManaCost: "{2}{G}"})
	answerSearchByID(t, g, me.ID, target)
	avacynNextChapter(t, g, me)
	avacynNextChapter(t, g, me)

	c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if c == nil {
		t.Fatal("chapter III asked nothing about a creature card")
	}
	if err := g.ResolveConfirm(c.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if !me.Hand.Contains(target) {
		t.Error("declining did not put the card into its owner's hand")
	}
	if _, still := avacynExileCard(g, target); still {
		t.Error("the card stayed in exile")
	}
}

func TestCreationOfAvacynChapterIIINonCreatureGoesStraightToHand(t *testing.T) {
	g, me, _, target := avacynTable(t, game.Card{TypeLine: "Artifact", ManaCost: "{3}"})
	answerSearchByID(t, g, me.ID, target)
	avacynNextChapter(t, g, me)
	avacynNextChapter(t, g, me)

	if c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID); c != nil {
		t.Fatal("chapter III offered the battlefield to a non-creature card")
	}
	if !me.Hand.Contains(target) {
		t.Error("a non-creature exiled card did not go to its owner's hand")
	}
}

// A Saga that left and came back is a new object (CR 400.7) with no claim
// on the card the old one exiled: its chapters find nothing to flip or move.
func TestCreationOfAvacynIgnoresACardTheSagaDidNotExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	stranger := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: stranger, Name: "Stranger", TypeLine: "Creature — Elf",
			ManaCost: "{4}", Owner: me.ID, Controller: me.ID, FaceDown: true})
	})
	castCatalogSpell(t, g, "The Creation of Avacyn", "Enchantment — Saga", creationOfAvacynOracle, nil)
	passPriorityAroundTable(t, g)
	// An unconditional search cannot come up empty, so the Saga takes a
	// library card; the stranger already in exile is not it.
	c0 := searchChoiceFor(g, me.ID)
	if c0 == nil {
		t.Fatal("chapter I opened no search")
	}
	answerSearchByID(t, g, me.ID, c0.SearchCards[0])
	life := me.Life
	avacynNextChapter(t, g, me)
	avacynNextChapter(t, g, me)

	c, ok := avacynExileCard(g, stranger)
	if !ok || !c.FaceDown {
		t.Error("the Saga touched a card it did not exile")
	}
	if me.Life != life {
		t.Errorf("life = %d, want %d", me.Life, life)
	}
}

func TestCreationOfAvacynHasNoCaveats(t *testing.T) {
	spec, ok := Lookup(creationOfAvacynOracle)
	if !ok {
		t.Fatal("The Creation of Avacyn is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness = %v caveats = %v, want full and none", spec.Completeness, spec.Caveats)
	}
}
