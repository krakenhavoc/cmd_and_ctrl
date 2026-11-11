package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s35_cards_test.go — batch 35 (Commander deck morph cards, #5368).

const (
	s35ColossusOracle     = "703ab192-7e0a-4f02-8381-568ad2d999ca"
	s35CloudscraperOracle = "a0ff742b-f709-43ac-84d2-dc00c723bad6"
	s35CavernOracle       = "3763de30-28e1-4689-a71c-07d2fea3a466"
	s35PearlsOracle       = "aca9df28-2df6-477c-b5b8-8e1863c84356"
	s35AngelOracle        = "2f6e9ab8-6951-4c5f-bcde-0e96e0fe87f2"
	s35HermitOracle       = "69d2e5f3-2f4d-4a36-bcd5-181164ecd894"
	s35NantukoOracle      = "b0a54050-3493-4ff6-8eca-e43b92c1f3d3"
	s35HellkiteOracle     = "99642419-ab24-4770-9f1c-84d2806441b5"
	s35RootOracle         = "182d38b4-48bb-4050-a19a-c0a6bb241bbd"
	s35VigilanteOracle    = "28ea78dc-7ad6-4006-aaf8-4b25057ec019"
)

func TestS35MorphCostsAreWired(t *testing.T) {
	for _, tc := range []struct{ name, oracle, faceUp string }{
		{"Krosan Colossus", s35ColossusOracle, "{6}{G}{G}"},
		{"Krosan Cloudscraper", s35CloudscraperOracle, "{7}{G}{G}"},
		{"Zoetic Cavern", s35CavernOracle, "{2}"},
		{"Master of Pearls", s35PearlsOracle, "{3}{W}{W}"},
		{"Exalted Angel", s35AngelOracle, "{2}{W}{W}"},
		{"Thelonite Hermit", s35HermitOracle, "{3}{G}{G}"},
		{"Broodhatch Nantuko", s35NantukoOracle, "{2}{G}"},
		{"Imperial Hellkite", s35HellkiteOracle, "{6}{R}{R}"},
		{"Root Elemental", s35RootOracle, "{5}{G}{G}"},
		{"Nantuko Vigilante", s35VigilanteOracle, "{1}{G}"},
	} {
		ac := game.AlternativeCostByKey(tc.oracle, "morph")
		if ac == nil || ac.FaceDown == nil {
			t.Errorf("%s offers no morph cast", tc.name)
			continue
		}
		if ac.ManaCost != MorphCastCost || ac.FaceDown.Kind != game.FaceDownMorphed || ac.FaceDown.FaceUpCost != tc.faceUp {
			t.Errorf("%s: cast %q, kind %q, face-up %q; want %q, morphed, %q",
				tc.name, ac.ManaCost, ac.FaceDown.Kind, ac.FaceDown.FaceUpCost, MorphCastCost, tc.faceUp)
		}
	}
}

// s35CastFaceDown casts a card face down for its morph and resolves it.
func s35CastFaceDown(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := castWithAltCost(t, g, name, typeLine, oracle, "morph")
	passPriorityAroundTable(t, g)
	if c := findCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatalf("%s did not enter face down", name)
	}
	return id
}

func s35TurnUp(t *testing.T, g *game.Game, id uuid.UUID, mana ...string) {
	t.Helper()
	p := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { p.ManaPool.AddMana(manaTokens(mana...)...) })
	if err := g.PerformSpecialAction(p.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	if findCardAnywhere(t, g, id).FaceDown {
		t.Fatal("still face down after the special action")
	}
}

func TestS35ZoeticCavernCastsFaceDownAndTurnsBackIntoALand(t *testing.T) {
	g := newCatalogGame(t)
	id := s35CastFaceDown(t, g, "Zoetic Cavern", "Land", s35CavernOracle)
	if c := findCardAnywhere(t, g, id); !c.IsCreature() || c.IsLand() {
		t.Fatalf("face-down Cavern should be a creature, not a land: %q", c.TypeLine)
	}
	s35TurnUp(t, g, id, "C", "C")
	if c := findCardAnywhere(t, g, id); !c.IsLand() || c.Name != "Zoetic Cavern" {
		t.Fatalf("turned up: %q %q, want the land back", c.Name, c.TypeLine)
	}
}

func TestS35MasterOfPearlsPumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	id := s35CastFaceDown(t, g, "Master of Pearls", "Creature — Human Monk", s35PearlsOracle)
	before := effectivePower(t, g, bear)
	s35TurnUp(t, g, id, "W", "W", "C", "C", "C")
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != before+2 {
		t.Fatalf("bear power %d, want %d", got, before+2)
	}
}

func TestS35TheloniteHermitMakesFourBuffedSaprolings(t *testing.T) {
	g := newCatalogGame(t)
	id := s35CastFaceDown(t, g, "Thelonite Hermit", "Creature — Elf Shaman", s35HermitOracle)
	s35TurnUp(t, g, id, "G", "G", "C", "C", "C")
	passPriorityAroundTable(t, g)
	saps := battlefieldIDsNamed(g, "Saproling")
	if len(saps) != 4 {
		t.Fatalf("saprolings = %d, want 4", len(saps))
	}
	for _, s := range saps {
		if p := effectivePower(t, g, s); p != 2 {
			t.Errorf("saproling power %d, want 2 under the Hermit", p)
		}
	}
}

func TestS35ExaltedAngelGainsLifeWhenItDealsDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushCatalogPermanent(g, me.ID, "Exalted Angel", "Creature — Angel", s35AngelOracle, false)
	life := me.Life
	pr7Hit(t, g, id, opp.ID, 4)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf("life %d, want %d", me.Life, life+4)
	}
}

func TestS35BroodhatchNantukoMakesThatManyInsects(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	id := pushCatalogPermanent(g, me.ID, "Broodhatch Nantuko", "Creature — Insect Druid", s35NantukoOracle, false)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, id, 3) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	s35AnswerYes(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Insect")); n != 3 {
		t.Fatalf("insects = %d, want 3", n)
	}
}

func TestS35ImperialHellkiteFetchesADragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dragon := pushLibraryCardForTest(me, game.Card{Name: "Test Dragon", TypeLine: "Creature — Dragon"})
	pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	id := s35CastFaceDown(t, g, "Imperial Hellkite", "Creature — Dragon", s35HellkiteOracle)
	s35TurnUp(t, g, id, "R", "R", "C", "C", "C", "C", "C", "C")
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, dragon)
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == dragon {
			found = true
		}
	}
	if !found {
		t.Fatal("the Dragon did not reach the hand")
	}
}

func TestS35RootElementalPutsACreatureFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := s35CastFaceDown(t, g, "Root Elemental", "Creature — Elemental", s35RootOracle)
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Hand Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	s35TurnUp(t, g, id, "G", "G", "C", "C", "C", "C", "C")
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, bear); !ok {
		t.Fatal("the creature card was not put onto the battlefield")
	}
}

func TestS35NantukoVigilanteDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := pushCatalogPermanent(g, opp.ID, "Test Rock", "Artifact", "", false)
	id := s35CastFaceDown(t, g, "Nantuko Vigilante", "Creature — Insect Druid Mutant", s35VigilanteOracle)
	s35TurnUp(t, g, id, "G", "C")
	_ = me
	answerFirstPickTarget(t, g)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, rock); ok {
		t.Fatal("the artifact survived")
	}
}

// s35AnswerYes says yes to the open "you may" prompt, whichever kind
// the engine raised (a trigger prompt, or a confirm from the effect).
func s35AnswerYes(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c == nil || c.Chooser != chooser {
			continue
		}
		switch c.Kind {
		case game.PendingChoiceTriggerPrompt:
			if err := g.ResolveTriggerPrompt(c.ID, chooser, true); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			return
		case game.PendingChoiceConfirm:
			if err := g.ResolveConfirm(c.ID, chooser, true); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			return
		}
	}
	for _, c := range g.PendingChoices {
		t.Logf("pending kind=%s chooser=%s reason=%q", c.Kind, c.Chooser, c.Reason)
	}
	t.Fatalf("no yes/no prompt for %s", chooser)
}
