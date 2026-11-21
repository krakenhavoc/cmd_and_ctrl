package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestValigarmandaIsRegisteredAndWhole(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.ValigarmandaOracle)
	if !ok || spec.Name != "Summon: Esper Valigarmanda" {
		t.Fatalf("not registered: %+v", spec)
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

func exiledWithNamed(g *game.Game, saga uuid.UUID) map[string]bool {
	out := map[string]bool{}
	for _, c := range g.Exile.Cards {
		if c.ExiledWith.ID == saga {
			out[c.Name] = true
		}
	}
	return out
}

// Chapter I: one instant or sorcery per graveyard, chosen by the Saga's
// controller where there is a real choice, none from a graveyard that
// has none, and every card linked to the Saga.
func TestValigarmandaChapterIExilesOneInstantOrSorceryFromEachGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp1, opp2, opp3 := g.Seats[(seat+1)%4], g.Seats[(seat+2)%4], g.Seats[(seat+3)%4]
	effectstest.ValigarmandaGraveCard(me, "My Bolt", "Instant", "{R}")
	mySorcery := effectstest.ValigarmandaGraveCard(me, "My Divination", "Sorcery", "{2}{U}")
	effectstest.ValigarmandaGraveCard(opp1, "Opp1 Only Spell", "Instant", "{U}")
	effectstest.ValigarmandaGraveCard(opp2, "Opp2 Bear", "Creature — Bear", "{1}{G}")
	effectstest.ValigarmandaGraveCard(opp3, "Opp3 First", "Instant", "{B}")
	opp3Second := effectstest.ValigarmandaGraveCard(opp3, "Opp3 Second", "Sorcery", "{1}{B}")

	saga := effectstest.ValigarmandaCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	// Two real choices (mine and the third opponent's), both made by me.
	picks := 0
	for p := effectstest.ChooseCardsChoiceFor(g, me.ID); p != nil; p = effectstest.ChooseCardsChoiceFor(g, me.ID) {
		var want uuid.UUID
		switch p.FromPlayer {
		case me.ID:
			want = mySorcery
		case opp3.ID:
			want = opp3Second
		default:
			t.Fatalf("asked about graveyard of %v, which has no choice to make", p.FromPlayer)
		}
		effectstest.AnswerChooseCards(t, g, me.ID, want)
		picks++
	}
	if picks != 2 {
		t.Fatalf("%d choices, want 2 (mine and the third opponent's)", picks)
	}
	effectstest.PassPriorityAroundTable(t, g)

	got := exiledWithNamed(g, saga)
	for _, name := range []string{"My Divination", "Opp1 Only Spell", "Opp3 Second"} {
		if !got[name] {
			t.Errorf("%s is not exiled with the Saga: %v", name, got)
		}
	}
	if len(got) != 3 {
		t.Errorf("exiled with the Saga = %v, want exactly three", got)
	}
	for _, kept := range []struct {
		p    *game.Player
		name string
	}{{me, "My Bolt"}, {opp2, "Opp2 Bear"}, {opp3, "Opp3 First"}} {
		found := false
		for _, c := range kept.p.Graveyard.Cards {
			if c.Name == kept.name {
				found = true
			}
		}
		if !found {
			t.Errorf("%s left the graveyard", kept.name)
		}
	}
}

// Chapter II adds {R} per lore counter and offers a cast of a card exiled
// with the Saga, with mana of any type: two red pay for {U}{U}.
func TestValigarmandaChapterIIAddsManaAndCastsWithAnyTypeOfMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.ValigarmandaGraveCard(me, "Blue Spell", "Instant", "{U}{U}")
	saga := effectstest.ValigarmandaCast(t, g)
	effectstest.PassPriorityAroundTable(t, g) // chapter I: the lone card is taken without a question
	if got := exiledWithNamed(g, saga); !got["Blue Spell"] {
		t.Fatalf("chapter I did not exile the spell: %v", got)
	}
	var spell uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.Name == "Blue Spell" {
			spell = c.InstanceID
		}
	}

	// Without the chapter the card is not castable.
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("cast an exiled card with no permission")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // chapter II
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool after chapter II = %d mana, want 2 ({R} per lore counter)", got)
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast {U}{U} with two red mana: %v", err)
	}
	if g.Exile.Contains(spell) {
		t.Error("the spell is still in exile after the cast")
	}
}

// The grant is one cast, and a card not cast waits for the next chapter,
// which adds one more {R} than the last.
func TestValigarmandaOffersOneCastPerChapterAndKeepsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.ValigarmandaGraveCard(me, "Spell A", "Instant", "{R}")
	effectstest.ValigarmandaGraveCard(g.Seats[(seat+1)%4], "Spell B", "Sorcery", "{R}")
	saga := effectstest.ValigarmandaCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	var a, b uuid.UUID
	for _, c := range g.Exile.Cards {
		switch c.Name {
		case "Spell A":
			a = c.InstanceID
		case "Spell B":
			b = c.InstanceID
		}
	}
	if len(exiledWithNamed(g, saga)) != 2 {
		t.Fatalf("exiled with the Saga = %v", exiledWithNamed(g, saga))
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // chapter II: two {R}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, a, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("first cast: %v", err)
	}
	if err := g.CastSpell(me.ID, b, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("a second card was castable from the same chapter")
	}
	effectstest.PassPriorityAroundTable(t, g) // resolve the spell and close the window

	// Next turn of mine: chapter III, three {R}, and B is still on offer.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("pool after chapter III = %d, want 3", got)
	}
	if err := g.CastSpell(me.ID, b, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Errorf("the card left in exile was not offered again: %v", err)
	}
}

// "Exiled with this Saga" is this Saga: a card another source exiled
// is not on offer.
func TestValigarmandaOnlyOffersCardsExiledWithIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	stray := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: stray, Name: "Stray", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	effectstest.ValigarmandaCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, stray, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast a card the Saga did not exile")
	}
}
