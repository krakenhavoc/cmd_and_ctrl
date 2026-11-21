package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b15_cards_test.go — Star Trek, batch 15 (#4973). Each test plays
// the real catalog entry and asserts what its text says.

const (
	trekB15CaptainOracle     = "2375e6e4-bea2-4fa2-b0d5-467e3076962c"
	trekB15KirkOracle        = "19edde54-153f-46b8-8a8d-a20fdef0667f"
	trekB15BurnhamOracle     = "9a71aa6c-2d5f-468d-b588-252fae37ccbb"
	trekB15TutelageOracle    = "1d41aacf-e15d-4457-b12a-c71bcdb7b5de"
	trekB15CruiserOracle     = "5e696f9a-425c-470a-870e-10d2ba3bb27d"
	trekB15DaharOracle       = "c9bc1a28-8d3c-4d9f-a67d-b044cfe9d6c1"
	trekB15InfiltrateOracle  = "30597aff-ebb6-4927-b350-fe559fb92c8f"
	trekB15EncounterOracle   = "78c423ad-1f8b-4262-b64f-a98899da9372"
	trekB15EricaOracle       = "8b2514e5-4124-4f6c-b6b7-24c0ab3d3971"
	trekB15ExploratoryOracle = "3fdba8ee-6dfe-480d-b84a-43922229d9b4"
)

func TestTrekB15EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB15CaptainOracle:     "Cantankerous Captain",
		trekB15KirkOracle:        "Captain James T. Kirk",
		trekB15BurnhamOracle:     "Captain Michael Burnham",
		trekB15TutelageOracle:    "Captain's Tutelage",
		trekB15CruiserOracle:     "D-7 Battle Cruiser",
		trekB15DaharOracle:       "Dahar Master",
		trekB15InfiltrateOracle:  "Dominion Infiltration",
		trekB15EncounterOracle:   "Encounter the Unknown",
		trekB15EricaOracle:       "Erica Ortegas, Plucky Pilot",
		trekB15ExploratoryOracle: "Exploratory Mission",
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

// --- shared helpers ------------------------------------------------------------

func trekB15Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// trekB15Tokens is the tokens named `name` that p controls.
func trekB15Tokens(g *game.Game, p uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == p && effects.IsToken(c) {
				out = append(out, c)
			}
		}
	})
	return out
}

// trekB15ToEndStep walks the turn to its end step and resolves what fires.
func trekB15ToEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Cantankerous Captain ------------------------------------------------------

func TestTrekB15CantankerousCaptainPumpsAndGivesMenaceOnlyAtEightPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB15Seats(g)
	effectstest.B27Push(g, me.ID, "Cantankerous Captain", "Creature — Tellarite Officer", trekB15CaptainOracle, "{2}{R}", 2, 3, "R")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("targeted a creature an opponent controls")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tg := effectstest.B38PT(t, g, bear); pw != 3 || tg != 2 {
		t.Errorf("the Bear = %d/%d, want 3/2", pw, tg)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("total power 5 gave menace")
	}
}

// Seven power before the pump and eight after: "Then" reads the total
// after the +1/+0.
func TestTrekB15CantankerousCaptainCountsItsOwnPumpTowardEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB15Seats(g)
	effectstest.B27Push(g, me.ID, "Cantankerous Captain", "Creature — Tellarite Officer", trekB15CaptainOracle, "{2}{R}", 2, 3, "R")
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 5, 5)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if pw, _ := effectstest.B38PT(t, g, big); pw != 6 {
		t.Errorf("power = %d, want 6", pw)
	}
	if !effectstest.HasEffectiveKeyword(t, g, big, "menace") {
		t.Error("total power 8 after the pump: the creature should have menace")
	}
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.HasEffectiveKeyword(t, g, big, "menace") {
		t.Error("menace outlasted the turn")
	}
}

// --- Captain James T. Kirk -----------------------------------------------------

func TestTrekB15KirkChoosesOneWithCardsInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB15Seats(g)
	if me.Hand.Size() == 0 {
		t.Fatal("the fixture should hold cards in hand")
	}
	effectstest.CastHoldCreature(t, g, "Captain James T. Kirk", "Legendary Creature — Human Officer", trekB15KirkOracle, 2, 2)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for a mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1, 2}); err == nil {
		t.Fatal("chose two modes while holding cards")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	officers := trekB15Tokens(g, me.ID, "Officer")
	if len(officers) != 1 {
		t.Fatalf("%d Officers, want 1", len(officers))
	}
	o := officers[0]
	if !o.IsCreature() || !o.HasSubtype("Officer") || len(o.Colors) != 1 || o.Colors[0] != "R" || o.Power != 1 || o.Toughness != 1 {
		t.Errorf("Officer = %s %d/%d %v, want a 1/1 red Officer", o.TypeLine, o.Power, o.Toughness, o.Colors)
	}
}

func TestTrekB15KirkWithAnEmptyHandChoosesOneOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB15Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.EmptyHandToLibrary(g, me)
	kirk := effectstest.CastHoldCreature(t, g, "Captain James T. Kirk", "Legendary Creature — Human Officer", trekB15KirkOracle, 2, 2)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for modes")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1, 2}); err != nil {
		t.Fatalf("all three modes with no cards in hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The discard finds an empty hand; the draw still happens.
	if c := effectstest.DiscardChoiceFor(g, me.ID); c != nil {
		t.Fatal("asked to discard from an empty hand")
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand = %d, want 1 (the draw)", me.Hand.Size())
	}
	officers := trekB15Tokens(g, me.ID, "Officer")
	if len(officers) != 1 {
		t.Fatalf("%d Officers, want 1", len(officers))
	}
	if pw, _ := effectstest.B38PT(t, g, bear); pw != 3 {
		t.Errorf("my Bear's power = %d, want 3", pw)
	}
	if pw, _ := effectstest.B38PT(t, g, kirk); pw != 3 {
		t.Errorf("Kirk's power = %d, want 3", pw)
	}
	// The Officer was created before the pump in printed order, so it is
	// one of the creatures that get +1/+0.
	if pw, _ := effectstest.B38PT(t, g, officers[0].InstanceID); pw != 2 {
		t.Errorf("the Officer's power = %d, want 2", pw)
	}
	if pw, _ := effectstest.B38PT(t, g, theirs); pw != 2 {
		t.Errorf("an opponent's creature got the pump: power %d", pw)
	}
}

func TestTrekB15KirkAttackTriggerRummages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB15Seats(g)
	kirk := effectstest.B27Push(g, me.ID, "Captain James T. Kirk", "Legendary Creature — Human Officer", trekB15KirkOracle, "{2}{R}", 2, 2, "R")
	effectstest.EmptyHandToLibrary(g, me)
	pitch := effectstest.HandCard(me, "Pitch", "Sorcery")
	g.WithWriteLock(func() { _, _ = game.MoveCard(me.Library, me.Hand, me.Library.Cards[0].InstanceID) })
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(kirk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the attack trigger did not ask for a mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, pitch) != game.ZoneGraveyard {
		t.Errorf("the discarded card is in %v, want the graveyard", effectstest.B12ZoneOf(g, pitch))
	}
	if me.Hand.Size() != 2 {
		t.Errorf("hand = %d, want 2 (one kept, one drawn)", me.Hand.Size())
	}
}

// --- Captain Michael Burnham ---------------------------------------------------

func TestTrekB15BurnhamPingsEachLandfallAndGrowsOnTheSecond(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB15Seats(g)
	burnham := effectstest.B27Push(g, me.ID, "Captain Michael Burnham", "Legendary Creature — Human Officer", trekB15BurnhamOracle, "{2}{R}", 3, 3, "R")
	effectstest.AssertKeywords(t, g, burnham, "trample")
	start := effectstest.LifeOfOpponents(g)
	mine := effectstest.LifeOf(g, me.ID)
	wantCounters := []int{0, 1, 1}
	for i, want := range wantCounters {
		effectstest.PlayLandFromHand(t, g, "Mountain", "")
		effectstest.PassPriorityAroundTable(t, g)
		for j, life := range effectstest.LifeOfOpponents(g) {
			if life != start[j]-(i+1) {
				t.Errorf("landfall %d: opponent %d at %d life, want %d", i+1, j, life, start[j]-(i+1))
			}
		}
		if effectstest.LifeOf(g, me.ID) != mine {
			t.Errorf("landfall %d: Burnham's controller took damage (life %d)", i+1, effectstest.LifeOf(g, me.ID))
		}
		if got := effectstest.PlusOnes(g, burnham); got != want {
			t.Errorf("landfall %d: %d +1/+1 counters, want %d", i+1, got, want)
		}
	}
}

// --- Captain's Tutelage --------------------------------------------------------

func TestTrekB15CaptainsTutelageFlashesInForFirstStrikeAndTwoPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB15Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// Flash: cast at the beginning of combat, outside a main phase.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AddPool(t, g, me.ID, "{C}{R}")
	effectstest.CastFromHandForTest(t, g, me, "Captain's Tutelage", "Enchantment — Aura", "{1}{R}", trekB15TutelageOracle,
		game.CastSpellParams{Targets: effectstest.CardRefs(bear), Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tg := effectstest.B38PT(t, g, bear); pw != 4 || tg != 2 {
		t.Errorf("enchanted Bear = %d/%d, want 4/2", pw, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the enters trigger gave no first strike")
	}
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("first strike outlasted the turn")
	}
	if pw, _ := effectstest.B38PT(t, g, bear); pw != 4 {
		t.Errorf("the +2/+0 ended with the turn: power %d", pw)
	}
}

// --- D-7 Battle Cruiser --------------------------------------------------------

func TestTrekB15D7MakesASacrificedKlingonAtThreeChargeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB15Seats(g)
	ship := effectstest.B27Push(g, me.ID, "D-7 Battle Cruiser", "Artifact — Spacecraft", trekB15CruiserOracle, "{1}{R}", 3, 3, "R")
	crew := effectstest.B12Creature(g, me.ID, "Crew", "Creature — Klingon", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.Station(t, g, me, ship, crew)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ship, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters = %d, want 2", got)
	}
	if effectstest.LayeredCard(t, g, ship).IsCreature() {
		t.Error("the Spacecraft is a creature below 8+")
	}
	// Two counters: no trigger.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(trekB15Tokens(g, me.ID, "Klingon Warrior")); n != 0 {
		t.Fatalf("%d Klingon Warriors below 3+, want 0", n)
	}

	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ship, game.CounterCharge, 1) })
	// This seat's next turn: walk round the table.
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	tokens := trekB15Tokens(g, me.ID, "Klingon Warrior")
	if len(tokens) != 1 {
		t.Fatalf("%d Klingon Warriors at 3+, want 1", len(tokens))
	}
	k := tokens[0]
	if !k.IsCreature() || !k.HasSubtype("Klingon") || !k.HasSubtype("Warrior") || k.Power != 2 || k.Toughness != 1 ||
		len(k.Colors) != 1 || k.Colors[0] != "R" {
		t.Errorf("token = %s %d/%d %v, want a 2/1 red Klingon Warrior", k.TypeLine, k.Power, k.Toughness, k.Colors)
	}
	if !effectstest.HasEffectiveKeyword(t, g, k.InstanceID, "haste") {
		t.Error("the Klingon Warrior has no haste")
	}
	trekB15ToEndStep(t, g)
	if g.Battlefield.Contains(k.InstanceID) {
		t.Error("the Klingon Warrior was not sacrificed at the end step")
	}
}

func TestTrekB15D7IsAFlyingCreatureAtEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB15Seats(g)
	ship := effectstest.B27Push(g, me.ID, "D-7 Battle Cruiser", "Artifact — Spacecraft", trekB15CruiserOracle, "{1}{R}", 3, 3, "R")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ship, game.CounterCharge, 7) })
	if effectstest.LayeredCard(t, g, ship).IsCreature() || effectstest.HasEffectiveKeyword(t, g, ship, "flying") {
		t.Error("seven counters: not yet a flying creature")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ship, game.CounterCharge, 1) })
	if !effectstest.LayeredCard(t, g, ship).IsCreature() || !effectstest.HasEffectiveKeyword(t, g, ship, "flying") {
		t.Error("eight counters: a flying artifact creature")
	}
	if pw, tg := effectstest.B38PT(t, g, ship); pw != 3 || tg != 3 {
		t.Errorf("D-7 = %d/%d, want 3/3", pw, tg)
	}
}

// --- Dahar Master --------------------------------------------------------------

func TestTrekB15DaharMasterSacrificesANonlandPermanentForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB15Seats(g)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "My Mountain", "Basic Land — Mountain")
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	dahar := effectstest.CastHoldCreature(t, g, "Dahar Master", "Creature — Klingon Warrior", trekB15DaharOracle, 4, 3)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{land}); err == nil {
		t.Fatal("sacrificed a land")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{dahar}); err == nil {
		t.Fatal("sacrificed Dahar Master itself")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, rock)
	effectstest.B04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(rock) {
		t.Fatal("the sacrifice happens before the reflexive trigger targets")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("2 damage did not kill the 2/2")
	}
}

func TestTrekB15DaharMasterDeclinedDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB15Seats(g)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	start := effectstest.LifeOfOpponents(g)
	effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Dahar Master", "Creature — Klingon Warrior", trekB15DaharOracle, 4, 3)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a declined sacrifice still asked for a target")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the Rock left with the sacrifice declined")
	}
	for j, life := range effectstest.LifeOfOpponents(g) {
		if life != start[j] {
			t.Errorf("opponent %d lost life", j)
		}
	}
}

// --- Dominion Infiltration -----------------------------------------------------

func TestTrekB15DominionInfiltrationCopiesXOpposingCreaturesAsHastyChangelings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB15Seats(g)
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{R}{R}{R}")
	if _, err := effectstest.CastFromHandErr(g, me, "Dominion Infiltration", "Sorcery", "{X}{R}{R}{R}", trekB15InfiltrateOracle,
		game.CastSpellParams{XValue: 2, Targets: effectstest.CardRefs(a, mine), Strict: true}); err == nil {
		t.Fatal("targeted a creature I control")
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Dominion Infiltration", "Sorcery", "{X}{R}{R}{R}", trekB15InfiltrateOracle,
		game.CastSpellParams{XValue: 2, Targets: effectstest.CardRefs(a), Strict: true}); err == nil {
		t.Fatal("X=2 with one target")
	}
	effectstest.CastFromHandForTest(t, g, me, "Dominion Infiltration", "Sorcery", "{X}{R}{R}{R}", trekB15InfiltrateOracle,
		game.CastSpellParams{XValue: 2, Targets: effectstest.CardRefs(a, b), Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	elves, ogres := trekB15Tokens(g, me.ID, "Their Elf"), trekB15Tokens(g, me.ID, "Their Ogre")
	if len(elves) != 1 || len(ogres) != 1 {
		t.Fatalf("copies: %d Elves, %d Ogres, want one of each", len(elves), len(ogres))
	}
	for _, tok := range []game.Card{elves[0], ogres[0]} {
		if !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
			t.Errorf("%s copy has no haste", tok.Name)
		}
		if !effectstest.CardByID(g, tok.InstanceID).HasSubtype("Klingon") {
			t.Errorf("%s copy is not every creature type", tok.Name)
		}
	}
	if pw, tg := effectstest.B38PT(t, g, ogres[0].InstanceID); pw != 3 || tg != 3 {
		t.Errorf("the Ogre copy = %d/%d, want 3/3", pw, tg)
	}
	if effectstest.CardByID(g, b).HasSubtype("Klingon") {
		t.Error("the original gained changeling")
	}
	trekB15ToEndStep(t, g)
	for _, tok := range []game.Card{elves[0], ogres[0]} {
		if g.Battlefield.Contains(tok.InstanceID) {
			t.Errorf("%s copy was not exiled at the end step", tok.Name)
		}
	}
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("an original left the battlefield")
	}
}

// --- Encounter the Unknown -----------------------------------------------------

func TestTrekB15EncounterTheUnknownComeInPeaceExilesTwoToPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB15Seats(g)
	effectstest.AdvanceToMain(t, g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	effectstest.AddPool(t, g, me.ID, "{C}{R}{R}")
	effectstest.CastFromHandForTest(t, g, me, "Encounter the Unknown", "Instant", "{1}{R}{R}", trekB15EncounterOracle,
		game.CastSpellParams{Modes: []int{0}, Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || !g.Exile.Contains(ids[1]) || !me.Library.Contains(ids[2]) {
		t.Fatal("the top two cards are exiled, the third stays")
	}
	for _, id := range ids[:2] {
		perm := effectstest.ExiledPermission(g, id)
		if perm.Player != me.ID || perm.CastOnly || !effectstest.PermissionLive(g, perm, me.ID) {
			t.Errorf("%+v: the caster may PLAY it", perm)
		}
	}
}

func TestTrekB15EncounterTheUnknownShootToKillHitsEveryCreatureAndPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB15Seats(g)
	effectstest.AdvanceToMain(t, g)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 3, 3)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = effectstest.LifeOf(g, p.ID)
	}
	effectstest.AddPool(t, g, me.ID, "{C}{R}{R}")
	effectstest.CastFromHandForTest(t, g, me, "Encounter the Unknown", "Instant", "{1}{R}{R}", trekB15EncounterOracle,
		game.CastSpellParams{Modes: []int{1}, Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(mine) {
		t.Error("the 2/2s survived 2 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 3/3 died to 2 damage")
	}
	for _, p := range g.Seats {
		if got := effectstest.LifeOf(g, p.ID); got != before[p.ID]-2 {
			t.Errorf("%s at %d life, want %d", p.Name, got, before[p.ID]-2)
		}
	}
}

// --- Erica Ortegas, Plucky Pilot -----------------------------------------------

func TestTrekB15EricaPutsALandWhenTappedAndReturnsOneAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB15Seats(g)
	erica := effectstest.B27Push(g, me.ID, "Erica Ortegas, Plucky Pilot", "Legendary Creature — Human Pilot", trekB15EricaOracle, "{3}{R}", 3, 3, "R")
	effectstest.AssertKeywords(t, g, erica, "haste")
	old := effectstest.B12Permanent(g, me.ID, "Old Mountain", "Basic Land — Mountain")
	land := effectstest.HandCard(me, "Hand Mountain", "Basic Land — Mountain")
	spell := effectstest.HandCard(me, "Hand Spell", "Sorcery")
	effectstest.DeclareAttack(t, g, opp.ID, erica)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt to put a land from hand")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{spell}); err == nil {
		t.Fatal("put a nonland card onto the battlefield")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, land) != game.ZoneBattlefield {
		t.Fatalf("the land is in %v, want the battlefield", effectstest.B12ZoneOf(g, land))
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	for i := 0; i < 8 && effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	// Any land you control: the old one, not the land that was put.
	effectstest.AnswerOwnPermanents(t, g, me.ID, old)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, old) != game.ZoneHand {
		t.Errorf("the chosen land is in %v, want its owner's hand", effectstest.B12ZoneOf(g, old))
	}
	if effectstest.B12ZoneOf(g, land) != game.ZoneBattlefield {
		t.Error("the land that was put left too")
	}
}

func TestTrekB15EricaDeclinedSchedulesNoReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB15Seats(g)
	erica := effectstest.B27Push(g, me.ID, "Erica Ortegas, Plucky Pilot", "Legendary Creature — Human Pilot", trekB15EricaOracle, "{3}{R}", 3, 3, "R")
	old := effectstest.B12Permanent(g, me.ID, "Old Mountain", "Basic Land — Mountain")
	effectstest.HandCard(me, "Hand Mountain", "Basic Land — Mountain")
	effectstest.DeclareAttack(t, g, opp.ID, erica)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt to put a land from hand")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers after declining, want 0", len(g.DelayedTriggers))
	}
	trekB15ToEndStep(t, g)
	if effectstest.B12ZoneOf(g, old) != game.ZoneBattlefield {
		t.Error("a land was returned though none was put")
	}
}

// --- Exploratory Mission -------------------------------------------------------

func TestTrekB15ExploratoryMissionExilesToANonlandAndAddsManaAtFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB15Seats(g)
	mission := effectstest.B27Push(g, me.ID, "Exploratory Mission", "Enchantment", trekB15ExploratoryOracle, "{2}{R}{R}", 0, 0, "R")
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Exiled Land", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	effectstest.PlayLandFromHand(t, g, "Mountain", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, mission, "quest"); got != 1 {
		t.Fatalf("quest counters = %d, want 1", got)
	}
	if !g.Exile.Contains(ids[0]) || !g.Exile.Contains(ids[1]) || !me.Library.Contains(ids[2]) {
		t.Fatal("exile until a nonland card: the land and the Shock, not the third card")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(ids[0]); perm != nil {
		t.Error("the land exiled on the way may be played; only \"that card\" may")
	}
	perm := effectstest.ExiledPermission(g, ids[1])
	if perm.Player != me.ID || !effectstest.PermissionLive(g, perm, me.ID) {
		t.Errorf("%+v: the nonland card may be played this turn", perm)
	}
	if got := me.ManaPool.CountColor("R"); got != 0 {
		t.Errorf("one quest counter added %d red mana", got)
	}
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, ids[1]), me.ID) {
		t.Error("the permission outlasted the turn")
	}
}

func TestTrekB15ExploratoryMissionAddsRedRedOnTheFifthCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB15Seats(g)
	mission := effectstest.B27Push(g, me.ID, "Exploratory Mission", "Enchantment", trekB15ExploratoryOracle, "{2}{R}{R}", 0, 0, "R")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mission, "quest", 4) })
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Shock", TypeLine: "Instant"})
	effectstest.PlayLandFromHand(t, g, "Mountain", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, mission, "quest"); got != 5 {
		t.Fatalf("quest counters = %d, want 5", got)
	}
	if got := me.ManaPool.CountColor("R"); got != 2 {
		t.Errorf("red mana = %d, want 2", got)
	}
}
