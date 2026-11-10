package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b59_test.go — Standard leftovers, batch 59 (#4452): Wilds of Eldraine.

const (
	stdB59SlumberOracle    = "48ee9007-2d1c-446f-aee3-795d23028e06"
	stdB59WitchkiteOracle  = "aee2cd1b-9bf9-4b91-83a4-2f1dbc82dcf4"
	stdB59DreamthiefOracle = "d6e68414-80cb-4e08-ae84-b084a3f4b6de"
)

func TestStdB59EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB59SlumberOracle:    "Faerie Slumber Party",
		stdB59WitchkiteOracle:  "Malevolent Witchkite",
		stdB59DreamthiefOracle: "Faerie Dreamthief",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered=%v name=%q completeness=%s", name, ok, spec.Name, spec.Completeness)
		}
	}
}

func TestStdB59DreamthiefSurveilsOneWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top A", "Top B")
	castCatalogSpell(t, g, "Faerie Dreamthief", "Creature — Faerie Warlock", stdB59DreamthiefOracle, nil)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil choice = %+v, want one card", c)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{c.ScryCards[0]}, nil); err != nil {
		t.Fatal(err)
	}
	if got := graveyardNames(me); len(got) != 1 || got[0] != "Top A" {
		t.Errorf("graveyard = %v, want [Top A]", got)
	}
}

func TestStdB59DreamthiefGraveyardAbilityDrawsAndLosesLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Drawn", "Next")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	life, hand := me.Life, me.Hand.Size()
	id := pushCatalogGraveyardCard(me, "Faerie Dreamthief", "Creature — Faerie Warlock", stdB59DreamthiefOracle, 1, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d→%d life %d→%d, want one card and 1 life", hand, me.Hand.Size(), life, me.Life)
	}
	if me.Graveyard.Contains(id) {
		t.Error("the card stayed in the graveyard")
	}
}

func TestStdB59SlumberPartyBouncesAllAndMakesFaeriesPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	b12Creature(g, o1.ID, "Theirs A", "Creature — Bear", 2, 2)
	b12Creature(g, o1.ID, "Theirs B", "Creature — Bear", 2, 2)
	_ = o2
	castCatalogSpell(t, g, "Faerie Slumber Party", "Sorcery", stdB59SlumberOracle, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(mine) {
		t.Error("my own creature stayed (the sweep is symmetric)")
	}
	// one opponent controlled creatures: two Faeries, however many they had
	if got := mshB1Tokens(g, me.ID, "Faerie"); got != 2 {
		t.Fatalf("%d Faeries, want 2 (one opponent controlled creatures)", got)
	}
	f := mshB1Token(t, g, me.ID, "Faerie")
	if f.Power != 1 || f.Toughness != 1 || !hasKeywordOnBattlefield(t, g, f.InstanceID, "flying") {
		t.Errorf("Faerie = %d/%d, want a 1/1 flier", f.Power, f.Toughness)
	}
}

func TestStdB59SlumberPartyMakesNothingWhenOnlyIControlCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Faerie Slumber Party", "Sorcery", stdB59SlumberOracle, nil)
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Faerie"); got != 0 {
		t.Errorf("%d Faeries, want 0", got)
	}
}

func TestStdB59SlumberPartyCountsEachOpponentOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	b12Creature(g, o1.ID, "A", "Creature — Bear", 2, 2)
	b12Creature(g, o2.ID, "B", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Faerie Slumber Party", "Sorcery", stdB59SlumberOracle, nil)
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Faerie"); got != 4 {
		t.Errorf("%d Faeries, want 4 (two opponents)", got)
	}
}

func TestStdB59WitchkiteSacrificesAnyNumberAndDrawsThatMany(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D")
	advanceToMain(t, g)
	art := pushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	ench := pushCatalogPermanent(g, me.ID, "Charm", "Enchantment", "", false)
	keep := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Malevolent Witchkite", "Creature — Dragon Warlock", stdB59WitchkiteOracle, nil)
	passPriorityAroundTable(t, g)

	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{art, ench}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if g.Battlefield.Contains(art) || g.Battlefield.Contains(ench) || !g.Battlefield.Contains(keep) {
		t.Error("wrong permanents sacrificed (an ordinary creature is not a legal pick)")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}
