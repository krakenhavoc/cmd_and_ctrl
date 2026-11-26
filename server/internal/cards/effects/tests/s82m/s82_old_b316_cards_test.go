package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b316_cards_test.go — Duel Decks: Zendikar vs. Eldrazi, batch 316
// (S82, #5548). Each test plays the real catalog entry through the engine.

const (
	oldB316KnightOracle    = "5045cfc3-ea63-4e0d-bf6a-bdeaf841ae3c"
	oldB316BeastbreakerOrc = "8c77f436-81ab-4200-b76d-58b1f91fc3d0"
	oldB316EscortOracle    = "9a9c221d-9bb4-4ec4-b5db-cb26db6065df"
	oldB316ArcherOracle    = "5ef498fa-31d5-4eca-873d-6b2b1f7f9a3c"
	oldB316SlingerOracle   = "64231f03-ea89-4589-8ef8-a87dc9448e7b"
	oldB316DespairOracle   = "1e8ff8b3-5b7f-47f2-9321-7a3b6f6e88aa"
	oldB316InvokerOracle   = "2dd3a6f7-65a9-4157-9ad9-13a885ebe927"
	oldB316GuideOracle     = "5f685846-60dc-45f9-b118-02e74a443e6a"
	oldB316GroveOracle     = "62934aab-b6fa-41b2-ac05-d3fa0e3c5a79"
	oldB316GriffinOracle   = "af2a951f-7630-48fe-bb37-273b66901eb6"
	oldB316MeekOracle      = "c2555c38-1bfc-48c1-85aa-7e40117db78b"
	oldB316MarshOracle     = "d2bd182c-532f-4c77-adb8-754f2c2fcd82"
	oldB316EruptionOracle  = "e1894a57-0faa-4e0d-8ab9-e4785f0457a4"
)

func TestOldB316EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB316KnightOracle:    "Knight of Cliffhaven",
		oldB316BeastbreakerOrc: "Beastbreaker of Bala Ged",
		oldB316EscortOracle:    "Caravan Escort",
		oldB316ArcherOracle:    "Tajuru Archer",
		oldB316SlingerOracle:   "Torch Slinger",
		oldB316DespairOracle:   "Induce Despair",
		oldB316InvokerOracle:   "Bloodrite Invoker",
		oldB316GuideOracle:     "Frontier Guide",
		oldB316GroveOracle:     "Turntimber Grove",
		oldB316GriffinOracle:   "Makindi Griffin",
		oldB316MeekOracle:      "Consume the Meek",
		oldB316MarshOracle:     "Marsh Casualties",
		oldB316EruptionOracle:  "Hellion Eruption",
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

func TestOldB316MakindiGriffinFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Makindi Griffin", "Creature — Griffin", oldB316GriffinOracle, 2, 4)
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("no flying")
	}
}

// --- the levelers ----------------------------------------------------------

func TestOldB316LevelersGrowWithLevelCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, cost string
		basePower, baseTough         int
		// stats[i] is the effective power, toughness and keywords after i+1 level-ups.
		stats []struct {
			p, tough int
			kw       []string
		}
	}{
		{"Knight of Cliffhaven", "Creature — Kor Knight", oldB316KnightOracle, "CCC", 2, 2, []struct {
			p, tough int
			kw       []string
		}{{2, 3, []string{"flying"}}, {2, 3, []string{"flying"}}, {2, 3, []string{"flying"}}, {4, 4, []string{"flying", "vigilance"}}}},
		{"Beastbreaker of Bala Ged", "Creature — Human Warrior", oldB316BeastbreakerOrc, "CCG", 2, 2, []struct {
			p, tough int
			kw       []string
		}{{4, 4, nil}, {4, 4, nil}, {4, 4, nil}, {6, 6, []string{"trample"}}}},
		{"Caravan Escort", "Creature — Human Knight", oldB316EscortOracle, "CC", 1, 1, []struct {
			p, tough int
			kw       []string
		}{{2, 2, nil}, {2, 2, nil}, {2, 2, nil}, {2, 2, nil}, {5, 5, []string{"first strike"}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.basePower, tc.baseTough)
			if p, tough := effectstest.PtOf(t, g, id); p != tc.basePower || tough != tc.baseTough {
				t.Fatalf("level 0 is %d/%d, want %d/%d", p, tough, tc.basePower, tc.baseTough)
			}
			for i, want := range tc.stats {
				effectstest.FloatForTest(g, me, tc.cost)
				effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
				effectstest.PassPriorityAroundTable(t, g)
				if n := effectstest.CounterCount(g, id, game.CounterLevel); n != i+1 {
					t.Fatalf("%d level counters, want %d", n, i+1)
				}
				if p, tough := effectstest.PtOf(t, g, id); p != want.p || tough != want.tough {
					t.Errorf("level %d is %d/%d, want %d/%d", i+1, p, tough, want.p, want.tough)
				}
				for _, kw := range want.kw {
					if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
						t.Errorf("level %d has no %s", i+1, kw)
					}
				}
			}
		})
	}
}

func TestOldB316LevelUpIsSorceryOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Caravan Escort", "Creature — Human Knight", oldB316EscortOracle, 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("level up was activated outside a main phase")
	}
}

func TestOldB316LevelersKeepTheirNoKeywordsAtLevelZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Knight of Cliffhaven", "Creature — Kor Knight", oldB316KnightOracle, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("the Knight flies at level 0")
	}
}

// --- Tajuru Archer ---------------------------------------------------------

func TestOldB316TajuruArcherShootsAFlyerForTheAllyCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Other Ally", "Creature — Human Ally", "", 1, 1)
	flyer := effectstest.B12Push(g, opp.ID, "Flyer", "Creature — Bird", "", 1, 4)
	ground := effectstest.B12Push(g, opp.ID, "Ground", "Creature — Bear", "", 1, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flyer {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	effectstest.CastHoldCreature(t, g, "Tajuru Archer", "Creature — Elf Archer Ally", oldB316ArcherOracle, 1, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, ground) {
		t.Error("a creature without flying is offered")
	}
	effectstest.PickCard(t, g, me.ID, flyer)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, flyer); got != 2 {
		t.Errorf("the flyer took %d damage, want 2 (two Allies)", got)
	}
}

// --- Torch Slinger ---------------------------------------------------------

func TestOldB316TorchSlingerShootsOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 2)
		cost := "CCR"
		var opts []int
		if kicked {
			cost, opts = "CCRCR", []int{0}
		}
		effectstest.FloatForTest(g, me, cost)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Torch Slinger", TypeLine: "Creature — Goblin Shaman", ManaCost: "{2}{R}", OracleID: oldB316SlingerOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: opts}); err != nil {
			t.Fatalf("kicked=%v cast: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if kicked {
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickCard(t, g, me.ID, victim)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(victim) {
				t.Error("kicked: the 2/2 victim survived 2 damage")
			}
		} else if effectstest.DamageOn(t, g, victim) != 0 {
			t.Error("unkicked: the Slinger dealt damage")
		}
	}
}

// --- Induce Despair --------------------------------------------------------

func TestOldB316InduceDespairShrinksByTheRevealedCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Giant", "", 4, 4)
	reveal := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: reveal, Name: "Revealed", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "CCB")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Induce Despair", TypeLine: "Instant", ManaCost: "{2}{B}", OracleID: oldB316DespairOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err == nil {
		t.Fatal("cast without naming a creature card to reveal")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(victim), RevealIDs: []uuid.UUID{reveal}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("the victim is %d/%d, want 1/1 (-3/-3)", p, tough)
	}
	if !me.Hand.Contains(reveal) {
		t.Error("the revealed card left the hand")
	}
}

func TestOldB316InduceDespairNeedsACreatureCardInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Giant", "", 4, 4)
	effectstest.FloatForTest(g, me, "CCB")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Induce Despair", TypeLine: "Instant", ManaCost: "{2}{B}", OracleID: oldB316DespairOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err == nil {
		t.Error("cast with no creature card in hand")
	}
}

// --- Bloodrite Invoker -----------------------------------------------------

func TestOldB316BloodriteInvokerDrainsThreeForEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	inv := effectstest.B12Push(g, me.ID, "Bloodrite Invoker", "Creature — Vampire Shaman", oldB316InvokerOracle, 3, 1)
	effectstest.FloatForTest(g, me, "CCCCCCC")
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("activated with only seven mana")
	}
	effectstest.FloatForTest(g, me, "C")
	effectstest.ApaActivate(t, g, me, inv, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 37 || me.Life != 43 {
		t.Errorf("opponent %d, me %d, want 37 and 43", opp.Life, me.Life)
	}
}

// --- Frontier Guide --------------------------------------------------------

func TestOldB316FrontierGuideFetchesATappedBasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	guide := effectstest.B12Push(g, me.ID, "Frontier Guide", "Creature — Elf Scout", oldB316GuideOracle, 1, 1)
	forest := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest"}
	nonbasic := game.Card{InstanceID: uuid.New(), Name: "Fancy Land", TypeLine: "Land"}
	effectstest.LibraryTop(me, forest, nonbasic)
	effectstest.FloatForTest(g, me, "CCCG")
	effectstest.ApaActivate(t, g, me, guide, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	if _, ok := effectstest.BattlefieldCard(g, forest.InstanceID); !ok {
		t.Fatal("the basic land did not enter the battlefield")
	}
	if !effectstest.TappedOnBattlefield(t, g, forest.InstanceID) {
		t.Error("the land entered untapped")
	}
	if !effectstest.TappedOnBattlefield(t, g, guide) {
		t.Error("the Guide did not tap for its cost")
	}
}

// --- Turntimber Grove ------------------------------------------------------

func TestOldB316TurntimberGroveEntersTappedAndPumpsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	grove := effectstest.PlayLandFromHand(t, g, "Turntimber Grove", oldB316GroveOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedOnBattlefield(t, g, grove) {
		t.Error("the Grove entered untapped")
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
	}
}

// --- Consume the Meek ------------------------------------------------------

func TestOldB316ConsumeTheMeekDestroysManaValueThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cheap := effectstest.B12Push(g, opp.ID, "Cheap", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cheap {
				g.Battlefield.Cards[i].ManaCost = "{3}"
			}
		}
	})
	dear := effectstest.B12Push(g, opp.ID, "Dear", "Creature — Giant", "", 4, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == dear {
				g.Battlefield.Cards[i].ManaCost = "{2}{G}{G}"
			}
		}
	})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Consume the Meek", TypeLine: "Instant", ManaCost: "{3}{B}{B}", OracleID: oldB316MeekOracle, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "CCCBB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("a mana value 3 creature survived")
	}
	if !g.Battlefield.Contains(dear) {
		t.Error("a mana value 4 creature was destroyed")
	}
}

// --- Marsh Casualties ------------------------------------------------------

func TestOldB316MarshCasualtiesShrinksOnePlayersCreatures(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		theirs := effectstest.B12Push(g, opp.ID, "Theirs", "Creature — Giant", "", 4, 4)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Giant", 4, 4)
		cost, opts, n := "BB", []int(nil), 1
		if kicked {
			cost, opts, n = "BBCCC", []int{0}, 2
		}
		effectstest.FloatForTest(g, me, cost)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Marsh Casualties", TypeLine: "Sorcery", ManaCost: "{B}{B}", OracleID: oldB316MarshOracle, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: opts,
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
			t.Fatalf("kicked=%v cast: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, theirs); p != 4-n || tough != 4-n {
			t.Errorf("kicked=%v: their creature is %d/%d, want %d/%d", kicked, p, tough, 4-n, 4-n)
		}
		if p, _ := effectstest.PtOf(t, g, mine); p != 4 {
			t.Errorf("kicked=%v: my creature shrank to %d", kicked, p)
		}
	}
}

// --- Hellion Eruption ------------------------------------------------------

func TestOldB316HellionEruptionTradesEachCreatureForAHellion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	effectstest.FloatForTest(g, me, "CCCCCR")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Hellion Eruption", TypeLine: "Sorcery", ManaCost: "{5}{R}", OracleID: oldB316EruptionOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, gone := range []uuid.UUID{a, b, c} {
		if g.Battlefield.Contains(gone) {
			t.Error("a creature of mine was not sacrificed")
		}
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("an opponent's creature was sacrificed")
	}
	n := 0
	for _, card := range g.Battlefield.Cards {
		if card.Name == "Hellion" && card.Controller == me.ID {
			n++
			if e := card.Effective(); e.Power != 4 || e.Toughness != 4 || !card.IsToken() || len(card.Colors) != 1 || card.Colors[0] != "R" {
				t.Errorf("Hellion %d/%d %v, want a red 4/4 token", e.Power, e.Toughness, card.Colors)
			}
		}
	}
	if n != 3 {
		t.Errorf("%d Hellions, want 3", n)
	}
}
