package effects

import (
	"testing"

	"github.com/google/uuid"

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

// --- Crossroads Watcher ---------------------------------------------------

func TestEcl4CrossroadsWatcherGetsPlusOneForEachOtherCreatureThatEnters(t *testing.T) {
	g := newCatalogGame(t)
	watcher := rfdCast(t, g, "Crossroads Watcher", "Creature — Kithkin Ranger", ecl4WatcherOracle, 3, 3)
	if p, tt := ptOf(t, g, watcher); p != 3 || tt != 3 {
		t.Fatalf("the Watcher entering does not trigger itself: %d/%d, want 3/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, watcher, "trample") {
		t.Error("no trample")
	}
	rfdCast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if p, tt := ptOf(t, g, watcher); p != 4 || tt != 3 {
		t.Errorf("after one other creature entered: %d/%d, want 4/3", p, tt)
	}
	rfdCast(t, g, "Second Bear", "Creature — Bear", "", 2, 2)
	if p, tt := ptOf(t, g, watcher); p != 5 || tt != 3 {
		t.Errorf("after two other creatures entered: %d/%d, want 5/3", p, tt)
	}
}

// --- Darkness Descends ----------------------------------------------------

func TestEcl4DarknessDescendsPutsTwoCountersOnEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	castCatalogSpell(t, g, "Darkness Descends", "Sorcery", ecl4DarknessOracle, nil)
	passPriorityAroundTable(t, g)
	if got := zoneKindOf(g, mine); got != game.ZoneGraveyard {
		t.Errorf("your 2/2 is in %v, want the graveyard", got)
	}
	if got := countersOn(g, theirs, game.CounterMinusOne); got != 2 {
		t.Errorf("-1/-1 counters on the opponent's 3/3 = %d, want 2", got)
	}
	if p, tt := ptOf(t, g, theirs); p != 1 || tt != 1 {
		t.Errorf("the opponent's 3/3 is %d/%d, want 1/1", p, tt)
	}
	if p, tt := ptOf(t, g, giant); p != 3 || tt != 3 {
		t.Errorf("the 5/5 is %d/%d, want 3/3", p, tt)
	}
	if got := countersOn(g, rock, game.CounterMinusOne); got != 0 {
		t.Errorf("a noncreature artifact got %d counters, want none", got)
	}
}

// --- Dawn's Light Archer --------------------------------------------------

func TestEcl4DawnsLightArcherHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepBeginCombat)
	plain := rfdHandCard(me, "Bear", "Creature — Bear", "", "", 2, 2)
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{}); err == nil {
		t.Fatal("a creature without flash was cast in the beginning of combat step")
	}
	archer := rfdHandCard(me, "Dawn's Light Archer", "Creature — Elf Archer", "", ecl4ArcherOracle, 4, 2)
	if err := g.CastSpell(me.ID, archer, game.CastSpellParams{}); err != nil {
		t.Fatalf("Dawn's Light Archer was not castable at instant speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(archer) {
		t.Fatal("the Archer did not resolve")
	}
	if !hasEffectiveKeyword(t, g, archer, "reach") {
		t.Error("no reach")
	}
}

func TestEcl4DawnsLightArcherBlocksAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	archer := b12Push(g, opp.ID, "Dawn's Light Archer", "Creature — Elf Archer", ecl4ArcherOracle, 4, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(flyer, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			second := game.Card{Name: "Bear", TypeLine: "Creature — Bear"}
			if tc.milledElf {
				second = game.Card{Name: "Milled Elf", TypeLine: "Creature — Elf Warrior"}
			}
			ids := samiLibraryTopFirst(me,
				game.Card{Name: "Shock", TypeLine: "Instant"},
				second,
				game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
				game.Card{Name: "Fourth", TypeLine: "Instant"},
			)
			if tc.graveElf {
				batch01GraveyardCard(me, "Dead Elf", "Creature — Elf")
			}
			lives := map[uuid.UUID]int{}
			for _, p := range g.Seats {
				lives[p.ID] = p.Life
			}
			eulogist := rfdCast(t, g, "Dawnhand Eulogist", "Creature — Elf Warlock", ecl4EulogistOracle, 3, 3)
			passPriorityAroundTable(t, g)
			for _, id := range ids[:3] {
				if got := zoneKindOf(g, id); got != game.ZoneGraveyard {
					t.Errorf("a milled card is in %v, want the graveyard", got)
				}
			}
			if got := zoneKindOf(g, ids[3]); got != game.ZoneLibrary {
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
			if !hasEffectiveKeyword(t, g, eulogist, "menace") {
				t.Error("no menace")
			}
		})
	}
}

// --- Deepchannel Duelist --------------------------------------------------

func TestEcl4DeepchannelDuelistPumpsOtherMerfolkAndUntapsOneAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	duelist := b12Push(g, me.ID, "Deepchannel Duelist", ecl4MerfolkTypeLine, ecl4DuelistOracle, 2, 2)
	merfolk := b12Creature(g, me.ID, "Merfolk", "Creature — Merfolk Wizard", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 2, 2)

	if p, tt := ptOf(t, g, merfolk); p != 3 || tt != 3 {
		t.Errorf("your other Merfolk is %d/%d, want 3/3", p, tt)
	}
	if p, tt := ptOf(t, g, duelist); p != 2 || tt != 2 {
		t.Errorf("the Duelist pumps only OTHER Merfolk: %d/%d, want 2/2", p, tt)
	}
	if p, tt := ptOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("a Bear is %d/%d, want 2/2", p, tt)
	}
	if p, tt := ptOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("an opponent's Merfolk is %d/%d, want 2/2", p, tt)
	}

	tapForTest(t, g, merfolk, bear, theirs)
	advanceTo(t, g, game.StepEnd)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at the end step")
	}
	if !hasID(pick.PickTargetCards, merfolk) || !hasID(pick.PickTargetCards, duelist) {
		t.Errorf("offered %v: want your Merfolk and the Duelist itself", pick.PickTargetCards)
	}
	if hasID(pick.PickTargetCards, bear) || hasID(pick.PickTargetCards, theirs) {
		t.Error("a non-Merfolk and an opponent's Merfolk are not legal targets")
	}
	pickCard(t, g, me.ID, merfolk)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, merfolk) {
		t.Error("the target Merfolk did not untap")
	}
	if !b16Tapped(t, g, bear) || !b16Tapped(t, g, theirs) {
		t.Error("only the target untaps")
	}
}

// --- Dundoolin Weaver -----------------------------------------------------

func TestEcl4DundoolinWeaverReturnsAPermanentCardWithThreeCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Second Bear", "Creature — Bear", 2, 2)
	dead := batch01GraveyardCard(me, "Dead Ogre", "Creature — Ogre")
	spell := batch01GraveyardCard(me, "Old Shock", "Instant")
	rfdCast(t, g, "Dundoolin Weaver", "Creature — Kithkin Druid", ecl4WeaverOracle, 2, 1)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("three creatures with the Weaver: no target prompt")
	}
	if !hasID(pick.PickTargetCards, dead) || hasID(pick.PickTargetCards, spell) {
		t.Fatalf("offered %v: want the creature card and not the instant", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the permanent card did not return to your hand")
	}
}

// With only two creatures (the Weaver and one other), the intervening
// "if" is false as it enters, so the ability never triggers.
func TestEcl4DundoolinWeaverDoesNotTriggerWithTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dead := batch01GraveyardCard(me, "Dead Ogre", "Creature — Ogre")
	weaver := rfdCast(t, g, "Dundoolin Weaver", "Creature — Kithkin Druid", ecl4WeaverOracle, 2, 1)
	if latestPickTarget(g, me.ID) != nil || triggerOnStack(g, weaver) != nil {
		t.Error("the Weaver triggered with only two creatures")
	}
	if got := zoneKindOf(g, dead); got != game.ZoneGraveyard {
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
				g := newCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				ids := samiLibraryTopFirst(me,
					game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
					game.Card{Name: "Kin", TypeLine: "Creature — " + tc.tribe + " Warrior"},
					game.Card{Name: "Land One", TypeLine: "Land — " + tc.land1},
					game.Card{Name: "Land Two", TypeLine: "Land — " + tc.land2},
					game.Card{Name: "Fifth", TypeLine: "Creature — " + tc.tribe},
				)
				lib := me.Library.Size()
				rfdCast(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
				c := chooseCardsChoiceFor(g, me.ID)
				if c == nil {
					t.Fatal("no choose-cards prompt")
				}
				if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[0]}); err == nil {
					t.Fatal("a Bear matching none of the three types was taken")
				}
				want := ids[1+pickIdx]
				answerChooseCards(t, g, me.ID, want)
				passPriorityAroundTable(t, g)
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
