package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82OneAdaptiveSporesingerOracle = "f6311b7a-7482-4b13-9f8f-13e80c98e98e"
	s82OneBlightbellyRatOracle      = "b0102259-7a2c-470e-b04b-4a1f615d057c"
	s82OneCacophonyScampOracle      = "4854a41e-db49-462f-ae57-0ba4005bae75"
	s82OneContagiousVorracOracle    = "1b6da140-0d91-4fae-a3ee-33e6808acf34"
	s82OneCopperLonglegsOracle      = "066f6a3d-5e7b-40a4-b348-6dfb5d5239de"
	s82OneGitaxianAnatomistOracle   = "aa9bde06-098c-41b3-b816-49daacc02c50"
	s82OneGulpingScraptrapOracle    = "d9e48750-d930-4095-abf0-dec2ac53be81"
	s82OneTaintedObserverOracle     = "420aaa77-f429-4f3d-a73e-91dc25cb38db"
	s82OneVenomousBrutalizerOracle  = "766c3528-2c5c-4e92-aa85-b1e0e115d491"
	s82OneMesmerizingDoseOracle     = "195ab0c7-0848-4803-a299-9155e1f56a85"
)

func TestS82OneB06EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82OneAdaptiveSporesingerOracle: "Adaptive Sporesinger",
		s82OneBlightbellyRatOracle:      "Blightbelly Rat",
		s82OneCacophonyScampOracle:      "Cacophony Scamp",
		s82OneContagiousVorracOracle:    "Contagious Vorrac",
		s82OneCopperLonglegsOracle:      "Copper Longlegs",
		s82OneGitaxianAnatomistOracle:   "Gitaxian Anatomist",
		s82OneGulpingScraptrapOracle:    "Gulping Scraptrap",
		s82OneTaintedObserverOracle:     "Tainted Observer",
		s82OneVenomousBrutalizerOracle:  "Venomous Brutalizer",
		s82OneMesmerizingDoseOracle:     "Mesmerizing Dose",
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

// s82OneGivePoison puts n poison counters on p.
func s82OneGivePoison(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(p.ID, game.CounterPoison, n) })
}

func TestS82OneB06BlightbellyRatProliferatesWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := b12Push(g, me.ID, "Blightbelly Rat", "Creature — Phyrexian Rat", s82OneBlightbellyRatOracle, 2, 2)
	s82OneGivePoison(g, opp, 1)
	if !hasEffectiveKeyword(t, g, rat, "toxic 1") {
		t.Error("the Rat has no toxic 1")
	}
	killCreature(t, g, me.ID, rat)
	settleAnsweringProliferate(t, g)
	if poisonOn(opp) != 2 {
		t.Errorf("poison %d, want 2 after the dies proliferate", poisonOn(opp))
	}
}

func TestS82OneB06GulpingScraptrapProliferatesOnEnterAndOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s82OneGivePoison(g, opp, 1)
	trap := castRingCreature(t, g, "Gulping Scraptrap", "Creature — Phyrexian Horror", s82OneGulpingScraptrapOracle, 4, 4)
	settleAnsweringProliferate(t, g)
	if poisonOn(opp) != 2 {
		t.Fatalf("after entering: poison %d, want 2", poisonOn(opp))
	}
	killCreature(t, g, me.ID, trap)
	settleAnsweringProliferate(t, g)
	if poisonOn(opp) != 3 {
		t.Errorf("after dying: poison %d, want 3", poisonOn(opp))
	}
}

func TestS82OneB06CopperLonglegsSacrificesToProliferate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legs := b12Push(g, me.ID, "Copper Longlegs", "Creature — Phyrexian Spider", s82OneCopperLonglegsOracle, 1, 3)
	s82OneGivePoison(g, opp, 1)
	if err := g.ActivateCatalogAbility(me.ID, legs, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana to pay {1}{G}")
	}
	if !g.Battlefield.Contains(legs) {
		t.Fatal("the failed activation sacrificed the Longlegs")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, legs, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(legs) {
		t.Error("the Longlegs was not sacrificed as a cost")
	}
	settleAnsweringProliferate(t, g)
	if poisonOn(opp) != 2 {
		t.Errorf("poison %d, want 2", poisonOn(opp))
	}
}

func TestS82OneB06VenomousBrutalizerProliferatesOnlyIfYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		brut := castRingCreature(t, g, "Venomous Brutalizer", "Creature — Phyrexian Knight", s82OneVenomousBrutalizerOracle, 4, 4)
		passPriorityAroundTable(t, g)
		me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"})
		answerPayUnless(t, g, me.ID, pay)
		settleAnsweringProliferate(t, g)
		want := 1
		if pay {
			want = 2
		}
		if poisonOn(opp) != want {
			t.Errorf("pay=%v: poison %d, want %d", pay, poisonOn(opp), want)
		}
		if !hasEffectiveKeyword(t, g, brut, "toxic 3") {
			t.Error("no toxic 3")
		}
	}
}

func TestS82OneB06TaintedObserverProliferatesForTwoWhenAnotherCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tainted Observer", "Creature — Phyrexian Bird", s82OneTaintedObserverOracle, 2, 3)
	s82OneGivePoison(g, opp, 1)
	castRingCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	answerPayUnless(t, g, me.ID, true)
	settleAnsweringProliferate(t, g)
	if poisonOn(opp) != 2 {
		t.Errorf("poison %d, want 2", poisonOn(opp))
	}
	// The Observer's own entry is not "another creature": declining or
	// nothing to pay asks nothing.
	if hasPayUnlessFor(g, me.ID) {
		t.Error("a payment prompt is still open")
	}
}

func TestS82OneB06TaintedObserverDeclinedPaymentDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tainted Observer", "Creature — Phyrexian Bird", s82OneTaintedObserverOracle, 2, 3)
	s82OneGivePoison(g, opp, 1)
	castRingCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		answerPayUnless(t, g, me.ID, false)
	}
	settleAnsweringProliferate(t, g)
	if poisonOn(opp) != 1 {
		t.Errorf("poison %d, want 1: nothing was paid", poisonOn(opp))
	}
}

func TestS82OneB06GitaxianAnatomistTapsItselfToProliferate(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		anat := castRingCreature(t, g, "Gitaxian Anatomist", "Creature — Phyrexian Wizard", s82OneGitaxianAnatomistOracle, 2, 5)
		passPriorityAroundTable(t, g)
		answerMayChoice(t, g, me.ID, yes)
		settleAnsweringProliferate(t, g)
		c := findBattlefieldCardByID(g, anat)
		if c == nil {
			t.Fatal("the Anatomist is gone")
		}
		want := 1
		if yes {
			want = 2
		}
		if c.Tapped != yes || poisonOn(opp) != want {
			t.Errorf("yes=%v: tapped %v, poison %d; want tapped %v, poison %d", yes, c.Tapped, poisonOn(opp), yes, want)
		}
	}
}

func TestS82OneB06MesmerizingDoseTapsThenProliferatesAndKeepsItDown(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	s82OneGivePoison(g, opp, 1)
	castCatalogSpell(t, g, "Mesmerizing Dose", auraTypeLine, s82OneMesmerizingDoseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	settleAnsweringProliferate(t, g)
	if !tappedForTest(t, g, bear) {
		t.Error("the enchanted creature was not tapped")
	}
	if poisonOn(opp) != 2 {
		t.Errorf("poison %d, want 2", poisonOn(opp))
	}
	advanceToUpkeepOf(t, g, 0)
	if !tappedForTest(t, g, bear) {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

func TestS82OneB06AdaptiveSporesingerModes(t *testing.T) {
	t.Run("pump", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		castRingCreature(t, g, "Adaptive Sporesinger", "Creature — Phyrexian Druid", s82OneAdaptiveSporesingerOracle, 2, 2)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		pickCard(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
		if !hasEffectiveKeyword(t, g, bear, "vigilance") {
			t.Error("the target did not gain vigilance")
		}
		var p, tough int
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == bear {
					e := c.Effective()
					p, tough = e.Power, e.Toughness
				}
			}
		})
		if p != 4 || tough != 4 {
			t.Errorf("target is %d/%d, want 4/4", p, tough)
		}
	})
	t.Run("proliferate", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		castRingCreature(t, g, "Adaptive Sporesinger", "Creature — Phyrexian Druid", s82OneAdaptiveSporesingerOracle, 2, 2)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		settleAnsweringProliferate(t, g)
		if poisonOn(opp) != 2 {
			t.Errorf("poison %d, want 2", poisonOn(opp))
		}
	})
}

func TestS82OneB06ContagiousVorracTakesALandOrProliferates(t *testing.T) {
	t.Run("land taken", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		seedLibrary(me, "A", "B", "C", "D", "E")
		land := topLibraryCard(me, "Forest", "Basic Land — Forest")
		castRingCreature(t, g, "Contagious Vorrac", "Creature — Phyrexian Boar Beast", s82OneContagiousVorracOracle, 3, 3)
		passPriorityAroundTable(t, g)
		answerChooseCards(t, g, me.ID, land)
		settleAnsweringProliferate(t, g)
		if !me.Hand.Contains(land) {
			t.Error("the land is not in hand")
		}
		if poisonOn(opp) != 1 {
			t.Errorf("poison %d, want 1: a card went to the hand", poisonOn(opp))
		}
	})
	t.Run("no land", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		seedLibrary(me, "A", "B", "C", "D", "E", "F")
		libBefore := me.Library.Size()
		castRingCreature(t, g, "Contagious Vorrac", "Creature — Phyrexian Boar Beast", s82OneContagiousVorracOracle, 3, 3)
		passPriorityAroundTable(t, g)
		settleAnsweringProliferate(t, g)
		if poisonOn(opp) != 2 {
			t.Errorf("poison %d, want 2: no card went to the hand", poisonOn(opp))
		}
		if me.Library.Size() != libBefore {
			t.Errorf("library %d, want %d", me.Library.Size(), libBefore)
		}
	})
	t.Run("declined", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		seedLibrary(me, "A", "B", "C", "D", "E")
		topLibraryCard(me, "Forest", "Basic Land — Forest")
		castRingCreature(t, g, "Contagious Vorrac", "Creature — Phyrexian Boar Beast", s82OneContagiousVorracOracle, 3, 3)
		passPriorityAroundTable(t, g)
		answerChooseCards(t, g, me.ID)
		settleAnsweringProliferate(t, g)
		if poisonOn(opp) != 2 {
			t.Errorf("poison %d, want 2: the reveal was declined", poisonOn(opp))
		}
	})
}

func TestS82OneB06CacophonyScamp(t *testing.T) {
	t.Run("sacrifice to proliferate, then its death deals damage", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		scamp := b12Push(g, me.ID, "Cacophony Scamp", "Creature — Phyrexian Goblin Warrior", s82OneCacophonyScampOracle, 1, 1)
		attackWith(t, g, opp.ID, scamp)
		passPriorityAroundTable(t, g)
		answerMayChoice(t, g, me.ID, true)
		life := opp.Life
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(scamp) {
			t.Fatal("the Scamp was not sacrificed")
		}
		// the dies trigger asks for a target
		if p := latestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("pick target: %v", err)
			}
		}
		settleAnsweringProliferate(t, g)
		if poisonOn(opp) != 2 {
			t.Errorf("poison %d, want 2", poisonOn(opp))
		}
		if opp.Life != life-1 {
			t.Errorf("life %d, want %d: the dies trigger deals 1", opp.Life, life-1)
		}
	})
	t.Run("declined keeps it", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		scamp := b12Push(g, me.ID, "Cacophony Scamp", "Creature — Phyrexian Goblin Warrior", s82OneCacophonyScampOracle, 1, 1)
		attackWith(t, g, opp.ID, scamp)
		passPriorityAroundTable(t, g)
		answerMayChoice(t, g, me.ID, false)
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(scamp) {
			t.Error("the Scamp left")
		}
		if poisonOn(opp) != 1 {
			t.Errorf("poison %d, want 1", poisonOn(opp))
		}
	})
}
