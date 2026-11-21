package testso

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_mp3_cards_test.go — Outlaws of Thunder Junction (OTJ), the M–P
// leftovers (#4468): the cards the first two M–P batches did not take.
// Each test plays the real catalog entry through the engine and asserts
// what the printed text says.

const (
	otjMP3MagebaneLizardOracle  = "0c38f159-367b-4e4b-94cd-bf1b0721c3a8"
	otjMP3MakeYourOwnLuckOracle = "41246140-4b44-40c4-b92b-501e49c50bf7"
	otjMP3MetamorphicBlast      = "69c8e690-c0a2-4c4e-910d-123c9c5b922f"
	otjMP3MiriamOracle          = "1d88014b-8189-4a88-bee3-c2310d36f6e8"
	otjMP3MobileHomestead       = "b71239fb-d7dc-4bac-b8aa-26b454409344"
	otjMP3OutlawStitcherOracle  = "9fda1239-97ca-41f0-8bfa-040484769396"
)

func TestOtjMP3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		otjMP3MagebaneLizardOracle:  "Magebane Lizard",
		otjMP3MakeYourOwnLuckOracle: "Make Your Own Luck",
		otjMP3MetamorphicBlast:      "Metamorphic Blast",
		otjMP3MiriamOracle:          "Miriam, Herd Whisperer",
		otjMP3MobileHomestead:       "Mobile Homestead",
		otjMP3OutlawStitcherOracle:  "Outlaw Stitcher",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	if sa := effectstest.PlotOf(otjMP3OutlawStitcherOracle); sa == nil || sa.Cost != "{4}{U}" || sa.CastCost != "" {
		t.Errorf("Outlaw Stitcher: plot %+v, want {4}{U} and a free cast", sa)
	}
}

// --- Magebane Lizard -------------------------------------------------------------

func TestOtjMP3MagebaneLizardCountsEachCastersNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Magebane Lizard", "Creature — Lizard", otjMP3MagebaneLizardOracle, false)
	effectstest.AdvanceToMain(t, g)

	life := me.Life
	effectstest.B24CastNoncreature(t, g, "First", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Fatalf("your first noncreature spell: life %d → %d, want 1 damage", life, me.Life)
	}
	effectstest.B24CastNoncreature(t, g, "Second", "Sorcery", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Fatalf("your second noncreature spell: life %d → %d, want 2 more", life, me.Life)
	}
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Power: 2, Toughness: 2})
	if me.Life != life-3 {
		t.Errorf("a creature spell triggered it: life %d", me.Life)
	}

	theirs := opp.Life
	effectstest.Batch01OpponentCasts(t, g, opp, "Their Bolt", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[2].ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-1 {
		t.Errorf("an opponent's first noncreature spell: life %d → %d, want 1 damage", theirs, opp.Life)
	}
}

// --- Make Your Own Luck ----------------------------------------------------------

func TestOtjMP3MakeYourOwnLuckPlotsOneAndTakesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.TopLibraryCard(me, "Forest", "Basic Land — Forest")
	spell := effectstest.TopLibraryCard(me, "Some Sorcery", "Sorcery")
	other := effectstest.TopLibraryCard(me, "Some Instant", "Instant")
	lib := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Make Your Own Luck", "Sorcery", otjMP3MakeYourOwnLuckOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick")
	}
	if len(pick.ChooseCards) != 2 || effectstest.HasID(pick.ChooseCards, land) || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("up to one of the two nonland cards: %v in [%d,%d]", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) {
		t.Fatal("the chosen card is exiled")
	}
	plotted := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventBecomesPlotted && ev.CardID == spell {
			plotted = true
		}
	}
	if !plotted {
		t.Error("the exiled card becomes plotted")
	}
	if !effectstest.HandHasID(me, land) || !effectstest.HandHasID(me, other) {
		t.Error("the rest go to your hand")
	}
	if me.Library.Size() != lib-3 {
		t.Errorf("library %d → %d, want the top three gone", lib, me.Library.Size())
	}
}

func TestOtjMP3MakeYourOwnLuckExilingNothingTakesAllThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	effectstest.CastCatalogSpell(t, g, "Make Your Own Luck", "Sorcery", otjMP3MakeYourOwnLuckOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:3] {
		if !effectstest.HandHasID(me, id) {
			t.Errorf("%s: all three go to your hand", id)
		}
	}
	if g.Exile.Size() != 0 {
		t.Error("nothing was exiled")
	}
}

// --- Metamorphic Blast -----------------------------------------------------------

func TestOtjMP3MetamorphicBlastMakesAWhiteRabbitUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	dragon := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shivan Dragon", TypeLine: "Creature — Dragon", Colors: []string{"R"},
		Power: 5, Toughness: 5, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastModal(t, g, "Metamorphic Blast", "Instant", otjMP3MetamorphicBlast,
		[]int{0}, []game.TargetRef{effectstest.ModeRef(game.TargetCard, dragon, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	d := effectstest.EffectiveCardForTest(t, g, dragon)
	if ch := d.Effective(); !reflect.DeepEqual(ch.Subtypes, []string{"Rabbit"}) || !reflect.DeepEqual(ch.Colors, []string{"W"}) {
		t.Errorf("the Dragon is %v %v, want a white Rabbit", ch.Colors, ch.Subtypes)
	}
	if p, tt := effectstest.PtOf(t, g, dragon); p != 0 || tt != 1 {
		t.Errorf("the Dragon is %d/%d, want 0/1", p, tt)
	}
	if !effectstest.EffectiveHasKeyword(&d, "flying") {
		t.Error("it keeps its abilities")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if p, tt := effectstest.PtOf(t, g, dragon); p != 5 || tt != 5 {
		t.Errorf("until end of turn only: %d/%d next turn", p, tt)
	}
}

func TestOtjMP3MetamorphicBlastMakesTargetPlayerDrawTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.SeedLibrary(opp, "One", "Two", "Three")
	hand := opp.Hand.Size()
	effectstest.CastModal(t, g, "Metamorphic Blast", "Instant", otjMP3MetamorphicBlast,
		[]int{1}, []game.TargetRef{effectstest.ModeRef(game.TargetPlayer, opp.ID, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+2 {
		t.Errorf("target player draws two: hand %d → %d", hand, opp.Hand.Size())
	}
}

// --- Miriam, Herd Whisperer ------------------------------------------------------

func TestOtjMP3MiriamGivesHexproofOnYourTurnAndCountersAttackingMounts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Miriam, Herd Whisperer", "Legendary Creature — Human Druid", otjMP3MiriamOracle, false)
	mount := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Horse",
		TypeLine: "Creature — Horse Mount", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	vehicle := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Wagon",
		TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Horse",
		TypeLine: "Creature — Horse Mount", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	for _, id := range []uuid.UUID{mount, vehicle} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "hexproof") {
			t.Errorf("%s: your Mount or Vehicle has hexproof on your turn", id)
		}
	}
	for _, id := range []uuid.UUID{bear, theirs} {
		if effectstest.HasEffectiveKeyword(t, g, id, "hexproof") {
			t.Errorf("%s: not your Mount or Vehicle", id)
		}
	}

	effectstest.DeclareAttack(t, g, opp.ID, mount, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, mount, game.CounterPlusOne); got != 1 {
		t.Errorf("the attacking Mount has %d +1/+1 counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("an attacking Bear got %d counters", got)
	}

	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, mount, "hexproof") {
		t.Error("only during your turn")
	}
}

// --- Mobile Homestead ------------------------------------------------------------

func TestOtjMP3MobileHomesteadHasHasteWithAMountAndFindsALandWhenItAttacks(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	home := s.PushVehicle("Mobile Homestead", otjMP3MobileHomestead, 3, 3)
	pilot := s.PushCreature("Pilot", "Creature — Human", 2, 2)
	if effectstest.HasEffectiveKeyword(t, s.G, home, "haste") {
		t.Fatal("no Mount: no haste")
	}
	s.PushCreature("My Horse", "Creature — Horse Mount", 1, 1)
	if !effectstest.HasEffectiveKeyword(t, s.G, home, "haste") {
		t.Fatal("you control a Mount: haste")
	}
	s.CrewAndResolve(t, home, 0, pilot)
	land := effectstest.PlTop(s.Me, "Forest", "Basic Land — Forest", "")
	effectstest.DeclareAttack(t, s.G, s.Opp.ID, home)
	effectstest.PassPriorityAroundTable(t, s.G)
	effectstest.AnswerChooseCards(t, s.G, s.Me.ID, land)
	if !s.G.Battlefield.Contains(land) || !effectstest.TappedOnBattlefield(t, s.G, land) {
		t.Error("the land enters tapped")
	}
}

func TestOtjMP3MobileHomesteadLeavesANonlandOnTop(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	home := s.PushVehicle("Mobile Homestead", otjMP3MobileHomestead, 3, 3)
	pilot := s.PushCreature("Pilot", "Creature — Human", 2, 2)
	s.CrewAndResolve(t, home, 0, pilot)
	top := effectstest.PlTop(s.Me, "Bolt", "Instant", "{R}")
	effectstest.DeclareAttack(t, s.G, s.Opp.ID, home)
	effectstest.PassPriorityAroundTable(t, s.G)
	if effectstest.LatestChooseCardsFor(s.G, s.Me.ID) != nil {
		t.Fatal("a nonland card offers nothing")
	}
	if got := s.Me.Library.Cards[len(s.Me.Library.Cards)-1].InstanceID; got != top {
		t.Error("the nonland card stays on top")
	}
}

// --- Outlaw Stitcher -------------------------------------------------------------

func otjMP3ZombieRogue(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie Rogue" && c.Controller == owner {
			return c.InstanceID
		}
	}
	t.Fatal("no Zombie Rogue token")
	return uuid.Nil
}

func TestOtjMP3OutlawStitcherCountersForEachSpellAfterTheFirst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B24CastNoncreature(t, g, "First", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B24CastNoncreature(t, g, "Second", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Outlaw Stitcher", TypeLine: "Creature — Human Warlock",
		ManaCost: "{3}{U}", OracleID: otjMP3OutlawStitcherOracle, Power: 1, Toughness: 4})
	zombie := otjMP3ZombieRogue(t, g, me.ID)
	if got := effectstest.CountersOn(g, zombie, game.CounterPlusOne); got != 4 {
		t.Errorf("the Stitcher was your third spell: %d counters, want 4", got)
	}
	if p, tt := effectstest.PtOf(t, g, zombie); p != 6 || tt != 6 {
		t.Errorf("the Zombie Rogue is %d/%d, want 6/6", p, tt)
	}
}

func TestOtjMP3OutlawStitcherAsYourFirstSpellMakesAPlainToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Outlaw Stitcher", TypeLine: "Creature — Human Warlock",
		ManaCost: "{3}{U}", OracleID: otjMP3OutlawStitcherOracle, Power: 1, Toughness: 4})
	zombie := otjMP3ZombieRogue(t, g, me.ID)
	if got := effectstest.CountersOn(g, zombie, game.CounterPlusOne); got != 0 {
		t.Errorf("no spell other than the first: %d counters", got)
	}
	if p, tt := effectstest.PtOf(t, g, zombie); p != 2 || tt != 2 {
		t.Errorf("the Zombie Rogue is %d/%d, want 2/2", p, tt)
	}
}
