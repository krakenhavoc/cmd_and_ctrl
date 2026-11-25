package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 297 (Oath of the Gatewatch: surge, #5545). Every test plays the
// real catalog entry through the engine.

const (
	oldB297Salvo      = "16df20e2-5228-4543-84b7-74a3170d9d31"
	oldB297Avenger    = "2db8fc61-13af-4227-83f9-9f426ab978d3"
	oldB297Freerunner = "372094f8-86dc-4c78-a5d0-444ffbf02630"
	oldB297Denial     = "45aa8bdb-54da-4069-bd6c-c63396154464"
	oldB297Crush      = "4ede153c-89cb-4121-bf23-a1d213d6f238"
	oldB297Bushwhack  = "4fb83127-00cf-4ce7-aedf-cb0b2740c66f"
	oldB297Titans     = "55e564b4-7c08-4109-bd67-da5f31fe002e"
	oldB297Analysis   = "59b869b9-cb48-4152-ad91-52f582e2df5b"
	oldB297Tyrant     = "a4ddd9e2-ea59-4bba-bb2f-a655cabbb65c"
	oldB297Grip       = "c3c87292-2fc8-4759-b7b6-18a8b539b5d1"
	oldB297Membrane   = "e3bf1a95-4363-4915-8f43-b423714861ec"
)

func TestOldB297EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB297Salvo: "Boulder Salvo", oldB297Avenger: "Jwar Isle Avenger",
		oldB297Freerunner: "Goblin Freerunner", oldB297Denial: "Overwhelming Denial",
		oldB297Crush: "Crush of Tentacles", oldB297Bushwhack: "Reckless Bushwhacker",
		oldB297Titans: "Fall of the Titans", oldB297Analysis: "Comparative Analysis",
		oldB297Tyrant: "Tyrant of Valakut", oldB297Grip: "Grip of the Roil",
		oldB297Membrane: "Containment Membrane",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != "surge" {
			t.Errorf("%s: want exactly the surge alternative cost", name)
		}
	}
}

// oldB297Cast puts a card in the active seat's hand in a main phase and casts
// it, for its surge cost when surge is set.
func oldB297Cast(t *testing.T, g *game.Game, c game.Card, surge bool, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	params := game.CastSpellParams{Strict: true, Targets: targets}
	if surge {
		params.AlternativeCost = "surge"
	}
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, params)
}

// oldB297Cast0 is oldB297Cast for a cast that must be accepted.
func oldB297Cast0(t *testing.T, g *game.Game, c game.Card, surge bool, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id, err := oldB297Cast(t, g, c, surge, targets)
	if err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return id
}

// oldB297Warm makes the active seat's tally say it has cast a spell this turn.
func oldB297Warm(g *game.Game) {
	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{g.Seats[g.Turn.ActiveSeat].ID: {Total: 1}}
	})
}

func oldB297Foe(g *game.Game, toughness int) uuid.UUID {
	id := uuid.New()
	foe := g.Seats[1]
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Foe", TypeLine: "Creature — Bear",
			Power: 2, Toughness: toughness, Owner: foe.ID, Controller: foe.ID})
	})
	return id
}

func TestOldB297SurgeNeedsAnotherSpellThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[0]
	foe := oldB297Foe(g, 4)
	card := game.Card{Name: "Boulder Salvo", TypeLine: "Sorcery", ManaCost: "{4}{R}", OracleID: oldB297Salvo, Colors: []string{"R"}}
	effectstest.FloatForTest(g, me, "CR")
	if _, err := oldB297Cast(t, g, card, true, effectstest.CardRefs(foe)); err == nil {
		t.Fatal("surge was accepted as the first spell of the turn")
	}
	oldB297Warm(g)
	if _, err := oldB297Cast(t, g, card, true, effectstest.CardRefs(foe)); err != nil {
		t.Fatalf("surge as the second spell: %v", err)
	}
}

func TestOldB297BoulderSalvoDealsFourToACreature(t *testing.T) {
	for _, tc := range []struct {
		toughness int
		dies      bool
	}{{4, true}, {5, false}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMainForCost(t, g)
		me := g.Seats[0]
		foe := oldB297Foe(g, tc.toughness)
		oldB297Warm(g)
		effectstest.FloatForTest(g, me, "CR")
		oldB297Cast0(t, g, game.Card{Name: "Boulder Salvo", TypeLine: "Sorcery", ManaCost: "{4}{R}", OracleID: oldB297Salvo, Colors: []string{"R"}},
			true, effectstest.CardRefs(foe))
		effectstest.PassPriorityAroundTable(t, g)
		if got := !effectstest.OnBattlefieldByID(g, foe); got != tc.dies {
			t.Errorf("toughness %d: died %v, want %v", tc.toughness, got, tc.dies)
		}
	}
}

func TestOldB297SurgedCreaturesKeepTheirKeywords(t *testing.T) {
	for _, tc := range []struct {
		card    game.Card
		mana    string
		keyword string
	}{
		{game.Card{Name: "Jwar Isle Avenger", TypeLine: "Creature — Sphinx", ManaCost: "{4}{U}", OracleID: oldB297Avenger, Colors: []string{"U"}, Power: 3, Toughness: 3}, "CCU", "flying"},
		{game.Card{Name: "Goblin Freerunner", TypeLine: "Creature — Goblin Warrior Ally", ManaCost: "{3}{R}", OracleID: oldB297Freerunner, Colors: []string{"R"}, Power: 3, Toughness: 2}, "CR", "menace"},
	} {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMainForCost(t, g)
		me := g.Seats[0]
		oldB297Warm(g)
		effectstest.FloatForTest(g, me, tc.mana)
		id := oldB297Cast0(t, g, tc.card, true, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s lacks %s", tc.card.Name, tc.keyword)
		}
		if me.ManaPool.Total() != 0 {
			t.Errorf("%s: %d mana left, the surge cost is the whole price", tc.card.Name, me.ManaPool.Total())
		}
	}
}

func TestOldB297OverwhelmingDenialCountersASpellAndCannotBeCountered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[0]
	victim := effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil) // the first spell: surge is on
	effectstest.FloatForTest(g, me, "UU")
	oldB297Cast0(t, g, game.Card{Name: "Overwhelming Denial", TypeLine: "Instant", ManaCost: "{2}{U}{U}", OracleID: oldB297Denial, Colors: []string{"U"}},
		true, effectstest.CardRefs(victim))
	spec, _ := effects.Lookup(oldB297Denial)
	if !spec.CantBeCountered {
		t.Error("Overwhelming Denial can be countered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(victim) {
		t.Error("the targeted spell was not countered")
	}
}

func TestOldB297CrushOfTentaclesBouncesEverythingAndSurgeMakesAnOctopus(t *testing.T) {
	for _, surge := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMainForCost(t, g)
		me := g.Seats[0]
		mine := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Mine")
		theirs := oldB297Foe(g, 2)
		land := effectstest.FreeSpellPermanent(g, me.ID, "Island", "Basic Land — Island")
		mana := "CCCCUU"
		if surge {
			oldB297Warm(g)
			mana = "CCCUU"
		}
		effectstest.FloatForTest(g, me, mana)
		oldB297Cast0(t, g, game.Card{Name: "Crush of Tentacles", TypeLine: "Sorcery", ManaCost: "{4}{U}{U}", OracleID: oldB297Crush, Colors: []string{"U"}}, surge, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldByID(g, mine) || effectstest.OnBattlefieldByID(g, theirs) {
			t.Errorf("surge %v: a creature stayed on the battlefield", surge)
		}
		if !effectstest.OnBattlefieldByID(g, land) {
			t.Errorf("surge %v: the land was bounced", surge)
		}
		if got := effectstest.CountOnBattlefield(g, "Octopus", me.ID); (got == 1) != surge || got > 1 {
			t.Errorf("surge %v: %d Octopus tokens", surge, got)
		}
	}
}

func TestOldB297RecklessBushwhackerPumpsAndHastesOthersOnlyWithSurge(t *testing.T) {
	for _, surge := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMainForCost(t, g)
		me := g.Seats[0]
		other := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Other")
		mana := "CCR"
		if surge {
			oldB297Warm(g)
			mana = "CR"
		}
		effectstest.FloatForTest(g, me, mana)
		self := oldB297Cast0(t, g, game.Card{Name: "Reckless Bushwhacker", TypeLine: "Creature — Goblin Warrior Ally", ManaCost: "{2}{R}",
			OracleID: oldB297Bushwhack, Colors: []string{"R"}, Power: 2, Toughness: 1}, surge, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		p, _ := effectstest.EffectivePTOf(t, g, other)
		haste := effectstest.HasKeywordOnBattlefield(t, g, other, "haste")
		if surge && (p != 3 || !haste) || !surge && (p != 2 || haste) {
			t.Errorf("surge %v: other creature power %d haste %v", surge, p, haste)
		}
		if sp, _ := effectstest.EffectivePTOf(t, g, self); sp != 2 {
			t.Errorf("surge %v: the Bushwhacker itself has power %d", surge, sp)
		}
	}
}

func TestOldB297FallOfTheTitansDealsXToEachOfTwoTargets(t *testing.T) {
	for _, tc := range []struct {
		surge bool
		x     int
		mana  string
	}{{true, 3, "CCCR"}, {false, 3, "CCCCCCR"}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMainForCost(t, g)
		me := g.Seats[0]
		a, b := oldB297Foe(g, 3), oldB297Foe(g, 4)
		if tc.surge {
			oldB297Warm(g)
		}
		effectstest.FloatForTest(g, me, tc.mana)
		id, err := oldB297CastX(t, g, game.Card{Name: "Fall of the Titans", TypeLine: "Instant", ManaCost: "{X}{X}{R}", OracleID: oldB297Titans, Colors: []string{"R"}},
			tc.surge, tc.x, effectstest.CardRefs(a, b))
		if err != nil {
			t.Fatalf("surge %v: cast %v", tc.surge, id)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldByID(g, a) {
			t.Errorf("surge %v: the 3-toughness target survived X=3", tc.surge)
		}
		if !effectstest.OnBattlefieldByID(g, b) {
			t.Errorf("surge %v: the 4-toughness target died to X=3", tc.surge)
		}
		if me.ManaPool.Total() != 0 {
			t.Errorf("surge %v: %d mana left over", tc.surge, me.ManaPool.Total())
		}
	}
}

func oldB297CastX(t *testing.T, g *game.Game, c game.Card, surge bool, x int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	params := game.CastSpellParams{Strict: true, Targets: targets, XValue: x}
	if surge {
		params.AlternativeCost = "surge"
	}
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, params)
}

func TestOldB297ComparativeAnalysisTargetPlayerDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me, foe := g.Seats[0], g.Seats[1]
	oldB297Warm(g)
	effectstest.FloatForTest(g, me, "CCU")
	before := foe.Hand.Size()
	oldB297Cast0(t, g, game.Card{Name: "Comparative Analysis", TypeLine: "Instant", ManaCost: "{3}{U}", OracleID: oldB297Analysis, Colors: []string{"U"}},
		true, effectstest.PlayerRefs(foe.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := foe.Hand.Size(); got != before+2 {
		t.Errorf("hand %d, want %d", got, before+2)
	}
}

func TestOldB297TyrantOfValakutShootsOnlyWithSurge(t *testing.T) {
	for _, surge := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMainForCost(t, g)
		me, foe := g.Seats[0], g.Seats[1]
		mana := "CCCCCRR"
		if surge {
			oldB297Warm(g)
			mana = "CCCRR"
		}
		effectstest.FloatForTest(g, me, mana)
		life := foe.Life
		oldB297Cast0(t, g, game.Card{Name: "Tyrant of Valakut", TypeLine: "Creature — Dragon", ManaCost: "{5}{R}{R}", OracleID: oldB297Tyrant,
			Colors: []string{"R"}, Power: 5, Toughness: 4}, surge, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if surge {
			ch := effectstest.PickTargetPrompt(t, g)
			if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: foe.ID}); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		want := life
		if surge {
			want -= 3
		}
		if foe.Life != want {
			t.Errorf("surge %v: life %d, want %d", surge, foe.Life, want)
		}
	}
}

func TestOldB297GripOfTheRoilTapsFreezesAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[0]
	foe := oldB297Foe(g, 2)
	oldB297Warm(g)
	effectstest.FloatForTest(g, me, "CU")
	hand := me.Hand.Size()
	oldB297Cast0(t, g, game.Card{Name: "Grip of the Roil", TypeLine: "Instant", ManaCost: "{2}{U}", OracleID: oldB297Grip, Colors: []string{"U"}},
		true, effectstest.CardRefs(foe))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, foe) {
		t.Error("the target is not tapped")
	}
	if got := me.Hand.Size(); got != hand+1 { // the spell left the hand, one card drawn
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if !effectstest.TappedForTest(t, g, foe) {
		t.Error("the target untapped during its controller's next untap step")
	}
}

func TestOldB297ContainmentMembraneHoldsTheCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[0]
	foe := oldB297Foe(g, 2)
	if err := g.TapCard(foe, true); err != nil {
		t.Fatal(err)
	}
	oldB297Warm(g)
	effectstest.FloatForTest(g, me, "U")
	oldB297Cast0(t, g, game.Card{Name: "Containment Membrane", TypeLine: "Enchantment — Aura", ManaCost: "{2}{U}", OracleID: oldB297Membrane, Colors: []string{"U"}},
		true, effectstest.CardRefs(foe))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if !effectstest.TappedForTest(t, g, foe) {
		t.Error("the enchanted creature untapped")
	}
}
