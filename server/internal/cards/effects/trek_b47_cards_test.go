package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b47_cards_test.go — Star Trek Commander, batch 47 (#4974). Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB47JeanLucOracle = "3855cb66-2e29-4d4b-a752-214843b647ff"
	trekB47TroiOracle    = "38561191-ecd1-4028-a0fd-4bca6a7eab58"
	trekB47OfficerOracle = "c8ed45e6-ae4f-4c44-9f73-45793f30acd6"
	trekB47PicardOracle  = "cda786aa-1259-45e8-a834-21c7ad0259a6"
	trekB47SpockOracle   = "a877820f-8999-4dfe-a50d-09f63c885ce3"
	trekB47TrelaneOracle = "3c516327-00a3-4d03-87a3-73b4a90719d3"
	trekB47AllocOracle   = "1ca22ca0-eca3-4578-9746-47d443d5fb73"
	trekB47QueenOracle   = "f9b46a1a-474f-4fac-8d71-131c1720e4c0"
	trekB47KirkOracle    = "6a70e571-6932-45d8-9f14-ede4a561cc49"
	trekB47KlingonOracle = "847efb61-5933-43c1-93ff-e8375f9469a9"
)

func TestTrekB47EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB47JeanLucOracle: "Jean-Luc, Mon Capitaine",
		trekB47TroiOracle:    "Troi, Winging It",
		trekB47OfficerOracle: "Operations Officer",
		trekB47PicardOracle:  "Picard, Steadfast Captain",
		trekB47SpockOracle:   "Spock, Logical Choice",
		trekB47TrelaneOracle: "Trelane, Squire of Gothos",
		trekB47AllocOracle:   "Bio-Asset Allocator",
		trekB47QueenOracle:   "Borg Queen, Perfection Manifest",
		trekB47KirkOracle:    "Kirk, Enterprising Captain",
		trekB47KlingonOracle: "Klingon Strike Force",
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

// --- fixtures -------------------------------------------------------------------

// trekB47Attack declares the attackers at the next player and locks the
// declaration in, leaving the attack triggers to answer or resolve.
func trekB47Attack(t *testing.T, g *game.Game, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	defender := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
}

// trekB47Tokens lists the tokens named `name` that p controls.
func trekB47Tokens(g *game.Game, p uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == p && c.Name == name && IsToken(c) {
				out = append(out, c)
			}
		}
	})
	return out
}

// trekB47OnTop puts a card on top of p's library.
func trekB47OnTop(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}

// trekB47OpponentsLife is every other seat's life total, in seat order.
func trekB47OpponentsLife(g *game.Game, me uuid.UUID) []int {
	var out []int
	for _, p := range g.Seats {
		if p.ID != me {
			out = append(out, p.Life)
		}
	}
	return out
}

// --- Jean-Luc, Mon Capitaine -----------------------------------------------------

func TestTrekB47JeanLucFaeriesEnterWithACounterPerNonBorgType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jl := b27Push(g, me.ID, "Jean-Luc, Mon Capitaine", "Legendary Creature — Human Officer", trekB47JeanLucOracle, "{2}{U}{R}{W}", 5, 3, "U", "R", "W")
	assertKeywords(t, g, jl, "flying")
	b12Creature(g, me.ID, "Elf Warrior", "Creature — Elf Warrior", 2, 2)
	// A second Human counts nothing new; the Borg's types count nothing.
	b12Creature(g, me.ID, "Ensign", "Creature — Human", 1, 1)
	b12Push(g, me.ID, "Drone", "Artifact Creature — Borg Soldier", "", 1, 1)
	faerie := castFromHand(t, g, game.Card{Name: "Faerie Seer", TypeLine: "Creature — Faerie Wizard", ManaCost: "{U}", Power: 1, Toughness: 1}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	// Human, Officer, Elf, Warrior.
	if n := plusOneCounters(g, faerie); n != 4 {
		t.Errorf("the Faerie entered with %d +1/+1 counters, want 4", n)
	}
	if n := plusOneCounters(g, jl); n != 0 {
		t.Errorf("Jean-Luc has %d counters, want none", n)
	}
}

func TestTrekB47JeanLucLeavesNonFaeriesAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b27Push(g, me.ID, "Jean-Luc, Mon Capitaine", "Legendary Creature — Human Officer", trekB47JeanLucOracle, "{2}{U}{R}{W}", 5, 3, "U", "R", "W")
	bear := castFromHand(t, g, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 2, Toughness: 2}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, bear); n != 0 {
		t.Errorf("a non-Faerie entered with %d counters, want none", n)
	}
}

// --- Troi, Winging It ------------------------------------------------------------

func TestTrekB47TroiMakesAFaeriePerOpponentDrawOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	pushCatalogPermanent(g, me.ID, "Troi, Winging It", "Legendary Creature — Betazoid Human Advisor", trekB47TroiOracle, false)
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	faeries := trekB47Tokens(g, me.ID, "Faerie")
	if len(faeries) != 2 {
		t.Fatalf("%d Faerie tokens, want 2", len(faeries))
	}
	v := layeredCard(t, g, faeries[0].InstanceID)
	if v.CurrentPower() != 1 || v.CurrentToughness() != 1 || !v.HasSubtype("Faerie") ||
		len(v.Colors) != 1 || v.Colors[0] != "U" {
		t.Errorf("the token is %q %d/%d %v, want a 1/1 blue Faerie", v.TypeLine, v.CurrentPower(), v.CurrentToughness(), v.Colors)
	}
	assertKeywords(t, g, faeries[0].InstanceID, "flying")
}

func TestTrekB47TroiIgnoresYourDrawsAndOtherTurns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	pushCatalogPermanent(g, me.ID, "Troi, Winging It", "Legendary Creature — Betazoid Human Advisor", trekB47TroiOracle, false)
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	// The opponent's own turn: their draw step and an extra draw.
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if n := len(trekB47Tokens(g, me.ID, "Faerie")); n != 0 {
		t.Errorf("%d Faerie tokens, want none", n)
	}
}

// --- Operations Officer ----------------------------------------------------------

func TestTrekB47OperationsOfficerDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	officer := castFromHand(t, g, game.Card{Name: "Operations Officer", TypeLine: "Creature — Human Officer", OracleID: trekB47OfficerOracle, ManaCost: "{3}{W}", Power: 2, Toughness: 2}, game.CastSpellParams{})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	assertKeywords(t, g, officer, "lifelink")
}

func TestTrekB47OperationsOfficerGrowsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	officer := b27Push(g, me.ID, "Operations Officer", "Creature — Human Officer", trekB47OfficerOracle, "{3}{W}", 2, 2, "W")
	trekB47Attack(t, g, officer)
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, officer); n != 1 {
		t.Errorf("%d +1/+1 counters after attacking, want 1", n)
	}
}

// --- Picard, Steadfast Captain ---------------------------------------------------

func TestTrekB47PicardPutsTwoCountersOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	picard := castFromHand(t, g, game.Card{Name: "Picard, Steadfast Captain", TypeLine: "Legendary Creature — Human Officer", OracleID: trekB47PicardOracle, ManaCost: "{4}{W}{W}", Power: 4, Toughness: 4}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, bear); n != 2 {
		t.Errorf("the bear has %d counters, want 2", n)
	}
	if n := plusOneCounters(g, picard); n != 0 {
		t.Errorf("Picard has %d counters, want none", n)
	}
}

func TestTrekB47PicardPutsTwoCountersWhenHeAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	picard := b27Push(g, me.ID, "Picard, Steadfast Captain", "Legendary Creature — Human Officer", trekB47PicardOracle, "{4}{W}{W}", 4, 4, "W")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	trekB47Attack(t, g, picard)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the attack trigger asked for no target")
	}
	for _, illegal := range []uuid.UUID{picard, theirs} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: illegal}); err == nil {
			t.Errorf("the trigger accepted %s as its target", findBattlefieldCardForTest(g, illegal).Name)
		}
	}
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, bear); n != 2 {
		t.Errorf("the bear has %d counters, want 2", n)
	}
}

// --- Spock, Logical Choice -------------------------------------------------------

func TestTrekB47SpockTapsToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	spock := pushCatalogPermanent(g, me.ID, "Spock, Logical Choice", "Legendary Creature — Vulcan Human Scientist", trekB47SpockOracle, false)
	assertKeywords(t, g, spock, "vigilance")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, spock, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if c, _ := battlefieldCard(g, spock); !c.Tapped {
		t.Error("Spock did not tap")
	}
	if err := g.ActivateCatalogAbility(me.ID, spock, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a second time while tapped")
	}
}

func TestTrekB47SpockCantTapTheTurnHeArrives(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	spock := pushCatalogPermanent(g, me.ID, "Spock, Logical Choice", "Legendary Creature — Vulcan Human Scientist", trekB47SpockOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, spock, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a summoning-sick Spock activated his {T} ability")
	}
}

// --- Trelane, Squire of Gothos ---------------------------------------------------

func TestTrekB47TrelaneIsUnblockableWithVigilanceAndWardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trelane := pushCatalogPermanent(g, me.ID, "Trelane, Squire of Gothos", "Legendary Creature — Q", trekB47TrelaneOracle, false)
	assertKeywords(t, g, trelane, "vigilance")
	assertRestrictions(t, g, trelane, game.CantBeBlocked)
	advanceToMainOf(t, g, 1)
	shock := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: shock, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{Targets: cardRefs(trelane)}); err != nil {
		t.Fatalf("an opponent targets Trelane: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("ward {2} did not ask the targeting opponent for payment")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(shock) {
		t.Error("declining the ward did not counter the spell")
	}
}

// --- Bio-Asset Allocator ---------------------------------------------------------

func TestTrekB47AllocatorGivesArtifactCreaturesAnAttackDrain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alloc := b27Push(g, me.ID, "Bio-Asset Allocator", "Artifact Creature — Borg", trekB47AllocOracle, "{3}{B}{B}", 3, 4, "B")
	robot := b12Push(g, me.ID, "Robot", "Artifact Creature — Robot", "", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before, mine := trekB47OpponentsLife(g, me.ID), me.Life
	trekB47Attack(t, g, alloc, robot, bear)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	// Two artifact creatures attacked; the bear has no such ability.
	for i, life := range trekB47OpponentsLife(g, me.ID) {
		if life != before[i]-2 {
			t.Errorf("opponent %d at %d, want %d", i, life, before[i]-2)
		}
	}
	if me.Life != mine+2 {
		t.Errorf("you are at %d, want %d", me.Life, mine+2)
	}
}

// --- Borg Queen, Perfection Manifest ---------------------------------------------

func TestTrekB47BorgQueenPumpsYourArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	queen := b27Push(g, me.ID, "Borg Queen, Perfection Manifest", "Legendary Artifact Creature — Borg Noble", trekB47QueenOracle, "{4}{B}{B}", 1, 4, "B")
	robot := b12Push(g, me.ID, "Robot", "Artifact Creature — Robot", "", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Push(g, opp.ID, "Their Robot", "Artifact Creature — Robot", "", 2, 2)
	for id, want := range map[uuid.UUID][2]int{queen: {3, 4}, robot: {4, 2}, bear: {2, 2}, theirs: {2, 2}} {
		if p, tg := effectivePower(t, g, id), effectiveToughness(t, g, id); p != want[0] || tg != want[1] {
			t.Errorf("%s is %d/%d, want %d/%d", findBattlefieldCardForTest(g, id).Name, p, tg, want[0], want[1])
		}
	}
}

func TestTrekB47BorgQueenAssimilatesFromAnOpponentsGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB15Seats(g)
	dead := pushToGraveyard(g, opp, game.Card{Name: "Fallen Knight", TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2})
	mine := pushToGraveyard(g, me, game.Card{Name: "My Knight", TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2})
	castFromHand(t, g, game.Card{Name: "Borg Queen, Perfection Manifest", TypeLine: "Legendary Artifact Creature — Borg Noble", OracleID: trekB47QueenOracle, ManaCost: "{4}{B}{B}", Power: 1, Toughness: 4}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Error("the trigger accepted a card from your own graveyard")
	}
	pickTriggerTarget(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	got := findBattlefieldCardForTest(g, dead)
	if got == nil {
		t.Fatal("the creature card was not put onto the battlefield")
	}
	if got.Controller != me.ID {
		t.Errorf("it entered under %v, want you", got.Controller)
	}
	if n := plusOneCounters(g, dead); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
	v := gaLayered(t, g, dead)
	if !v.IsArtifact() || !v.IsCreature() || !v.HasSubtype("Borg") || v.HasSubtype("Human") || v.HasSubtype("Knight") {
		t.Errorf("it is %q, want a Borg artifact creature with no other creature type", v.TypeLine)
	}
	// 2/2 printed, +1/+1 counter, +2/+0 from the Queen.
	if lc := layeredCard(t, g, dead); lc.CurrentPower() != 5 || lc.CurrentToughness() != 3 {
		t.Errorf("it is %d/%d, want 5/3", lc.CurrentPower(), lc.CurrentToughness())
	}
}

// --- Kirk, Enterprising Captain --------------------------------------------------

func TestTrekB47KirkExilesTheTopCardAtEightPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kirk := b27Push(g, me.ID, "Kirk, Enterprising Captain", "Legendary Creature — Human Officer", trekB47KirkOracle, "{3}{R}", 4, 3, "R")
	assertKeywords(t, g, kirk, "menace")
	b12Creature(g, me.ID, "Beast", "Creature — Beast", 4, 4)
	top := trekB47OnTop(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	trekB47Attack(t, g, kirk)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, top) != game.ZoneExile {
		t.Fatalf("the top card is in %v, want exile", b12ZoneOf(g, top))
	}
	if grantedPermissionOn(g, me.ID, top, game.ZoneExile) == nil {
		t.Error("you may not play the exiled card")
	}
}

func TestTrekB47KirkDoesNothingBelowEightPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kirk := b27Push(g, me.ID, "Kirk, Enterprising Captain", "Legendary Creature — Human Officer", trekB47KirkOracle, "{3}{R}", 4, 3, "R")
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	top := trekB47OnTop(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	trekB47Attack(t, g, kirk)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, top) != game.ZoneLibrary {
		t.Errorf("the top card is in %v at total power 7, want the library", b12ZoneOf(g, top))
	}
}

// --- Klingon Strike Force --------------------------------------------------------

func TestTrekB47KlingonStrikeForceDealsTwoToEachOpponentWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	ksf := b27Push(g, me.ID, "Klingon Strike Force", "Creature — Klingon Warrior", trekB47KlingonOracle, "{3}{R}{R}", 5, 5, "R")
	assertKeywords(t, g, ksf, "trample", "haste")
	before, mine := trekB47OpponentsLife(g, me.ID), me.Life
	b27Kill(g, ksf)
	passPriorityAroundTable(t, g)
	for i, life := range trekB47OpponentsLife(g, me.ID) {
		if life != before[i]-2 {
			t.Errorf("opponent %d at %d, want %d", i, life, before[i]-2)
		}
	}
	if me.Life != mine {
		t.Errorf("you are at %d, want %d", me.Life, mine)
	}
}
