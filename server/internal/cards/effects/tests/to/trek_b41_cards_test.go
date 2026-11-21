package teststo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b41_cards_test.go — Star Trek Commander, batch 41 (#4974). Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB41TPringOracle   = "29775eb6-f658-4839-a434-2fa063069f1c"
	trekB41UnifyOracle    = "b0efb6fb-f2d1-4027-90b5-44fb3f5dd538"
	trekB41VicOracle      = "1d6f4573-d2f6-4771-a748-291846ef14b9"
	trekB41WorfOracle     = "a2249af3-2e3e-484b-a31a-26432edd13ae"
	trekB41CubeOracle     = "9c14e21b-a733-4882-8e09-6dbfc85ac3aa"
	trekB41ChairOracle    = "7132b0ad-7bab-4379-ad82-a9af289f835e"
	trekB41GuardianOracle = "5144361a-0308-49b8-96e6-c948d4beae09"
	trekB41NeghVarOracle  = "373e2b58-ccf0-444b-aeb2-0b6e8d4e80e7"
	trekB41TubeOracle     = "2ca270f1-a8a2-438f-9321-2c653165f3e4"
	trekB41KtingaOracle   = "8ea8ee52-30e1-415c-af52-66b209f05f0a"
)

func TestTrekB41EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB41TPringOracle:   "T'Pring, Never Parted",
		trekB41UnifyOracle:    "Unification",
		trekB41VicOracle:      "Vic Fontaine, One Night Only",
		trekB41WorfOracle:     "Worf, Son of Mogh",
		trekB41CubeOracle:     "Class-4 Tactical Cube",
		trekB41ChairOracle:    "Command Chair",
		trekB41GuardianOracle: "The Guardian of Forever",
		trekB41NeghVarOracle:  "I.K.S. Negh'Var, Imperial Flagship",
		trekB41TubeOracle:     "Jefferies Tube",
		trekB41KtingaOracle:   "K't'inga-Class Battle Cruiser",
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

// --- fixtures -------------------------------------------------------------------

// trekB41Draw has p draw one card through the engine and settles the
// stack, answering no prompt.
func trekB41Draw(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, 1) })
}

// trekB41Tokens lists the tokens named `name` that p controls.
func trekB41Tokens(g *game.Game, p uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p && c.Name == name && effects.IsToken(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func trekB41Card(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// trekB41CastTPring casts T'Pring and resolves the creature spell, leaving
// the enters trigger waiting on its two targets.
func trekB41CastTPring(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := effectstest.CastFromHand(t, g, game.Card{Name: "T'Pring, Never Parted", TypeLine: "Legendary Creature — Vulcan Citizen",
		OracleID: trekB41TPringOracle, Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// --- T'Pring, Never Parted ----------------------------------------------------

func TestTrekB41TPringFightsAndDrawsWhenAFighterDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tpring := trekB41CastTPring(t, g)
	// "Another target creature you control": not T'Pring, not theirs.
	for _, bad := range []uuid.UUID{tpring, theirs} {
		if err := effectstest.SgAnswerTargets(t, g, bad); !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("%s as the first target: err = %v, want ErrIllegalTarget", effects.CardNameFor(g, bad), err)
		}
	}
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PickCard(t, g, me.ID, theirs)
	hand := me.Hand.Size()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the 2/2 survived the fight with a 4/4")
	}
	if c := effectstest.LayeredCard(t, g, mine); c.DamageMarked != 2 {
		t.Errorf("my creature took %d damage, want 2", c.DamageMarked)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("drew %d when the creature that fought died, want 1", got)
	}
	// The survivor that fought dies later this turn: one more card.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d in all after my fighter died too, want 2", got)
	}
}

func TestTrekB41TPringDeclinedFightsNothingAndDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	trekB41CastTPring(t, g)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PickCard(t, g, me.ID, theirs)
	hand := me.Hand.Size()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) || effectstest.LayeredCard(t, g, mine).DamageMarked != 0 {
		t.Fatal("declining the fight still dealt damage")
	}
	// A creature that did not fight dying draws nothing.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d with no fight, want 0", got)
	}
}

// --- Unification -------------------------------------------------------------------

func trekB41Unification(g *game.Game, owner uuid.UUID, lore int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Unification", TypeLine: "Enchantment — Saga",
		OracleID: trekB41UnifyOracle, ManaCost: "{3}{W}{U}", Colors: []string{"W", "U"},
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLore: lore},
	})
}

func TestTrekB41UnificationChapterOneOwnerDrawsTwoThenShufflesItIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	trekB41Unification(g, me.ID, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	myHand, theirHand, theirLib := me.Hand.Size(), opp.Hand.Size(), opp.Library.Size()
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - theirHand; got != 2 {
		t.Errorf("the owner drew %d, want 2", got)
	}
	if me.Hand.Size() != myHand {
		t.Error("the Saga's controller drew for an opponent's creature")
	}
	if g.Battlefield.Contains(theirs) || !opp.Library.Contains(theirs) {
		t.Fatal("the creature was not shuffled into its owner's library")
	}
	if got := opp.Library.Size(); got != theirLib-2+1 {
		t.Errorf("owner's library is %d, want %d", got, theirLib-1)
	}
}

func TestTrekB41UnificationChapterTwoPutsASpacecraftFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	trekB41Unification(g, me.ID, 1)
	ship := effectstest.Batch01GraveyardCard(me, "Old Ship", "Artifact — Spacecraft")
	bear := effectstest.Batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	inHand := trekB23Hand(me, "Hand Ship", "Artifact — Spacecraft")
	vehicle := trekB23Hand(me, "Hand Car", "Artifact — Vehicle")
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.LatestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 1 {
		t.Fatalf("want an optional pick of one card: %+v", ch)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range ch.ChooseCards {
		offered[id] = true
	}
	if !offered[ship] || !offered[inHand] || offered[bear] || offered[vehicle] {
		t.Fatalf("offered %v; want both Spacecraft cards and nothing else", ch.ChooseCards)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{ship}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ship) {
		t.Error("the Spacecraft from the graveyard is not on the battlefield")
	}
	if !me.Hand.Contains(inHand) {
		t.Error("the Spacecraft in hand moved too")
	}
}

func TestTrekB41UnificationChapterTwoPutsASpacecraftFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	trekB41Unification(g, me.ID, 1)
	inHand := trekB23Hand(me, "Hand Ship", "Artifact — Spacecraft")
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.LatestChooseCards(g, me.ID)
	if ch == nil {
		t.Fatal("no pick was offered")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{inHand}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(inHand) || me.Hand.Contains(inHand) {
		t.Error("the Spacecraft from hand is not on the battlefield")
	}
}

func TestTrekB41UnificationChapterThreeBothDrawTwoAndItIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saga := trekB41Unification(g, me.ID, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.SgAnswerTargets(t, g, me.ID); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("yourself as the target opponent: err = %v, want ErrIllegalTarget", err)
	}
	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for chapter III")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size()-myHand != 2 || opp.Hand.Size()-theirHand != 2 {
		t.Errorf("drew %d and %d, want 2 each", me.Hand.Size()-myHand, opp.Hand.Size()-theirHand)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

// --- Vic Fontaine, One Night Only ----------------------------------------------------

func TestTrekB41VicFontainePumpsTwoCreaturesOnYourSecondDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	effectstest.B27Push(g, me.ID, "Vic Fontaine, One Night Only", "Legendary Artifact Creature — Illusion Performer", trekB41VicOracle, "{1}{R}{W}", 3, 3, "R", "W")
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, other.ID, "Bear B", "Creature — Bear", 2, 2)
	trekB41Draw(g, other)
	trekB41Draw(g, other)
	trekB41Draw(g, me)
	if effectstest.LatestPickTarget(g, me.ID) != nil || !effectstest.StackFullyEmpty(g) {
		t.Fatal("triggered before my second draw")
	}
	trekB41Draw(g, me)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the second draw did not trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{trekB41Card(a), trekB41Card(b)}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c := effectstest.LayeredCard(t, g, id)
		if c.CurrentPower() != 4 || c.CurrentToughness() != 2 || !game.HasKeyword(&c, "lifelink") {
			t.Errorf("%s is %d/%d lifelink=%v, want 4/2 with lifelink", c.Name, c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "lifelink"))
		}
	}
	trekB41Draw(g, me)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the third draw triggered again")
	}
}

// --- Worf, Son of Mogh -------------------------------------------------------------

func TestTrekB41WorfPumpsOtherOfficersAndMakesOneOnTheSecondDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	worf := effectstest.B27Push(g, me.ID, "Worf, Son of Mogh", "Legendary Creature — Klingon Warrior Officer", trekB41WorfOracle, "{1}{R}{W}", 3, 3, "R", "W")
	officer := effectstest.B12Creature(g, me.ID, "Ensign", "Creature — Human Officer", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Officer", "Creature — Human Officer", 1, 1)
	if c := effectstest.LayeredCard(t, g, worf); c.CurrentPower() != 3 || !game.HasKeyword(&c, "menace") {
		t.Errorf("Worf is %d power, menace=%v; the anthem is for other Officers", c.CurrentPower(), game.HasKeyword(&c, "menace"))
	}
	if c := effectstest.LayeredCard(t, g, officer); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("my Officer is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	for _, id := range []uuid.UUID{bear, theirs} {
		if c := effectstest.LayeredCard(t, g, id); c.CurrentPower() != c.Power {
			t.Errorf("%s was pumped", c.Name)
		}
	}
	trekB41Draw(g, me)
	trekB41Draw(g, me)
	effectstest.PassPriorityAroundTable(t, g)
	toks := trekB41Tokens(g, me.ID, "Officer")
	if len(toks) != 1 {
		t.Fatalf("%d Officer tokens after the second draw, want 1", len(toks))
	}
	c := effectstest.LayeredCard(t, g, toks[0])
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !c.HasColor("R") {
		t.Errorf("the token is %d/%d red=%v, want a red 1/1 pumped to 2/2", c.CurrentPower(), c.CurrentToughness(), c.HasColor("R"))
	}
	trekB41Draw(g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(trekB41Tokens(g, me.ID, "Officer")); got != 1 {
		t.Errorf("%d tokens after a third draw, want still 1", got)
	}
}

// --- Class-4 Tactical Cube ----------------------------------------------------------

func TestTrekB41TacticalCubeMakesTwoBorgAndLordsArtifactCreaturesAtSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cube := effectstest.CastFromHand(t, g, game.Card{Name: "Class-4 Tactical Cube", TypeLine: "Artifact — Spacecraft",
		OracleID: trekB41CubeOracle, ManaCost: "{4}", Power: 4, Toughness: 4}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	borg := trekB41Tokens(g, me.ID, "Borg")
	if len(borg) != 2 {
		t.Fatalf("%d Borg tokens, want 2", len(borg))
	}
	if c := effectstest.LayeredCard(t, g, borg[0]); !c.IsArtifact() || !c.IsCreature() || c.CurrentPower() != 1 || len(c.Colors) != 0 {
		t.Errorf("Borg is artifact=%v creature=%v power=%d colors=%v; want a colorless 1/1 artifact creature", c.IsArtifact(), c.IsCreature(), c.CurrentPower(), c.Colors)
	}
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 6, 6)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.Station(t, g, me, cube, golem)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LayeredCard(t, g, cube).IsCreature() || effectstest.LayeredCard(t, g, borg[0]).CurrentPower() != 1 {
		t.Fatal("at six charge counters the Cube is a creature or the anthem is on")
	}
	effectstest.Station(t, g, me, cube, borg[1])
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, cube)
	if !c.IsCreature() || !game.HasKeyword(&c, "flying") || c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("at 7+ the Cube is creature=%v flying=%v %d/%d; want a flying 4/4 pumped to 5/5", c.IsCreature(), game.HasKeyword(&c, "flying"), c.CurrentPower(), c.CurrentToughness())
	}
	if p := effectstest.LayeredCard(t, g, borg[0]).CurrentPower(); p != 2 {
		t.Errorf("Borg power %d at 7+, want 2", p)
	}
	if p := effectstest.LayeredCard(t, g, golem).CurrentPower(); p != 7 {
		t.Errorf("artifact Golem power %d at 7+, want 7", p)
	}
	if p := effectstest.LayeredCard(t, g, bear).CurrentPower(); p != 2 {
		t.Errorf("a non-artifact creature was pumped to %d", p)
	}
}

// --- Command Chair ----------------------------------------------------------------

func TestTrekB41CommandChairEntersTappedAndTapsForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	chair := effectstest.CastFromHand(t, g, game.Card{Name: "Command Chair", TypeLine: "Artifact", OracleID: trekB41ChairOracle, ManaCost: "{6}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, chair); !ok || !c.Tapped {
		t.Fatalf("Command Chair on the battlefield %v tapped %v; it enters tapped", ok, c.Tapped)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == chair {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, chair, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if me.ManaPool.Total() != 3 {
		t.Errorf("pool %d, want three mana of one color", me.ManaPool.Total())
	}
	for _, u := range me.ManaPool.Units() {
		if u.Color != "U" {
			t.Errorf("a %s unit in the pool, want only U", u.Color)
		}
	}
}

func TestTrekB41CommandChairTapsForYourGreatestCommanderManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chair := effectstest.PushCatalogPermanent(g, me.ID, "Command Chair", "Artifact", trekB41ChairOracle, false)
	small := game.Card{InstanceID: uuid.New(), Name: "Small Commander", TypeLine: "Legendary Creature — Human",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true}
	big := game.Card{InstanceID: uuid.New(), Name: "Big Commander", TypeLine: "Legendary Creature — Human",
		ManaCost: "{3}{G}{G}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID, IsCommander: true}
	theirs := game.Card{InstanceID: uuid.New(), Name: "Their Commander", TypeLine: "Legendary Creature — Human",
		ManaCost: "{5}{B}{B}", Power: 7, Toughness: 7, Owner: opp.ID, Controller: opp.ID, IsCommander: true}
	for _, c := range []game.Card{small, big, theirs} {
		effectstest.PushBattlefieldCardWithTimestamp(g, c)
	}
	if err := g.ActivateManaAbility(me.ID, chair, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if me.ManaPool.Total() != 5 {
		t.Errorf("pool %d, want 5 (the greatest mana value among MY commanders)", me.ManaPool.Total())
	}
}

func TestTrekB41CommandChairWithNoCommanderAddsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chair := effectstest.PushCatalogPermanent(g, me.ID, "Command Chair", "Artifact", trekB41ChairOracle, false)
	_ = g.ActivateManaAbility(me.ID, chair, 1, game.ManaAbilityParams{})
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick != nil {
		_ = g.ResolveManaChoice(pick.ID, me.ID, "G")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool %d with no commander, want 0", me.ManaPool.Total())
	}
}

// --- The Guardian of Forever ----------------------------------------------------------

func TestTrekB41GuardianReturnsTheSacrificedCardAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	guardian := effectstest.PushCatalogPermanent(g, me.ID, "The Guardian of Forever", "Legendary Artifact Creature — Construct", trekB41GuardianOracle, false)
	if c := effectstest.LayeredCard(t, g, guardian); !game.HasKeyword(&c, "defender") {
		t.Error("the Guardian has no defender")
	}
	// A creature I control but an opponent owns.
	borrowed := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Borrowed Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, guardian, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{guardian}, Strict: true}); err == nil {
		t.Fatal("the Guardian sacrificed itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, guardian, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{borrowed}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !opp.Graveyard.Contains(borrowed) {
		t.Fatal("the sacrificed creature is not in its owner's graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(borrowed) {
		t.Fatal("the card returned before the end step")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, borrowed)
	if !ok {
		t.Fatal("the card did not return at the end step")
	}
	if c.Controller != opp.ID {
		t.Errorf("it returned under %v, want its owner %v", c.Controller, opp.ID)
	}
}

func TestTrekB41GuardianCantSacrificeAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	guardian := effectstest.PushCatalogPermanent(g, me.ID, "The Guardian of Forever", "Legendary Artifact Creature — Construct", trekB41GuardianOracle, false)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		_ = effects.CreateToken{Controller: me.ID, Template: effects.TokenCard("1/1 red Officer"), N: 1}.Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	toks := trekB41Tokens(g, me.ID, "Officer")
	if len(toks) != 1 {
		t.Fatalf("seeded %d tokens", len(toks))
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, guardian, 0, game.ActivateAbilityParams{SacrificeIDs: toks, Strict: true}); err == nil {
		t.Fatal("sacrificed a token to the Guardian")
	}
}

// --- I.K.S. Negh'Var, Imperial Flagship --------------------------------------------------

func TestTrekB41NeghVarGivesYourKlingonsAndWarriorsVigilanceAndStationsAtTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ship := effectstest.PushCatalogPermanent(g, me.ID, "I.K.S. Negh'Var, Imperial Flagship", "Legendary Artifact — Spacecraft", trekB41NeghVarOracle, false)
	klingon := effectstest.B12Creature(g, me.ID, "Klingon", "Creature — Klingon", 6, 6)
	warrior := effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Klingon", "Creature — Klingon", 2, 2)
	for id, want := range map[uuid.UUID]bool{klingon: true, warrior: true, bear: false, theirs: false} {
		if got := effectstest.EffectiveAbilitiesContain(t, g, id, "vigilance"); got != want {
			t.Errorf("%s vigilance = %v, want %v", effects.CardNameFor(g, id), got, want)
		}
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.Station(t, g, me, ship, klingon)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LayeredCard(t, g, ship).IsCreature() {
		t.Fatal("a creature at six charge counters")
	}
	effectstest.Station(t, g, me, ship, warrior)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 6 || c.CurrentToughness() != 6 || !game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "vigilance") {
		t.Errorf("at 10+ Negh'Var is creature=%v %d/%d flying=%v vigilance=%v; want a 6/6 flier with vigilance",
			c.IsCreature(), c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"), game.HasKeyword(&c, "vigilance"))
	}
}

// --- Jefferies Tube ---------------------------------------------------------------------

func TestTrekB41JefferiesTubeChargesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tube := effectstest.PushCatalogPermanent(g, me.ID, "Jefferies Tube", "Artifact", trekB41TubeOracle, false)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tube, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{trekB41Card(bear)}, Strict: true}); err == nil {
		t.Fatal("put charge counters on a non-artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, tube, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{trekB41Card(rock)}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, rock, game.CounterCharge); got != 2 {
		t.Errorf("charge counters %d, want 2", got)
	}
	if !effectstest.LayeredCard(t, g, tube).Tapped {
		t.Error("the {T} cost did not tap the Tube")
	}
}

func TestTrekB41JefferiesTubeMakesYourCreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tube := effectstest.PushCatalogPermanent(g, me.ID, "Jefferies Tube", "Artifact", trekB41TubeOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tube, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{trekB41Card(theirs)}, Strict: true}); err == nil {
		t.Fatal("targeted an opponent's creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, tube, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{trekB41Card(mine)}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraRestrictions(t, g, mine)&game.CantBeBlocked == 0 {
		t.Error("the creature can still be blocked")
	}
}

func TestTrekB41JefferiesTubeDrawsThenPutsALandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tube := effectstest.PushCatalogPermanent(g, me.ID, "Jefferies Tube", "Artifact", trekB41TubeOracle, false)
	effectstest.AdvanceToMain(t, g)
	lib := trekB23Library(me, game.Card{Name: "Drawn Forest", TypeLine: "Basic Land — Forest"})
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tube, 2, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated the {4} ability for three mana")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, tube, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.LatestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 {
		t.Fatalf("want an optional land pick: %+v", ch)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[0]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if c, ok := effectstest.BattlefieldCard(g, lib[0]); !ok || !c.Tapped {
		t.Errorf("the drawn land is on the battlefield %v, tapped %v; want both", ok, c.Tapped)
	}
}

// --- K't'inga-Class Battle Cruiser ---------------------------------------------------------

func TestTrekB41KtingaGivesAKlingonFlyingAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	klingon := effectstest.B12Creature(g, me.ID, "Klingon", "Creature — Klingon", 2, 2)
	warrior := effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastFromHand(t, g, game.Card{Name: "K't'inga-Class Battle Cruiser", TypeLine: "Artifact — Spacecraft",
		OracleID: trekB41KtingaOracle, ManaCost: "{3}", Power: 5, Toughness: 5}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.SgAnswerTargets(t, g, bear); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a Bear as the target: err = %v, want ErrIllegalTarget", err)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{trekB41Card(klingon), trekB41Card(warrior)}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]bool{klingon: true, warrior: true, bear: false} {
		if got := effectstest.EffectiveAbilitiesContain(t, g, id, "flying"); got != want {
			t.Errorf("%s flying = %v, want %v", effects.CardNameFor(g, id), got, want)
		}
	}
}

func TestTrekB41KtingaStationsToAFlyingFiveFiveAtFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ship := effectstest.PushCatalogPermanent(g, me.ID, "K't'inga-Class Battle Cruiser", "Artifact — Spacecraft", trekB41KtingaOracle, false)
	crew := effectstest.B12Creature(g, me.ID, "Crew", "Creature — Human", 3, 3)
	more := effectstest.B12Creature(g, me.ID, "More Crew", "Creature — Human", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.Station(t, g, me, ship, crew)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LayeredCard(t, g, ship).IsCreature() {
		t.Fatal("a creature at three charge counters")
	}
	effectstest.Station(t, g, me, ship, more)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 5 || !game.HasKeyword(&c, "flying") {
		t.Errorf("at 4+ creature=%v power=%d flying=%v; want a flying 5/5", c.IsCreature(), c.CurrentPower(), game.HasKeyword(&c, "flying"))
	}
}

func TestTrekB41KtingaTriggersWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	ship := effectstest.PushCatalogPermanent(g, me.ID, "K't'inga-Class Battle Cruiser", "Artifact — Spacecraft", trekB41KtingaOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, ship, game.CounterCharge, 4) })
	klingon := effectstest.B12Creature(g, me.ID, "Klingon", "Creature — Klingon", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	effectstest.DeclareAttack(t, g, opp.ID, ship)
	effectstest.LockInAttacks(t, g)
	effectstest.PickCard(t, g, me.ID, klingon)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, klingon, "flying") {
		t.Error("the attack trigger did not give the Klingon flying")
	}
}
