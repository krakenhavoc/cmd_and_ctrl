package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b19_cards_test.go — Ravnica: Clue Edition, S82 BNS batch 19 (#5929).
// Each test plays the real catalog entries through the engine.

const (
	bnsB19Insight   = "a2cb30d7-c0a9-4f79-bc19-fc88c4986725"
	bnsB19Code      = "b2f83483-24a4-45db-a04e-42ff732ca98a"
	bnsB19Assault   = "ba14843d-6b7f-4e3d-b5ba-fc573fa8bf50"
	bnsB19Churn     = "e2aeac0b-af74-4c47-9ac1-51fbf869c95f"
	bnsB19Gift      = "fe971240-268a-47d4-a14e-3d6352656091"
	bnsB19Wizardry  = "f621dcaa-c500-4a17-a65a-60d9b5516a3b"
	bnsB19Gods      = "48e603a2-b965-4fbc-ad57-4388bce5ac8b"
	bnsB19Leap      = "593caa15-e1d1-477f-bbb4-124d7a7b81f3"
	bnsB19Pit       = "ef4b0ced-70a1-41d8-9aed-34e75f979c5f"
	bnsB19Rally     = "9d8d88dd-c266-4b7d-a756-ac2d8976ca89"
	bnsB19Repeal    = "b1f2b26c-91b2-4f3c-9f48-6429ec3fde48"
	bnsB19Snakeform = "76ae07ce-ca0c-4308-b8bd-d999eea0a65f"
)

func TestBnsB19EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB19Insight: "Chemister's Insight", bnsB19Code: "Code of Constraint",
		bnsB19Assault: "Coordinated Assault", bnsB19Churn: "Corpse Churn",
		bnsB19Gift: "Gift of Strength", bnsB19Wizardry: "Goblin Wizardry",
		bnsB19Gods: "Gods Willing", bnsB19Leap: "Mighty Leap", bnsB19Pit: "Pit Fight",
		bnsB19Rally: "Rally to Battle", bnsB19Repeal: "Repeal", bnsB19Snakeform: "Snakeform",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestBnsB19PumpsGiveTheirBoostAndKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		keyword      string
	}{
		{"Gift of Strength", bnsB19Gift, 3, 3, "reach"},
		{"Mighty Leap", bnsB19Leap, 2, 2, "flying"},
	} {
		g := newCatalogGame(t)
		bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
		castCatalogSpell(t, g, tc.name, "Instant", tc.oracle, cardTarget(bear))
		passPriorityAroundTable(t, g)
		if p, th := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2+tc.power || th != 2+tc.tough {
			t.Errorf("%s: %d/%d", tc.name, p, th)
		}
		if !hasKeywordOnBattlefield(t, g, bear, tc.keyword) {
			t.Errorf("%s: no %s", tc.name, tc.keyword)
		}
	}
}

func TestBnsB19CoordinatedAssaultBoostsUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	a := b12Creature(g, g.Seats[0].ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, g.Seats[1].ID, "B", "Creature — Bear", 3, 3)
	c := b12Creature(g, g.Seats[1].ID, "C", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Coordinated Assault", "Instant", bnsB19Assault, cardTargets(a, b))
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{a: 3, b: 4, c: 3} {
		if got := effectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
	if !hasKeywordOnBattlefield(t, g, a, "first strike") || !hasKeywordOnBattlefield(t, g, b, "first strike") ||
		hasKeywordOnBattlefield(t, g, c, "first strike") {
		t.Error("first strike should be on exactly the two targets")
	}
	if err := castCatalogSpellErr(t, g, "Coordinated Assault", "Instant", bnsB19Assault, cardTargets(a, b, c)); err == nil {
		t.Error("three targets were accepted")
	}
}

func TestBnsB19RallyToBattleBoostsAndUntapsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].Tapped = true
		}
	})
	castCatalogSpell(t, g, "Rally to Battle", "Instant", bnsB19Rally, nil)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, mine); c.Tapped {
		t.Error("my creature is still tapped")
	}
	if c, _ := battlefieldCard(g, theirs); !c.Tapped {
		t.Error("an opponent's creature was untapped")
	}
	if p, th := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 3 || th != 5 {
		t.Errorf("mine is %d/%d, want 3/5", p, th)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("theirs is %d power, want 2", p)
	}
}

func TestBnsB19PitFightFightsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Pit Fight", "Instant", bnsB19Pit, cardTargets(mine, mine)); err == nil {
		t.Error("a creature was allowed to fight itself")
	}
	castCatalogSpell(t, g, "Pit Fight", "Instant", bnsB19Pit, cardTargets(mine, theirs))
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("the 2/2 survived 3 damage")
	}
	if c, ok := battlefieldCard(g, mine); !ok || c.DamageMarked != 2 {
		t.Errorf("mine: on battlefield %v, damage %d, want 2", ok, c.DamageMarked)
	}
}

func TestBnsB19GoblinWizardryMakesTwoProwessWizards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Goblin Wizardry", "Instant", bnsB19Wizardry, nil)
	passPriorityAroundTable(t, g)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Goblin Wizard" {
				n++
				if !c.IsToken() || c.Power != 1 || c.Toughness != 1 || !game.HasKeyword(&c, "prowess") {
					t.Errorf("token %+v is not a 1/1 prowess token", c)
				}
			}
		}
	})
	if n != 2 {
		t.Errorf("%d Goblin Wizards, want 2", n)
	}
}

func TestBnsB19CorpseChurnMillsThreeThenReturnsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
			Owner: me.ID, Controller: me.ID})
	})
	lib := me.Library.Size()
	castCatalogSpell(t, g, "Corpse Churn", "Instant", bnsB19Churn, nil)
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != lib-3 {
		t.Errorf("library %d, want %d", got, lib-3)
	}
	answerChooseCards(t, g, me.ID, bear)
	if !inHandOf(me, bear) {
		t.Error("the creature card did not return to hand")
	}
}

func TestBnsB19DrawSpellsDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Snakeform", "Instant", bnsB19Snakeform, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 { // the Snakeform left the hand, one card drawn
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestBnsB19SnakeformMakesAOneOneGreenSnakeWithoutAbilities(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	dragon := b12Creature(g, opp.ID, "Dragon", "Creature — Dragon", 5, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == dragon {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
				g.Battlefield.Cards[i].Colors = []string{"R"}
			}
		}
	})
	castCatalogSpell(t, g, "Snakeform", "Instant", bnsB19Snakeform, cardTarget(dragon))
	passPriorityAroundTable(t, g)
	var c game.Characteristic
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	g.ReadSnapshot(func() {
		for _, bc := range g.Battlefield.Cards {
			if bc.InstanceID == dragon {
				c = bc.Effective()
			}
		}
	})
	if c.Power != 1 || c.Toughness != 1 {
		t.Errorf("%d/%d, want 1/1", c.Power, c.Toughness)
	}
	if len(c.Colors) != 1 || c.Colors[0] != "G" {
		t.Errorf("colors %v, want [G]", c.Colors)
	}
	if hasKeywordOnBattlefield(t, g, dragon, "flying") {
		t.Error("it kept flying")
	}
	if !slices.Contains(c.Subtypes, "Snake") || slices.Contains(c.Subtypes, "Dragon") {
		t.Errorf("subtypes %v is not a Snake", c.Subtypes)
	}
}

func TestBnsB19RepealBouncesTheNonlandPermanentWithManaValueX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	three := b12Creature(g, opp.ID, "Three", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == three {
				g.Battlefield.Cards[i].ManaCost = "{2}{G}"
			}
			if g.Battlefield.Cards[i].InstanceID == other {
				g.Battlefield.Cards[i].ManaCost = "{1}"
			}
		}
	})
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Repeal", TypeLine: "Instant", ManaCost: "{X}{U}",
			OracleID: bnsB19Repeal, Owner: me.ID, Controller: me.ID})
	})
	s39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3, Targets: cardTarget(other)}); err == nil {
		t.Error("a mana value 1 permanent was a legal target for X=3")
	}
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3, Targets: cardTarget(three)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, three); ok {
		t.Error("the target is still on the battlefield")
	}
	if !inHandOf(opp, three) {
		t.Error("the target did not return to its owner's hand")
	}
	if got := me.Hand.Size(); got != hand { // Repeal left the hand, one card drawn
		t.Errorf("hand %d, want %d", got, hand)
	}
}

func TestBnsB19CodeOfConstraintAddendumTapsAndFreezesInYourMainPhase(t *testing.T) {
	for _, main := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
		if main {
			castCatalogSpell(t, g, "Code of Constraint", "Instant", bnsB19Code, cardTarget(bear))
		} else {
			advanceTo(t, g, game.StepBeginCombat)
			id := uuid.New()
			g.WithWriteLock(func() {
				me.Hand.PushTop(game.Card{InstanceID: id, Name: "Code of Constraint", TypeLine: "Instant",
					OracleID: bnsB19Code, Owner: me.ID, Controller: me.ID})
			})
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardTarget(bear)}); err != nil {
				t.Fatalf("cast: %v", err)
			}
		}
		hand := me.Hand.Size()
		passPriorityAroundTable(t, g)
		if got := effectivePower(t, g, bear); got != -1 {
			t.Errorf("main=%v: power %d, want -1", main, got)
		}
		if c, _ := battlefieldCard(g, bear); c.Tapped != main {
			t.Errorf("main=%v: tapped %v", main, c.Tapped)
		}
		if got := me.Hand.Size(); got != hand+1 {
			t.Errorf("main=%v: hand %d, want %d", main, got, hand+1)
		}
	}
}

func TestBnsB19GodsWillingChoosesAColorAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Gods Willing", "Instant", bnsB19Gods, cardTarget(b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2))); err == nil {
		t.Error("an opponent's creature was a legal target")
	}
	castCatalogSpell(t, g, "Gods Willing", "Instant", bnsB19Gods, cardTarget(bear))
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "R")
	answerScryKeepAll(t, g, me.ID)
	if !hasKeywordOnBattlefield(t, g, bear, "protection from red") {
		t.Error("the Bear did not gain protection from red")
	}
}

func TestBnsB19ChemistersInsightJumpStartsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedGraveyardCard(t, g, "Chemister's Insight", "Instant", bnsB19Insight)
	fodder := handCard(me, "Fodder", "Sorcery")
	s39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "jump-start"}); err == nil {
		t.Fatal("jump-start was castable without discarding a card")
	}
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "jump-start", AltCostIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("jump-start cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand-1+2 {
		t.Errorf("hand %d, want %d (one discarded, two drawn)", got, hand+1)
	}
	if !g.Exile.Contains(id) {
		t.Error("the jump-started instant is not exiled")
	}
}
