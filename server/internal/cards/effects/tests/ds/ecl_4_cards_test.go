package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ecl_4_cards_test.go — Lorwyn Eclipsed, fourth batch (#4458). Each test
// plays the real catalog entry and asserts what its text says.

const (
	ecl4WatcherOracle   = "ab5f8b8c-e3f7-4265-bbc6-814bbfd4693a"
	ecl4DarknessOracle  = "a3435757-d049-4aaf-a2cb-897dd4940997"
	ecl4ArcherOracle    = "1f2d1de6-42e0-4f45-b43d-ce4980146469"
	ecl4EulogistOracle  = "57bbd80b-290f-461e-8dc6-26ba4d673fa9"
	ecl4DuelistOracle   = "c892d28c-4a78-4a7f-b8d5-c3663b819c5e"
	ecl4WeaverOracle    = "28564d10-bfc5-4beb-b0e2-b9aa5c03f740"
	ecl4BoggartOracle   = "9cc3b91b-ad70-4b01-b718-fda3819f84d0"
	ecl4ElfOracle       = "7c89ef4d-b375-4644-ac8b-034effdbe98c"
	ecl4FlamekinOracle  = "48c641df-d511-4e35-b623-9d1fdb5240da"
	ecl4KithkinOracle   = "e6227840-6e3e-48c7-b4a4-42cfd8f14b68"
	ecl4MerfolkTypeLine = "Creature — Merfolk Soldier"
)

func TestEcl4EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		ecl4WatcherOracle:  "Crossroads Watcher",
		ecl4DarknessOracle: "Darkness Descends",
		ecl4ArcherOracle:   "Dawn's Light Archer",
		ecl4EulogistOracle: "Dawnhand Eulogist",
		ecl4DuelistOracle:  "Deepchannel Duelist",
		ecl4WeaverOracle:   "Dundoolin Weaver",
		ecl4BoggartOracle:  "Eclipsed Boggart",
		ecl4ElfOracle:      "Eclipsed Elf",
		ecl4FlamekinOracle: "Eclipsed Flamekin",
		ecl4KithkinOracle:  "Eclipsed Kithkin",
	}
	for oracle, name := range want {
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

// --- Crossroads Watcher ---------------------------------------------------

func TestEcl4CrossroadsWatcherGetsPlusOneForEachOtherCreatureThatEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	watcher := effectstest.RfdCast(t, g, "Crossroads Watcher", "Creature — Kithkin Ranger", ecl4WatcherOracle, 3, 3)
	if p, tt := effectstest.PtOf(t, g, watcher); p != 3 || tt != 3 {
		t.Fatalf("the Watcher entering does not trigger itself: %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, watcher, "trample") {
		t.Error("no trample")
	}
	effectstest.RfdCast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if p, tt := effectstest.PtOf(t, g, watcher); p != 4 || tt != 3 {
		t.Errorf("after one other creature entered: %d/%d, want 4/3", p, tt)
	}
	effectstest.RfdCast(t, g, "Second Bear", "Creature — Bear", "", 2, 2)
	if p, tt := effectstest.PtOf(t, g, watcher); p != 5 || tt != 3 {
		t.Errorf("after two other creatures entered: %d/%d, want 5/3", p, tt)
	}
}

// --- Darkness Descends ----------------------------------------------------

func TestEcl4DarknessDescendsPutsTwoCountersOnEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	giant := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Darkness Descends", "Sorcery", ecl4DarknessOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ZoneKindOf(g, mine); got != game.ZoneGraveyard {
		t.Errorf("your 2/2 is in %v, want the graveyard", got)
	}
	if got := effectstest.CountersOn(g, theirs, game.CounterMinusOne); got != 2 {
		t.Errorf("-1/-1 counters on the opponent's 3/3 = %d, want 2", got)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 1 || tt != 1 {
		t.Errorf("the opponent's 3/3 is %d/%d, want 1/1", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, giant); p != 3 || tt != 3 {
		t.Errorf("the 5/5 is %d/%d, want 3/3", p, tt)
	}
	if got := effectstest.CountersOn(g, rock, game.CounterMinusOne); got != 0 {
		t.Errorf("a noncreature artifact got %d counters, want none", got)
	}
}

// --- Dawn's Light Archer --------------------------------------------------

func TestEcl4DawnsLightArcherHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	plain := effectstest.RfdHandCard(me, "Bear", "Creature — Bear", "", "", 2, 2)
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{}); err == nil {
		t.Fatal("a creature without flash was cast in the beginning of combat step")
	}
	archer := effectstest.RfdHandCard(me, "Dawn's Light Archer", "Creature — Elf Archer", "", ecl4ArcherOracle, 4, 2)
	if err := g.CastSpell(me.ID, archer, game.CastSpellParams{}); err != nil {
		t.Fatalf("Dawn's Light Archer was not castable at instant speed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(archer) {
		t.Fatal("the Archer did not resolve")
	}
	if !effectstest.HasEffectiveKeyword(t, g, archer, "reach") {
		t.Error("no reach")
	}
}

func TestEcl4DawnsLightArcherBlocksAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	archer := effectstest.B12Push(g, opp.ID, "Dawn's Light Archer", "Creature — Elf Archer", ecl4ArcherOracle, 4, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(flyer, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, flyer); err == nil {
		t.Error("a creature without reach or flying blocked a flyer")
	}
	if err := g.DeclareBlocker(archer, flyer); err != nil {
		t.Errorf("the Archer could not block a flyer: %v", err)
	}
}

// --- Dawnhand Eulogist ----------------------------------------------------

func TestEcl4DawnhandEulogistDrainsOnlyWithAnElfCardInYourGraveyard(t *testing.T) {
	for _, tc := range []struct {
		name       string
		milledElf  bool
		graveElf   bool
		wantDrains bool
	}{
		{"an Elf among the milled cards", true, false, true},
		{"an Elf already in the graveyard", false, true, true},
		{"no Elf card", false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			second := game.Card{Name: "Bear", TypeLine: "Creature — Bear"}
			if tc.milledElf {
				second = game.Card{Name: "Milled Elf", TypeLine: "Creature — Elf Warrior"}
			}
			ids := effectstest.SamiLibraryTopFirst(me,
				game.Card{Name: "Shock", TypeLine: "Instant"},
				second,
				game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
				game.Card{Name: "Fourth", TypeLine: "Instant"},
			)
			if tc.graveElf {
				effectstest.Batch01GraveyardCard(me, "Dead Elf", "Creature — Elf")
			}
			lives := map[uuid.UUID]int{}
			for _, p := range g.Seats {
				lives[p.ID] = p.Life
			}
			eulogist := effectstest.RfdCast(t, g, "Dawnhand Eulogist", "Creature — Elf Warlock", ecl4EulogistOracle, 3, 3)
			effectstest.PassPriorityAroundTable(t, g)
			for _, id := range ids[:3] {
				if got := effectstest.ZoneKindOf(g, id); got != game.ZoneGraveyard {
					t.Errorf("a milled card is in %v, want the graveyard", got)
				}
			}
			if got := effectstest.ZoneKindOf(g, ids[3]); got != game.ZoneLibrary {
				t.Errorf("the fourth card is in %v; only three are milled", got)
			}
			for _, p := range g.Seats {
				want := lives[p.ID]
				switch {
				case !tc.wantDrains:
				case p.ID == me.ID:
					want += 2
				default:
					want -= 2
				}
				if p.Life != want {
					t.Errorf("seat %s life %d -> %d, want %d", p.Name, lives[p.ID], p.Life, want)
				}
			}
			if !effectstest.HasEffectiveKeyword(t, g, eulogist, "menace") {
				t.Error("no menace")
			}
		})
	}
}

// --- Deepchannel Duelist --------------------------------------------------

func TestEcl4DeepchannelDuelistPumpsOtherMerfolkAndUntapsOneAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	duelist := effectstest.B12Push(g, me.ID, "Deepchannel Duelist", ecl4MerfolkTypeLine, ecl4DuelistOracle, 2, 2)
	merfolk := effectstest.B12Creature(g, me.ID, "Merfolk", "Creature — Merfolk Wizard", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 2, 2)

	if p, tt := effectstest.PtOf(t, g, merfolk); p != 3 || tt != 3 {
		t.Errorf("your other Merfolk is %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, duelist); p != 2 || tt != 2 {
		t.Errorf("the Duelist pumps only OTHER Merfolk: %d/%d, want 2/2", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("a Bear is %d/%d, want 2/2", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("an opponent's Merfolk is %d/%d, want 2/2", p, tt)
	}

	effectstest.TapForTest(t, g, merfolk, bear, theirs)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at the end step")
	}
	if !effectstest.HasID(pick.PickTargetCards, merfolk) || !effectstest.HasID(pick.PickTargetCards, duelist) {
		t.Errorf("offered %v: want your Merfolk and the Duelist itself", pick.PickTargetCards)
	}
	if effectstest.HasID(pick.PickTargetCards, bear) || effectstest.HasID(pick.PickTargetCards, theirs) {
		t.Error("a non-Merfolk and an opponent's Merfolk are not legal targets")
	}
	effectstest.PickCard(t, g, me.ID, merfolk)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, merfolk) {
		t.Error("the target Merfolk did not untap")
	}
	if !effectstest.B16Tapped(t, g, bear) || !effectstest.B16Tapped(t, g, theirs) {
		t.Error("only the target untaps")
	}
}

// --- Dundoolin Weaver -----------------------------------------------------

func TestEcl4DundoolinWeaverReturnsAPermanentCardWithThreeCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Second Bear", "Creature — Bear", 2, 2)
	dead := effectstest.Batch01GraveyardCard(me, "Dead Ogre", "Creature — Ogre")
	spell := effectstest.Batch01GraveyardCard(me, "Old Shock", "Instant")
	effectstest.RfdCast(t, g, "Dundoolin Weaver", "Creature — Kithkin Druid", ecl4WeaverOracle, 2, 1)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("three creatures with the Weaver: no target prompt")
	}
	if !effectstest.HasID(pick.PickTargetCards, dead) || effectstest.HasID(pick.PickTargetCards, spell) {
		t.Fatalf("offered %v: want the creature card and not the instant", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the permanent card did not return to your hand")
	}
}

// With only two creatures (the Weaver and one other), the intervening
// "if" is false as it enters, so the ability never triggers.
func TestEcl4DundoolinWeaverDoesNotTriggerWithTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dead := effectstest.Batch01GraveyardCard(me, "Dead Ogre", "Creature — Ogre")
	weaver := effectstest.RfdCast(t, g, "Dundoolin Weaver", "Creature — Kithkin Druid", ecl4WeaverOracle, 2, 1)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TriggerOnStack(g, weaver) != nil {
		t.Error("the Weaver triggered with only two creatures")
	}
	if got := effectstest.ZoneKindOf(g, dead); got != game.ZoneGraveyard {
		t.Errorf("the creature card is in %v, want the graveyard", got)
	}
}

// --- The Eclipsed cycle ---------------------------------------------------

func TestEcl4EclipsedCycleTakesTheCreatureTypeOrEitherLandType(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		tribe, land1, land2    string
	}{
		{"Eclipsed Boggart", "Creature — Goblin Scout", ecl4BoggartOracle, 2, 3, "Goblin", "Swamp", "Mountain"},
		{"Eclipsed Elf", "Creature — Elf Scout", ecl4ElfOracle, 3, 2, "Elf", "Swamp", "Forest"},
		{"Eclipsed Flamekin", "Creature — Elemental Scout", ecl4FlamekinOracle, 1, 4, "Elemental", "Island", "Mountain"},
		{"Eclipsed Kithkin", "Creature — Kithkin Scout", ecl4KithkinOracle, 2, 1, "Kithkin", "Forest", "Plains"},
	} {
		for pickIdx, what := range []string{"creature", "first land", "second land"} {
			t.Run(tc.name+" takes the "+what, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				ids := effectstest.SamiLibraryTopFirst(me,
					game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
					game.Card{Name: "Kin", TypeLine: "Creature — " + tc.tribe + " Warrior"},
					game.Card{Name: "Land One", TypeLine: "Land — " + tc.land1},
					game.Card{Name: "Land Two", TypeLine: "Land — " + tc.land2},
					game.Card{Name: "Fifth", TypeLine: "Creature — " + tc.tribe},
				)
				lib := me.Library.Size()
				effectstest.RfdCast(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
				c := effectstest.ChooseCardsChoiceFor(g, me.ID)
				if c == nil {
					t.Fatal("no choose-cards prompt")
				}
				if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[0]}); err == nil {
					t.Fatal("a Bear matching none of the three types was taken")
				}
				want := ids[1+pickIdx]
				effectstest.AnswerChooseCards(t, g, me.ID, want)
				effectstest.PassPriorityAroundTable(t, g)
				if !me.Hand.Contains(want) {
					t.Fatalf("the %s did not reach the hand", what)
				}
				if me.Library.Size() != lib-1 {
					t.Errorf("library %d -> %d, want one card taken", lib, me.Library.Size())
				}
				if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[4] {
					t.Error("the fifth card must now be on top; the other three went to the bottom")
				}
			})
		}
	}
}
