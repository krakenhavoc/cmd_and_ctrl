package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b140_test.go — Throne of Eldraine adamant cards, S82 set OLD batch
// 140 (#5516). Every test plays the real catalog entry through the engine,
// paying a floating pool strictly, so adamant reads what was actually spent.

const (
	oldB140Rally       = "0a557383-596b-4b8f-b81f-21706fef7b63"
	oldB140Cauldron    = "152bbb62-195b-4de1-90c6-10dd953f8c4b"
	oldB140Outmuscle   = "185d84f4-6095-46ef-917f-3d662a86aa81"
	oldB140Vantress    = "1965ca0e-1357-4437-82d3-1a78353ab5f1"
	oldB140Barrage     = "26c17025-1123-46d9-a8bd-edbad01c4b6e"
	oldB140Ardenvale   = "30393cb0-f461-4899-8aae-609bc8378f36"
	oldB140Garenbrig   = "33fdbff6-0e89-4f02-b429-9b17869134cd"
	oldB140Clockwork   = "67dae813-907e-456d-9e91-2d7c85f1d574"
	oldB140Locthwain   = "6fea3c95-9663-4857-983e-eaf522efa1a8"
	oldB140HengeWalker = "d4c13349-2faa-46ec-bd8e-80a4e59ef407"
	oldB140Silverflame = "d511573b-35c7-4934-aa46-906e14101664"
	oldB140Embereth    = "eec31fe0-a633-4424-84d2-90fe88016310"
	oldB140OnceFuture  = "f666c38d-e601-4aa9-b140-4b6607bb12dd"
)

func TestOldB140EveryCardIsRegisteredWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB140Rally: "Rally for the Throne", oldB140Cauldron: "Cauldron's Gift", oldB140Outmuscle: "Outmuscle",
		oldB140Vantress: "Vantress Paladin", oldB140Barrage: "Searing Barrage", oldB140Ardenvale: "Ardenvale Paladin",
		oldB140Garenbrig: "Garenbrig Paladin", oldB140Clockwork: "Clockwork Servant", oldB140Locthwain: "Locthwain Paladin",
		oldB140HengeWalker: "Henge Walker", oldB140Silverflame: "Silverflame Ritual", oldB140Embereth: "Embereth Paladin",
		oldB140OnceFuture: "Once and Future",
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

// oldB140Cast casts a catalog spell from the active player's hand, strictly
// paid from `pool`, and resolves it.
func oldB140Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle, pool string, targets []game.TargetRef, pt ...int) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.FloatForTest(g, me, pool)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID,
			Power: pt0(pt, 0), Toughness: pt0(pt, 1),
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// pt0 is the i-th of an optional power/toughness pair, zero when absent.
func pt0(pt []int, i int) int {
	if i < len(pt) {
		return pt[i]
	}
	return 0
}

func oldB140Ref(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// oldB140Permanent finds the battlefield card with a name.
func oldB140Permanent(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	t.Fatalf("%s is not on the battlefield", name)
	return uuid.Nil
}

func TestOldB140PaladinsEnterWithACounterOnlyWhenAdamantWasSpent(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cost, adamant, plain string
		power, toughness                   int
	}{
		{"Vantress Paladin", oldB140Vantress, "{3}{U}", "UUUC", "UCCC", 2, 2},
		{"Embereth Paladin", oldB140Embereth, "{3}{R}", "RRRC", "RCCC", 4, 1},
		{"Locthwain Paladin", oldB140Locthwain, "{3}{B}", "BBBC", "BCCC", 3, 2},
		{"Ardenvale Paladin", oldB140Ardenvale, "{3}{W}", "WWWC", "WCCC", 2, 5},
		{"Garenbrig Paladin", oldB140Garenbrig, "{4}{G}", "GGGCC", "GCCCC", 4, 4},
	} {
		for _, paid := range []struct {
			pool string
			plus int
		}{{tc.adamant, 1}, {tc.plain, 0}} {
			t.Run(tc.name+" paid with "+paid.pool, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				oldB140Cast(t, g, tc.name, "Creature — Knight", tc.cost, tc.oracle, paid.pool, nil, tc.power, tc.toughness)
				id := oldB140Permanent(t, g, tc.name)
				if got := effectstest.B12Counter(t, g, id, game.CounterPlusOne); got != paid.plus {
					t.Errorf("%d +1/+1 counters, want %d", got, paid.plus)
				}
				if p, th := effectstest.PtOf(t, g, id); p != tc.power+paid.plus || th != tc.toughness+paid.plus {
					t.Errorf("%d/%d, want %d/%d", p, th, tc.power+paid.plus, tc.toughness+paid.plus)
				}
			})
		}
	}
}

func TestOldB140APaladinPutOntoTheBattlefieldGetsNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Ardenvale Paladin", "Creature — Human Knight", oldB140Ardenvale, 2, 5)
	if got := effectstest.B12Counter(t, g, id, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters on a Paladin that was never cast", got)
	}
}

func TestOldB140GarenbrigPaladinCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	for _, tc := range []struct {
		power int
		legal bool
	}{{2, false}, {3, true}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		pal := effectstest.B12Push(g, me.ID, "Garenbrig Paladin", "Creature — Giant Knight", oldB140Garenbrig, 4, 4)
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", tc.power, 4)
		effectstest.DeclareAttack(t, g, opp.ID, pal)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(blocker, pal); (err == nil) != tc.legal {
			t.Errorf("power %d blocker: err %v, want legal=%v", tc.power, err, tc.legal)
		}
	}
}

func TestOldB140HengeWalkerNeedsThreeManaOfOneColour(t *testing.T) {
	for _, tc := range []struct {
		pool string
		plus int
	}{{"WWW", 1}, {"RRR", 1}, {"WWU", 0}, {"CCC", 0}, {"WWC", 0}} {
		t.Run(tc.pool, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			oldB140Cast(t, g, "Henge Walker", "Artifact Creature — Golem", "{3}", oldB140HengeWalker, tc.pool, nil, 2, 2)
			id := oldB140Permanent(t, g, "Henge Walker")
			if got := effectstest.B12Counter(t, g, id, game.CounterPlusOne); got != tc.plus {
				t.Errorf("%d counters, want %d", got, tc.plus)
			}
		})
	}
}

func TestOldB140ClockworkServantDrawsOnlyWithThreeManaOfOneColour(t *testing.T) {
	for _, tc := range []struct {
		pool string
		draw int
	}{{"BBB", 1}, {"GGG", 1}, {"BBG", 0}, {"CCC", 0}} {
		t.Run(tc.pool, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			hand := me.Hand.Size()
			oldB140Cast(t, g, "Clockwork Servant", "Artifact Creature — Gnome", "{3}", oldB140Clockwork, tc.pool, nil)
			if got := me.Hand.Size() - hand; got != tc.draw {
				t.Errorf("drew %d, want %d", got, tc.draw)
			}
		})
	}
}

func TestOldB140ClockworkServantThatLeftBeforeTheTriggerResolvesDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.FloatForTest(g, me, "BBB")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Clockwork Servant", TypeLine: "Artifact Creature — Gnome", ManaCost: "{3}",
		OracleID: oldB140Clockwork, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	// Resolve only the creature spell, leaving the enter trigger on the stack.
	effectstest.PassUntilOnBattlefield(t, g, id)
	if effectstest.TriggersOnStackFrom(g, id) != 1 {
		t.Fatalf("%d triggers on the stack, want 1", effectstest.TriggersOnStackFrom(g, id))
	}
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}, game.DestroyOptions{}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: a Servant that left still drew", me.Hand.Size(), hand)
	}
}

func TestOldB140RallyForTheThroneMakesTwoHumansAndGainsLifeOnlyWithAdamant(t *testing.T) {
	for _, tc := range []struct {
		pool string
		gain int
	}{{"WWW", 3}, {"WWC", 0}} {
		t.Run(tc.pool, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
			life := me.Life
			oldB140Cast(t, g, "Rally for the Throne", "Instant", "{2}{W}", oldB140Rally, tc.pool, nil)
			if n, _ := effectstest.CountTokensNamed(g, me.ID, "Human"); n != 2 {
				t.Errorf("%d Human tokens, want 2", n)
			}
			if me.Life != life+tc.gain {
				t.Errorf("life %d, want %d", me.Life, life+tc.gain)
			}
		})
	}
}

func TestOldB140CauldronsGiftMillsOnlyWithAdamantThenReturnsACreatureWithACounter(t *testing.T) {
	for _, tc := range []struct {
		pool string
		mill int
	}{{"BBBCC", 4}, {"BCCCC", 0}} {
		t.Run(tc.pool, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedCheapLibrary(g, me, 10)
			beast := effectstest.PushGraveyardPermanent(me, "Dead Beast", "Creature — Beast", "{3}")
			land := effectstest.PushGraveyardPermanent(me, "Dead Rock", "Artifact", "{2}")
			lib := me.Library.Size()
			oldB140Cast(t, g, "Cauldron's Gift", "Sorcery", "{4}{B}", oldB140Cauldron, tc.pool, nil)
			if got := lib - me.Library.Size(); got != tc.mill {
				t.Errorf("milled %d, want %d", got, tc.mill)
			}
			pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
			if pick == nil {
				t.Fatal("no prompt for the creature card")
			}
			if effectstest.HasID(pick.ChooseCards, land) || !effectstest.HasID(pick.ChooseCards, beast) {
				t.Errorf("offered %v: want the creature card only", pick.ChooseCards)
			}
			effectstest.AnswerChooseCards(t, g, me.ID, beast)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(beast) {
				t.Fatal("the creature did not return to the battlefield")
			}
			if got := effectstest.B12Counter(t, g, beast, game.CounterPlusOne); got != 1 {
				t.Errorf("%d +1/+1 counters, want 1", got)
			}
		})
	}
}

func TestOldB140CauldronsGiftMayChooseNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	beast := effectstest.PushGraveyardPermanent(me, "Dead Beast", "Creature — Beast", "{3}")
	oldB140Cast(t, g, "Cauldron's Gift", "Sorcery", "{4}{B}", oldB140Cauldron, "BCCCC", nil)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(beast) || !me.Graveyard.Contains(beast) {
		t.Error("a card was returned although none was chosen")
	}
}

func TestOldB140OutmuscleCountersFightsAndAdamantGrantsIndestructible(t *testing.T) {
	for _, tc := range []struct {
		pool string
		indy bool
	}{{"GGGC", true}, {"GCCC", false}} {
		t.Run(tc.pool, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 5)
			oldB140Cast(t, g, "Outmuscle", "Sorcery", "{3}{G}", oldB140Outmuscle, tc.pool, []game.TargetRef{oldB140Ref(mine), oldB140Ref(theirs)})
			if got := effectstest.B12Counter(t, g, mine, game.CounterPlusOne); got != 1 {
				t.Errorf("%d counters on your creature, want 1", got)
			}
			if got := effectstest.DamageMarkedOn(g, theirs); got != 3 {
				t.Errorf("their creature took %d, want 3 (the countered power)", got)
			}
			if got := effectstest.HasEffectiveKeyword(t, g, mine, "indestructible"); got != tc.indy {
				t.Errorf("indestructible = %v, want %v", got, tc.indy)
			}
			if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
				t.Error("the other creature gained indestructible")
			}
		})
	}
}

func TestOldB140SearingBarrageHitsTheControllerOnlyWithAdamant(t *testing.T) {
	for _, tc := range []struct {
		pool string
		hit  int
	}{{"RRRCC", 3}, {"RCCCC", 0}} {
		t.Run(tc.pool, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
			big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 6, 6)
			small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
			life := opp.Life
			oldB140Cast(t, g, "Searing Barrage", "Instant", "{4}{R}", oldB140Barrage, tc.pool, []game.TargetRef{oldB140Ref(small)})
			if g.Battlefield.Contains(small) {
				t.Error("the target survived 5 damage")
			}
			if !g.Battlefield.Contains(big) {
				t.Error("another creature died")
			}
			if opp.Life != life-tc.hit {
				t.Errorf("its controller is at %d, want %d", opp.Life, life-tc.hit)
			}
		})
	}
}

func TestOldB140SilverflameRitualCountersEveryCreatureAndAdamantGrantsVigilance(t *testing.T) {
	for _, tc := range []struct {
		pool string
		vig  bool
	}{{"WWWC", true}, {"WCCC", false}} {
		t.Run(tc.pool, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
			b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			oldB140Cast(t, g, "Silverflame Ritual", "Sorcery", "{3}{W}", oldB140Silverflame, tc.pool, nil)
			for _, id := range []uuid.UUID{a, b} {
				if effectstest.B12Counter(t, g, id, game.CounterPlusOne) != 1 {
					t.Error("a creature you control got no counter")
				}
				if effectstest.HasEffectiveKeyword(t, g, id, "vigilance") != tc.vig {
					t.Errorf("vigilance = %v, want %v", !tc.vig, tc.vig)
				}
			}
			if effectstest.B12Counter(t, g, theirs, game.CounterPlusOne) != 0 || effectstest.HasEffectiveKeyword(t, g, theirs, "vigilance") {
				t.Error("an opponent's creature was affected")
			}
		})
	}
}

func TestOldB140OnceAndFutureReturnsOneAndPutsTheOtherOnTopOrReturnsBothWithAdamant(t *testing.T) {
	for _, tc := range []struct {
		pool     string
		bothHand bool
	}{{"GCCC", false}, {"GGGC", true}} {
		t.Run(tc.pool, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			first := effectstest.PushGraveyardCardForTest(me, "First Card")
			second := effectstest.PushGraveyardCardForTest(me, "Second Card")
			spell := oldB140Cast(t, g, "Once and Future", "Instant", "{3}{G}", oldB140OnceFuture, tc.pool,
				[]game.TargetRef{oldB140Ref(first), oldB140Ref(second)})
			if !me.Hand.Contains(first) {
				t.Error("the first card did not return to hand")
			}
			if tc.bothHand {
				if !me.Hand.Contains(second) {
					t.Error("adamant: the second card did not return to hand")
				}
			} else {
				top, err := me.Library.Top()
				if err != nil || top.InstanceID != second {
					t.Errorf("the second card is not on top of the library (%v)", err)
				}
				if me.Hand.Contains(second) {
					t.Error("the second card went to hand without adamant")
				}
			}
			if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
				t.Error("Once and Future was not exiled")
			}
		})
	}
}

func TestOldB140OnceAndFutureTakesOnlyOneTargetAndNeedsAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := effectstest.PushGraveyardCardForTest(me, "First Card")
	if err := effectstest.CastCatalogSpellErr(t, g, "Once and Future", "Instant", oldB140OnceFuture, []game.TargetRef{oldB140Ref(first), oldB140Ref(first)}); err == nil {
		t.Error("the same card was accepted as both targets")
	}
	spell := oldB140Cast(t, g, "Once and Future", "Instant", "{3}{G}", oldB140OnceFuture, "GCCC", []game.TargetRef{oldB140Ref(first)})
	if !me.Hand.Contains(first) || !g.Exile.Contains(spell) {
		t.Error("a single target did not resolve")
	}
}
