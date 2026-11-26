package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o375Aerie     = "674c34eb-1080-4811-8f2a-5006e08d98b8"
	o375PainSeer  = "910036e7-ae27-4adb-a314-d9760bce1eb0"
	o375Felhide   = "91e3f057-c24f-42b5-8525-74b44bdefc83"
	o375Pseudamma = "92680de3-6ee2-4b73-b0f5-38383482dcca"
	o375Arbiter   = "a6cefecd-459e-439a-a62c-689c2919b689"
	o375Siren     = "abb9a164-34cd-41d7-8e89-5700fff336dd"
)

func TestO375CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		o375Aerie: "Aerie Worshippers", o375PainSeer: "Pain Seer", o375Felhide: "Felhide Spiritbinder",
		o375Pseudamma: "Forlorn Pseudamma", o375Arbiter: "Arbiter of the Ideal", o375Siren: "Siren of the Silent Song",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// o375Inspire taps the creature and untaps it through the untap verb, so the inspired
// trigger goes on the stack.
func o375Inspire(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).Tapped = true })
	if err := g.UntapAll(p.ID); err != nil {
		t.Fatalf("UntapAll: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func o375Named(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == name {
				out = append(out, c)
			}
		}
	})
	return out
}

func TestO375InspiredTokenMakers(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, token string
		colour                        string
		flying                        bool
	}{
		{"Aerie Worshippers", "Creature — Human Cleric", o375Aerie, "Bird", "U", true},
		{"Forlorn Pseudamma", "Creature — Zombie", o375Pseudamma, "Zombie", "B", false},
	} {
		for _, pay := range []bool{true, false} {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			o375Inspire(t, g, me, id)
			if !effectstest.HasPayUnlessFor(g, me.ID) {
				t.Fatalf("%s: no pay prompt for the inspired trigger", tc.name)
			}
			if pay {
				effectstest.B09AddColorless(me, 2)
				me.ManaPool.AddMana(game.ManaToken{Color: tc.colour})
			}
			effectstest.AnswerPayUnless(t, g, me.ID, pay)
			effectstest.PassPriorityAroundTable(t, g)
			toks := o375Named(g, me.ID, tc.token)
			if !pay {
				if len(toks) != 0 {
					t.Errorf("%s: a token without paying", tc.name)
				}
				continue
			}
			if len(toks) != 1 {
				t.Fatalf("%s: %d %s tokens, want 1", tc.name, len(toks), tc.token)
			}
			c := toks[0]
			if !c.IsEnchantment() || !c.IsCreature() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
				t.Errorf("%s: token %s %d/%d, want a 2/2 enchantment creature", tc.name, c.TypeLine, c.CurrentPower(), c.CurrentToughness())
			}
			if tc.flying && !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Errorf("%s: the Bird has no flying", tc.name)
			}
		}
	}
}

func TestO375PainSeerPutsTheTopCardInHandAndLosesLifeEqualToItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Pain Seer", "Creature — Human Wizard", o375PainSeer, 2, 2)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Costly Card", TypeLine: "Creature — Test", ManaCost: "{3}{R}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	life := me.Life
	o375Inspire(t, g, me, id)
	if !me.Hand.Contains(top) {
		t.Fatal("the revealed card did not go to hand")
	}
	if me.Life != life-4 {
		t.Fatalf("life %d, want %d (lost its mana value, 4)", me.Life, life-4)
	}
}

func TestO375SirenMakesEachOpponentDiscardThenMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Siren of the Silent Song", "Creature — Zombie Siren", o375Siren, 2, 1)
	type snap struct{ hand, lib int }
	before := map[uuid.UUID]snap{}
	for _, s := range g.Seats {
		if s.ID == me.ID {
			continue
		}
		for i := 0; i < 2; i++ {
			s.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Hand Card", TypeLine: "Sorcery", Owner: s.ID, Controller: s.ID})
			s.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lib Card", TypeLine: "Sorcery", Owner: s.ID, Controller: s.ID})
		}
		before[s.ID] = snap{s.Hand.Size(), s.Library.Size()}
	}
	o375Inspire(t, g, me, id)
	for _, s := range g.Seats {
		if s.ID == me.ID {
			continue
		}
		if effectstest.DiscardChoiceFor(g, s.ID) != nil {
			effectstest.DiscardFromHand(t, g, s.ID)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if s.ID == me.ID {
			continue
		}
		b := before[s.ID]
		if s.Hand.Size() != b.hand-1 || s.Library.Size() != b.lib-1 {
			t.Errorf("seat %s hand %d→%d library %d→%d, want one discarded and one milled", s.ID, b.hand, s.Hand.Size(), b.lib, s.Library.Size())
		}
	}
}

func TestO375FelhideSpiritbinderCopiesAnotherCreatureAsAHastyEnchantmentThatLeavesAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Felhide Spiritbinder", "Creature — Minotaur Shaman", o375Felhide, 3, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).Tapped = true })
	if err := g.UntapAll(me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatalf("no pay prompt: %+v", g.PendingChoices)
	}
	effectstest.B09AddColorless(me, 1)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	toks := o375Named(g, me.ID, "Their Ogre")
	if len(toks) != 1 {
		t.Fatalf("%d Ogre copies under my control, want 1", len(toks))
	}
	c := toks[0]
	if !c.IsEnchantment() || !c.IsCreature() || c.CurrentPower() != 4 {
		t.Errorf("copy is %s %d/%d, want a 4/4 enchantment creature", c.TypeLine, c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "haste") {
		t.Error("the copy has no haste")
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(c.InstanceID) {
		t.Error("the copy was not exiled at the beginning of the next end step")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the original left")
	}
}

func TestO375ArbiterPutsARevealedCreatureOntoTheBattlefieldAsAnEnchantmentWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Arbiter of the Ideal", "Creature — Sphinx", o375Arbiter, 4, 5)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).Tapped = true })
	if err := g.UntapAll(me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCardByID(g, top)
	if !ok {
		t.Fatal("the revealed creature did not enter")
	}
	if !c.IsEnchantment() || !c.IsCreature() {
		t.Errorf("entered as %s, want a creature that is also an enchantment", c.TypeLine)
	}
	if c.Counters["manifestation"] != 1 {
		t.Errorf("manifestation counters %d, want 1", c.Counters["manifestation"])
	}
}

func TestO375ArbiterLeavesADeclinedOrNonmatchingCardOnTop(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		asked    bool
	}{
		{"declined creature", "Creature — Bear", true},
		{"sorcery", "Sorcery", false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		id := effectstest.B12Push(g, me.ID, "Arbiter of the Ideal", "Creature — Sphinx", o375Arbiter, 4, 5)
		top := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Card", TypeLine: tc.typeLine, Owner: me.ID, Controller: me.ID})
		g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).Tapped = true })
		if err := g.UntapAll(me.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if tc.asked {
			effectstest.AnswerConfirmFor(t, g, me, false)
			effectstest.PassPriorityAroundTable(t, g)
		}
		if g.Battlefield.Contains(top) {
			t.Errorf("%s: the card entered", tc.name)
		}
		if !me.Library.Contains(top) {
			t.Errorf("%s: the card left the library", tc.name)
		}
	}
}
