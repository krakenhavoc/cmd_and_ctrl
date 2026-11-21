package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// s82OneGivePoison puts n poison counters on p.
func s82OneGivePoison(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(p.ID, game.CounterPoison, n) })
}

func TestS82OneB06BlightbellyRatProliferatesWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := effectstest.B12Push(g, me.ID, "Blightbelly Rat", "Creature — Phyrexian Rat", s82OneBlightbellyRatOracle, 2, 2)
	s82OneGivePoison(g, opp, 1)
	if !effectstest.HasEffectiveKeyword(t, g, rat, "toxic 1") {
		t.Error("the Rat has no toxic 1")
	}
	effectstest.KillCreature(t, g, me.ID, rat)
	effectstest.SettleAnsweringProliferate(t, g)
	if effectstest.PoisonOn(opp) != 2 {
		t.Errorf("poison %d, want 2 after the dies proliferate", effectstest.PoisonOn(opp))
	}
}

func TestS82OneB06GulpingScraptrapProliferatesOnEnterAndOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s82OneGivePoison(g, opp, 1)
	trap := effectstest.CastRingCreature(t, g, "Gulping Scraptrap", "Creature — Phyrexian Horror", s82OneGulpingScraptrapOracle, 4, 4)
	effectstest.SettleAnsweringProliferate(t, g)
	if effectstest.PoisonOn(opp) != 2 {
		t.Fatalf("after entering: poison %d, want 2", effectstest.PoisonOn(opp))
	}
	effectstest.KillCreature(t, g, me.ID, trap)
	effectstest.SettleAnsweringProliferate(t, g)
	if effectstest.PoisonOn(opp) != 3 {
		t.Errorf("after dying: poison %d, want 3", effectstest.PoisonOn(opp))
	}
}

func TestS82OneB06CopperLonglegsSacrificesToProliferate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legs := effectstest.B12Push(g, me.ID, "Copper Longlegs", "Creature — Phyrexian Spider", s82OneCopperLonglegsOracle, 1, 3)
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
	effectstest.SettleAnsweringProliferate(t, g)
	if effectstest.PoisonOn(opp) != 2 {
		t.Errorf("poison %d, want 2", effectstest.PoisonOn(opp))
	}
}

func TestS82OneB06VenomousBrutalizerProliferatesOnlyIfYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		brut := effectstest.CastRingCreature(t, g, "Venomous Brutalizer", "Creature — Phyrexian Knight", s82OneVenomousBrutalizerOracle, 4, 4)
		effectstest.PassPriorityAroundTable(t, g)
		me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"})
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.SettleAnsweringProliferate(t, g)
		want := 1
		if pay {
			want = 2
		}
		if effectstest.PoisonOn(opp) != want {
			t.Errorf("pay=%v: poison %d, want %d", pay, effectstest.PoisonOn(opp), want)
		}
		if !effectstest.HasEffectiveKeyword(t, g, brut, "toxic 3") {
			t.Error("no toxic 3")
		}
	}
}

func TestS82OneB06TaintedObserverProliferatesForTwoWhenAnotherCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Tainted Observer", "Creature — Phyrexian Bird", s82OneTaintedObserverOracle, 2, 3)
	s82OneGivePoison(g, opp, 1)
	effectstest.CastRingCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.SettleAnsweringProliferate(t, g)
	if effectstest.PoisonOn(opp) != 2 {
		t.Errorf("poison %d, want 2", effectstest.PoisonOn(opp))
	}
	// The Observer's own entry is not "another creature": declining or
	// nothing to pay asks nothing.
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a payment prompt is still open")
	}
}

func TestS82OneB06TaintedObserverDeclinedPaymentDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Tainted Observer", "Creature — Phyrexian Bird", s82OneTaintedObserverOracle, 2, 3)
	s82OneGivePoison(g, opp, 1)
	effectstest.CastRingCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.AnswerPayUnless(t, g, me.ID, false)
	}
	effectstest.SettleAnsweringProliferate(t, g)
	if effectstest.PoisonOn(opp) != 1 {
		t.Errorf("poison %d, want 1: nothing was paid", effectstest.PoisonOn(opp))
	}
}

func TestS82OneB06GitaxianAnatomistTapsItselfToProliferate(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		anat := effectstest.CastRingCreature(t, g, "Gitaxian Anatomist", "Creature — Phyrexian Wizard", s82OneGitaxianAnatomistOracle, 2, 5)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, yes)
		effectstest.SettleAnsweringProliferate(t, g)
		c := effectstest.FindBattlefieldCardByID(g, anat)
		if c == nil {
			t.Fatal("the Anatomist is gone")
		}
		want := 1
		if yes {
			want = 2
		}
		if c.Tapped != yes || effectstest.PoisonOn(opp) != want {
			t.Errorf("yes=%v: tapped %v, poison %d; want tapped %v, poison %d", yes, c.Tapped, effectstest.PoisonOn(opp), yes, want)
		}
	}
}

func TestS82OneB06MesmerizingDoseTapsThenProliferatesAndKeepsItDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	s82OneGivePoison(g, opp, 1)
	effectstest.CastCatalogSpell(t, g, "Mesmerizing Dose", effectstest.AuraTypeLine, s82OneMesmerizingDoseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SettleAnsweringProliferate(t, g)
	if !effectstest.TappedForTest(t, g, bear) {
		t.Error("the enchanted creature was not tapped")
	}
	if effectstest.PoisonOn(opp) != 2 {
		t.Errorf("poison %d, want 2", effectstest.PoisonOn(opp))
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.TappedForTest(t, g, bear) {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

func TestS82OneB06AdaptiveSporesingerModes(t *testing.T) {
	t.Run("pump", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.CastRingCreature(t, g, "Adaptive Sporesinger", "Creature — Phyrexian Druid", s82OneAdaptiveSporesingerOracle, 2, 2)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
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
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		effectstest.CastRingCreature(t, g, "Adaptive Sporesinger", "Creature — Phyrexian Druid", s82OneAdaptiveSporesingerOracle, 2, 2)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		effectstest.SettleAnsweringProliferate(t, g)
		if effectstest.PoisonOn(opp) != 2 {
			t.Errorf("poison %d, want 2", effectstest.PoisonOn(opp))
		}
	})
}

func TestS82OneB06ContagiousVorracTakesALandOrProliferates(t *testing.T) {
	t.Run("land taken", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
		land := effectstest.TopLibraryCard(me, "Forest", "Basic Land — Forest")
		effectstest.CastRingCreature(t, g, "Contagious Vorrac", "Creature — Phyrexian Boar Beast", s82OneContagiousVorracOracle, 3, 3)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID, land)
		effectstest.SettleAnsweringProliferate(t, g)
		if !me.Hand.Contains(land) {
			t.Error("the land is not in hand")
		}
		if effectstest.PoisonOn(opp) != 1 {
			t.Errorf("poison %d, want 1: a card went to the hand", effectstest.PoisonOn(opp))
		}
	})
	t.Run("no land", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
		libBefore := me.Library.Size()
		effectstest.CastRingCreature(t, g, "Contagious Vorrac", "Creature — Phyrexian Boar Beast", s82OneContagiousVorracOracle, 3, 3)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.SettleAnsweringProliferate(t, g)
		if effectstest.PoisonOn(opp) != 2 {
			t.Errorf("poison %d, want 2: no card went to the hand", effectstest.PoisonOn(opp))
		}
		if me.Library.Size() != libBefore {
			t.Errorf("library %d, want %d", me.Library.Size(), libBefore)
		}
	})
	t.Run("declined", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
		effectstest.TopLibraryCard(me, "Forest", "Basic Land — Forest")
		effectstest.CastRingCreature(t, g, "Contagious Vorrac", "Creature — Phyrexian Boar Beast", s82OneContagiousVorracOracle, 3, 3)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID)
		effectstest.SettleAnsweringProliferate(t, g)
		if effectstest.PoisonOn(opp) != 2 {
			t.Errorf("poison %d, want 2: the reveal was declined", effectstest.PoisonOn(opp))
		}
	})
}

func TestS82OneB06CacophonyScamp(t *testing.T) {
	t.Run("sacrifice to proliferate, then its death deals damage", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		scamp := effectstest.B12Push(g, me.ID, "Cacophony Scamp", "Creature — Phyrexian Goblin Warrior", s82OneCacophonyScampOracle, 1, 1)
		effectstest.AttackWith(t, g, opp.ID, scamp)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, true)
		life := opp.Life
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(scamp) {
			t.Fatal("the Scamp was not sacrificed")
		}
		// the dies trigger asks for a target
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("pick target: %v", err)
			}
		}
		effectstest.SettleAnsweringProliferate(t, g)
		if effectstest.PoisonOn(opp) != 2 {
			t.Errorf("poison %d, want 2", effectstest.PoisonOn(opp))
		}
		if opp.Life != life-1 {
			t.Errorf("life %d, want %d: the dies trigger deals 1", opp.Life, life-1)
		}
	})
	t.Run("declined keeps it", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		scamp := effectstest.B12Push(g, me.ID, "Cacophony Scamp", "Creature — Phyrexian Goblin Warrior", s82OneCacophonyScampOracle, 1, 1)
		effectstest.AttackWith(t, g, opp.ID, scamp)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(scamp) {
			t.Error("the Scamp left")
		}
		if effectstest.PoisonOn(opp) != 1 {
			t.Errorf("poison %d, want 1", effectstest.PoisonOn(opp))
		}
	})
}
