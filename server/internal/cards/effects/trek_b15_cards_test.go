package effects

import (
	"testing"

	"github.com/google/uuid"

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

// --- shared helpers ------------------------------------------------------------

func trekB15Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// trekB15Tokens is the tokens named `name` that p controls.
func trekB15Tokens(g *game.Game, p uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == p && IsToken(c) {
				out = append(out, c)
			}
		}
	})
	return out
}

// trekB15ToEndStep walks the turn to its end step and resolves what fires.
func trekB15ToEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
}

// --- Cantankerous Captain ------------------------------------------------------

func TestTrekB15CantankerousCaptainPumpsAndGivesMenaceOnlyAtEightPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	b27Push(g, me.ID, "Cantankerous Captain", "Creature — Tellarite Officer", trekB15CaptainOracle, "{2}{R}", 2, 3, "R")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("targeted a creature an opponent controls")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if pw, tg := b38PT(t, g, bear); pw != 3 || tg != 2 {
		t.Errorf("the Bear = %d/%d, want 3/2", pw, tg)
	}
	if hasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("total power 5 gave menace")
	}
}

// Seven power before the pump and eight after: "Then" reads the total
// after the +1/+0.
func TestTrekB15CantankerousCaptainCountsItsOwnPumpTowardEight(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB15Seats(g)
	b27Push(g, me.ID, "Cantankerous Captain", "Creature — Tellarite Officer", trekB15CaptainOracle, "{2}{R}", 2, 3, "R")
	big := b12Creature(g, me.ID, "Big", "Creature — Giant", 5, 5)
	advanceTo(t, g, game.StepBeginCombat)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if pw, _ := b38PT(t, g, big); pw != 6 {
		t.Errorf("power = %d, want 6", pw)
	}
	if !hasEffectiveKeyword(t, g, big, "menace") {
		t.Error("total power 8 after the pump: the creature should have menace")
	}
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if hasEffectiveKeyword(t, g, big, "menace") {
		t.Error("menace outlasted the turn")
	}
}

// --- Captain James T. Kirk -----------------------------------------------------

func TestTrekB15KirkChoosesOneWithCardsInHand(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB15Seats(g)
	if me.Hand.Size() == 0 {
		t.Fatal("the fixture should hold cards in hand")
	}
	castHoldCreature(t, g, "Captain James T. Kirk", "Legendary Creature — Human Officer", trekB15KirkOracle, 2, 2)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for a mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1, 2}); err == nil {
		t.Fatal("chose two modes while holding cards")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	emptyHandToLibrary(g, me)
	kirk := castHoldCreature(t, g, "Captain James T. Kirk", "Legendary Creature — Human Officer", trekB15KirkOracle, 2, 2)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for modes")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1, 2}); err != nil {
		t.Fatalf("all three modes with no cards in hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The discard finds an empty hand; the draw still happens.
	if c := discardChoiceFor(g, me.ID); c != nil {
		t.Fatal("asked to discard from an empty hand")
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand = %d, want 1 (the draw)", me.Hand.Size())
	}
	officers := trekB15Tokens(g, me.ID, "Officer")
	if len(officers) != 1 {
		t.Fatalf("%d Officers, want 1", len(officers))
	}
	if pw, _ := b38PT(t, g, bear); pw != 3 {
		t.Errorf("my Bear's power = %d, want 3", pw)
	}
	if pw, _ := b38PT(t, g, kirk); pw != 3 {
		t.Errorf("Kirk's power = %d, want 3", pw)
	}
	// The Officer was created before the pump in printed order, so it is
	// one of the creatures that get +1/+0.
	if pw, _ := b38PT(t, g, officers[0].InstanceID); pw != 2 {
		t.Errorf("the Officer's power = %d, want 2", pw)
	}
	if pw, _ := b38PT(t, g, theirs); pw != 2 {
		t.Errorf("an opponent's creature got the pump: power %d", pw)
	}
}

func TestTrekB15KirkAttackTriggerRummages(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	kirk := b27Push(g, me.ID, "Captain James T. Kirk", "Legendary Creature — Human Officer", trekB15KirkOracle, "{2}{R}", 2, 2, "R")
	emptyHandToLibrary(g, me)
	pitch := handCard(me, "Pitch", "Sorcery")
	g.WithWriteLock(func() { _, _ = game.MoveCard(me.Library, me.Hand, me.Library.Cards[0].InstanceID) })
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(kirk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the attack trigger did not ask for a mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, pitch) != game.ZoneGraveyard {
		t.Errorf("the discarded card is in %v, want the graveyard", b12ZoneOf(g, pitch))
	}
	if me.Hand.Size() != 2 {
		t.Errorf("hand = %d, want 2 (one kept, one drawn)", me.Hand.Size())
	}
}

// --- Captain Michael Burnham ---------------------------------------------------

func TestTrekB15BurnhamPingsEachLandfallAndGrowsOnTheSecond(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB15Seats(g)
	burnham := b27Push(g, me.ID, "Captain Michael Burnham", "Legendary Creature — Human Officer", trekB15BurnhamOracle, "{2}{R}", 3, 3, "R")
	assertKeywords(t, g, burnham, "trample")
	start := lifeOfOpponents(g)
	mine := lifeOf(g, me.ID)
	wantCounters := []int{0, 1, 1}
	for i, want := range wantCounters {
		playLandFromHand(t, g, "Mountain", "")
		passPriorityAroundTable(t, g)
		for j, life := range lifeOfOpponents(g) {
			if life != start[j]-(i+1) {
				t.Errorf("landfall %d: opponent %d at %d life, want %d", i+1, j, life, start[j]-(i+1))
			}
		}
		if lifeOf(g, me.ID) != mine {
			t.Errorf("landfall %d: Burnham's controller took damage (life %d)", i+1, lifeOf(g, me.ID))
		}
		if got := plusOnes(g, burnham); got != want {
			t.Errorf("landfall %d: %d +1/+1 counters, want %d", i+1, got, want)
		}
	}
}

// --- Captain's Tutelage --------------------------------------------------------

func TestTrekB15CaptainsTutelageFlashesInForFirstStrikeAndTwoPower(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB15Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// Flash: cast at the beginning of combat, outside a main phase.
	advanceTo(t, g, game.StepBeginCombat)
	addPool(t, g, me.ID, "{C}{R}")
	castFromHandForTest(t, g, me, "Captain's Tutelage", "Enchantment — Aura", "{1}{R}", trekB15TutelageOracle,
		game.CastSpellParams{Targets: cardRefs(bear), Strict: true})
	passPriorityAroundTable(t, g)
	if pw, tg := b38PT(t, g, bear); pw != 4 || tg != 2 {
		t.Errorf("enchanted Bear = %d/%d, want 4/2", pw, tg)
	}
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the enters trigger gave no first strike")
	}
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("first strike outlasted the turn")
	}
	if pw, _ := b38PT(t, g, bear); pw != 4 {
		t.Errorf("the +2/+0 ended with the turn: power %d", pw)
	}
}

// --- D-7 Battle Cruiser --------------------------------------------------------

func TestTrekB15D7MakesASacrificedKlingonAtThreeChargeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB15Seats(g)
	ship := b27Push(g, me.ID, "D-7 Battle Cruiser", "Artifact — Spacecraft", trekB15CruiserOracle, "{1}{R}", 3, 3, "R")
	crew := b12Creature(g, me.ID, "Crew", "Creature — Klingon", 2, 2)
	advanceToMain(t, g)
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ship, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters = %d, want 2", got)
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("the Spacecraft is a creature below 8+")
	}
	// Two counters: no trigger.
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if n := len(trekB15Tokens(g, me.ID, "Klingon Warrior")); n != 0 {
		t.Fatalf("%d Klingon Warriors below 3+, want 0", n)
	}

	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ship, game.CounterCharge, 1) })
	// This seat's next turn: walk round the table.
	seat := g.Turn.ActiveSeat
	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	advanceToUpkeepOf(t, g, seat)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	tokens := trekB15Tokens(g, me.ID, "Klingon Warrior")
	if len(tokens) != 1 {
		t.Fatalf("%d Klingon Warriors at 3+, want 1", len(tokens))
	}
	k := tokens[0]
	if !k.IsCreature() || !k.HasSubtype("Klingon") || !k.HasSubtype("Warrior") || k.Power != 2 || k.Toughness != 1 ||
		len(k.Colors) != 1 || k.Colors[0] != "R" {
		t.Errorf("token = %s %d/%d %v, want a 2/1 red Klingon Warrior", k.TypeLine, k.Power, k.Toughness, k.Colors)
	}
	if !hasEffectiveKeyword(t, g, k.InstanceID, "haste") {
		t.Error("the Klingon Warrior has no haste")
	}
	trekB15ToEndStep(t, g)
	if g.Battlefield.Contains(k.InstanceID) {
		t.Error("the Klingon Warrior was not sacrificed at the end step")
	}
}

func TestTrekB15D7IsAFlyingCreatureAtEight(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB15Seats(g)
	ship := b27Push(g, me.ID, "D-7 Battle Cruiser", "Artifact — Spacecraft", trekB15CruiserOracle, "{1}{R}", 3, 3, "R")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ship, game.CounterCharge, 7) })
	if layeredCard(t, g, ship).IsCreature() || hasEffectiveKeyword(t, g, ship, "flying") {
		t.Error("seven counters: not yet a flying creature")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ship, game.CounterCharge, 1) })
	if !layeredCard(t, g, ship).IsCreature() || !hasEffectiveKeyword(t, g, ship, "flying") {
		t.Error("eight counters: a flying artifact creature")
	}
	if pw, tg := b38PT(t, g, ship); pw != 3 || tg != 3 {
		t.Errorf("D-7 = %d/%d, want 3/3", pw, tg)
	}
}

// --- Dahar Master --------------------------------------------------------------

func TestTrekB15DaharMasterSacrificesANonlandPermanentForTwoDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	land := b12Permanent(g, me.ID, "My Mountain", "Basic Land — Mountain")
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	dahar := castHoldCreature(t, g, "Dahar Master", "Creature — Klingon Warrior", trekB15DaharOracle, 4, 3)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{land}); err == nil {
		t.Fatal("sacrificed a land")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{dahar}); err == nil {
		t.Fatal("sacrificed Dahar Master itself")
	}
	answerOwnPermanents(t, g, me.ID, rock)
	b04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(rock) {
		t.Fatal("the sacrifice happens before the reflexive trigger targets")
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("2 damage did not kill the 2/2")
	}
}

func TestTrekB15DaharMasterDeclinedDealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	start := lifeOfOpponents(g)
	b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Dahar Master", "Creature — Klingon Warrior", trekB15DaharOracle, 4, 3)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a declined sacrifice still asked for a target")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the Rock left with the sacrifice declined")
	}
	for j, life := range lifeOfOpponents(g) {
		if life != start[j] {
			t.Errorf("opponent %d lost life", j)
		}
	}
}

// --- Dominion Infiltration -----------------------------------------------------

func TestTrekB15DominionInfiltrationCopiesXOpposingCreaturesAsHastyChangelings(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	advanceToMain(t, g)
	a := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 2, 2)
	b := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	addPool(t, g, me.ID, "{C}{C}{R}{R}{R}")
	if _, err := castFromHandErr(g, me, "Dominion Infiltration", "Sorcery", "{X}{R}{R}{R}", trekB15InfiltrateOracle,
		game.CastSpellParams{XValue: 2, Targets: cardRefs(a, mine), Strict: true}); err == nil {
		t.Fatal("targeted a creature I control")
	}
	if _, err := castFromHandErr(g, me, "Dominion Infiltration", "Sorcery", "{X}{R}{R}{R}", trekB15InfiltrateOracle,
		game.CastSpellParams{XValue: 2, Targets: cardRefs(a), Strict: true}); err == nil {
		t.Fatal("X=2 with one target")
	}
	castFromHandForTest(t, g, me, "Dominion Infiltration", "Sorcery", "{X}{R}{R}{R}", trekB15InfiltrateOracle,
		game.CastSpellParams{XValue: 2, Targets: cardRefs(a, b), Strict: true})
	passPriorityAroundTable(t, g)
	elves, ogres := trekB15Tokens(g, me.ID, "Their Elf"), trekB15Tokens(g, me.ID, "Their Ogre")
	if len(elves) != 1 || len(ogres) != 1 {
		t.Fatalf("copies: %d Elves, %d Ogres, want one of each", len(elves), len(ogres))
	}
	for _, tok := range []game.Card{elves[0], ogres[0]} {
		if !hasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
			t.Errorf("%s copy has no haste", tok.Name)
		}
		if !cardByID(g, tok.InstanceID).HasSubtype("Klingon") {
			t.Errorf("%s copy is not every creature type", tok.Name)
		}
	}
	if pw, tg := b38PT(t, g, ogres[0].InstanceID); pw != 3 || tg != 3 {
		t.Errorf("the Ogre copy = %d/%d, want 3/3", pw, tg)
	}
	if cardByID(g, b).HasSubtype("Klingon") {
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
	g := newCatalogGame(t)
	me, _ := trekB15Seats(g)
	advanceToMain(t, g)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	addPool(t, g, me.ID, "{C}{R}{R}")
	castFromHandForTest(t, g, me, "Encounter the Unknown", "Instant", "{1}{R}{R}", trekB15EncounterOracle,
		game.CastSpellParams{Modes: []int{0}, Strict: true})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || !g.Exile.Contains(ids[1]) || !me.Library.Contains(ids[2]) {
		t.Fatal("the top two cards are exiled, the third stays")
	}
	for _, id := range ids[:2] {
		perm := exiledPermission(g, id)
		if perm.Player != me.ID || perm.CastOnly || !permissionLive(g, perm, me.ID) {
			t.Errorf("%+v: the caster may PLAY it", perm)
		}
	}
}

func TestTrekB15EncounterTheUnknownShootToKillHitsEveryCreatureAndPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	advanceToMain(t, g)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 3, 3)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = lifeOf(g, p.ID)
	}
	addPool(t, g, me.ID, "{C}{R}{R}")
	castFromHandForTest(t, g, me, "Encounter the Unknown", "Instant", "{1}{R}{R}", trekB15EncounterOracle,
		game.CastSpellParams{Modes: []int{1}, Strict: true})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(mine) {
		t.Error("the 2/2s survived 2 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 3/3 died to 2 damage")
	}
	for _, p := range g.Seats {
		if got := lifeOf(g, p.ID); got != before[p.ID]-2 {
			t.Errorf("%s at %d life, want %d", p.Name, got, before[p.ID]-2)
		}
	}
}

// --- Erica Ortegas, Plucky Pilot -----------------------------------------------

func TestTrekB15EricaPutsALandWhenTappedAndReturnsOneAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	erica := b27Push(g, me.ID, "Erica Ortegas, Plucky Pilot", "Legendary Creature — Human Pilot", trekB15EricaOracle, "{3}{R}", 3, 3, "R")
	assertKeywords(t, g, erica, "haste")
	old := b12Permanent(g, me.ID, "Old Mountain", "Basic Land — Mountain")
	land := handCard(me, "Hand Mountain", "Basic Land — Mountain")
	spell := handCard(me, "Hand Spell", "Sorcery")
	declareAttack(t, g, opp.ID, erica)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt to put a land from hand")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{spell}); err == nil {
		t.Fatal("put a nonland card onto the battlefield")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, land) != game.ZoneBattlefield {
		t.Fatalf("the land is in %v, want the battlefield", b12ZoneOf(g, land))
	}
	advanceTo(t, g, game.StepEnd)
	for i := 0; i < 8 && latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	// Any land you control: the old one, not the land that was put.
	answerOwnPermanents(t, g, me.ID, old)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, old) != game.ZoneHand {
		t.Errorf("the chosen land is in %v, want its owner's hand", b12ZoneOf(g, old))
	}
	if b12ZoneOf(g, land) != game.ZoneBattlefield {
		t.Error("the land that was put left too")
	}
}

func TestTrekB15EricaDeclinedSchedulesNoReturn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	erica := b27Push(g, me.ID, "Erica Ortegas, Plucky Pilot", "Legendary Creature — Human Pilot", trekB15EricaOracle, "{3}{R}", 3, 3, "R")
	old := b12Permanent(g, me.ID, "Old Mountain", "Basic Land — Mountain")
	handCard(me, "Hand Mountain", "Basic Land — Mountain")
	declareAttack(t, g, opp.ID, erica)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt to put a land from hand")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers after declining, want 0", len(g.DelayedTriggers))
	}
	trekB15ToEndStep(t, g)
	if b12ZoneOf(g, old) != game.ZoneBattlefield {
		t.Error("a land was returned though none was put")
	}
}

// --- Exploratory Mission -------------------------------------------------------

func TestTrekB15ExploratoryMissionExilesToANonlandAndAddsManaAtFive(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB15Seats(g)
	mission := b27Push(g, me.ID, "Exploratory Mission", "Enchantment", trekB15ExploratoryOracle, "{2}{R}{R}", 0, 0, "R")
	ids := seedSearchLibrary(me,
		game.Card{Name: "Exiled Land", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	playLandFromHand(t, g, "Mountain", "")
	passPriorityAroundTable(t, g)
	if got := countersOn(g, mission, "quest"); got != 1 {
		t.Fatalf("quest counters = %d, want 1", got)
	}
	if !g.Exile.Contains(ids[0]) || !g.Exile.Contains(ids[1]) || !me.Library.Contains(ids[2]) {
		t.Fatal("exile until a nonland card: the land and the Shock, not the third card")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(ids[0]); perm != nil {
		t.Error("the land exiled on the way may be played; only \"that card\" may")
	}
	perm := exiledPermission(g, ids[1])
	if perm.Player != me.ID || !permissionLive(g, perm, me.ID) {
		t.Errorf("%+v: the nonland card may be played this turn", perm)
	}
	if got := me.ManaPool.CountColor("R"); got != 0 {
		t.Errorf("one quest counter added %d red mana", got)
	}
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if permissionLive(g, exiledPermission(g, ids[1]), me.ID) {
		t.Error("the permission outlasted the turn")
	}
}

func TestTrekB15ExploratoryMissionAddsRedRedOnTheFifthCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB15Seats(g)
	mission := b27Push(g, me.ID, "Exploratory Mission", "Enchantment", trekB15ExploratoryOracle, "{2}{R}{R}", 0, 0, "R")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mission, "quest", 4) })
	seedSearchLibrary(me, game.Card{Name: "Shock", TypeLine: "Instant"})
	playLandFromHand(t, g, "Mountain", "")
	passPriorityAroundTable(t, g)
	if got := countersOn(g, mission, "quest"); got != 5 {
		t.Fatalf("quest counters = %d, want 5", got)
	}
	if got := me.ManaPool.CountColor("R"); got != 2 {
		t.Errorf("red mana = %d, want 2", got)
	}
}
