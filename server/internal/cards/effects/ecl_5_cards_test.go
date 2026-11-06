package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ecl_5_cards_test.go — Lorwyn Eclipsed, fifth batch (#4458). Each test
// plays the real catalog entry and asserts what its text says.

const (
	ecl5MerrowOracle      = "a3a00b62-f07e-44ef-8d97-22a3e7a06f72"
	ecl5AuntieOracle      = "079f0d7f-57a5-4101-a9e3-c5a57148b3fb"
	ecl5ReejereyOracle    = "7668f7bc-9a43-4024-87c3-995aa0a65aaa"
	ecl5FlamecasterOracle = "e35d7886-8c03-471c-b598-118d7eceddb4"
	ecl5ProdigyOracle     = "22d059fc-4e47-44b5-9e21-4a56f1dff6a7"
	ecl5SpikelingOracle   = "fcf796fa-43ae-43d1-a124-776338bc3542"
	ecl5MaulerOracle      = "d6eef6a7-0cf4-4ee8-90e0-727a283c3be9"
	ecl5GildweaverOracle  = "bc8427fb-e5ee-4b44-9a4f-6d9f01b4b896"
	ecl5CinderOracle      = "f70d65bf-5f1c-4aee-9bd1-f5ee6974b058"
	ecl5ImpostorOracle    = "6522070b-7002-44fd-8deb-832fe67ef9a9"
)

func TestEcl5EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		ecl5MerrowOracle:      "Eclipsed Merrow",
		ecl5AuntieOracle:      "Elder Auntie",
		ecl5ReejereyOracle:    "Encumbered Reejerey",
		ecl5FlamecasterOracle: "Enraged Flamecaster",
		ecl5ProdigyOracle:     "Explosive Prodigy",
		ecl5SpikelingOracle:   "Feisty Spikeling",
		ecl5MaulerOracle:      "Flame-Chain Mauler",
		ecl5GildweaverOracle:  "Flamekin Gildweaver",
		ecl5CinderOracle:      "Flaring Cinder",
		ecl5ImpostorOracle:    "Flock Impostor",
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

// ecl5Cast casts a creature with a mana cost and colours from the active
// seat's hand and passes priority until the stack is empty or a prompt
// opens.
func ecl5Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, colors []string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Colors: colors, Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	toMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// ecl5Card reads a battlefield card, failing the test if it is gone.
func ecl5Card(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c
}

// --- Eclipsed Merrow ------------------------------------------------------

func TestEcl5EclipsedMerrowTakesAMerfolkPlainsOrIslandFromTheTopFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	// Fifth from the top: a Plains the look never reaches.
	deep := hornTopFive(me, []struct{ name, typeLine string }{{"Deep Plains", "Basic Land — Plains"}})[0]
	top := hornTopFive(me, []struct{ name, typeLine string }{
		{"Shock", "Instant"},
		{"Merfolk Card", "Creature — Merfolk Wizard"},
		{"Bear", "Creature — Bear"},
		{"Island", "Basic Land — Island"},
	})
	ecl5Cast(t, g, "Eclipsed Merrow", "Creature — Merfolk Scout", "{W/U}{W/U}{W/U}", ecl5MerrowOracle, []string{"W", "U"}, 2, 3)

	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("the entry trigger raised no prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range prompt.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[top[1]] || !offered[top[3]] {
		t.Errorf("offered %v; want only the Merfolk card and the Island of the top four", prompt.ChooseCards)
	}
	if offered[deep] {
		t.Error("the fifth card was offered: the look is the top four")
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Errorf("bounds = [%d,%d], want an optional single take", prompt.ChooseMin, prompt.ChooseMax)
	}
	if latestChooseCards(g, opp.ID) != nil {
		t.Error("the opponent was asked about the Merrow's look")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[3]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top[3]) {
		t.Error("the Island did not go to hand")
	}
	if !hornRevealed(g, top[3]) {
		t.Error("the card taken was not revealed")
	}
	// The other three are now the bottom three of the library.
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:3] {
		bottom[c.InstanceID] = true
	}
	for _, i := range []int{0, 1, 2} {
		if !bottom[top[i]] {
			t.Errorf("%s did not go to the bottom", []string{"Shock", "Merfolk Card", "Bear"}[i])
		}
	}
	if !inLibrary(g, me.ID, deep) {
		t.Error("the fifth card left the library")
	}
}

// --- Elder Auntie ---------------------------------------------------------

func TestEcl5ElderAuntieMakesABlackAndRedGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ecl5Cast(t, g, "Elder Auntie", "Creature — Goblin Warlock", "{2}{R}", ecl5AuntieOracle, []string{"R"}, 2, 2)
	if got := ecl1Named(g, me.ID, "Goblin"); got != 1 {
		t.Fatalf("Goblin tokens = %d, want 1", got)
	}
	goblin := ecl5Card(t, g, findBattlefieldByName(g, "Goblin"))
	if !goblin.IsToken() || !goblin.IsCreature() || !goblin.HasSubtype("Goblin") {
		t.Errorf("the token is %q, want a Goblin creature token", goblin.TypeLine)
	}
	if p, tt := ptOf(t, g, goblin.InstanceID); p != 1 || tt != 1 {
		t.Errorf("the Goblin is %d/%d, want 1/1", p, tt)
	}
	if cols := goblin.EffectiveColors(); len(cols) != 2 || !slices.Contains(cols, "B") || !slices.Contains(cols, "R") {
		t.Errorf("the Goblin's colours are %v, want black and red", cols)
	}
}

// --- Encumbered Reejerey --------------------------------------------------

func TestEcl5EncumberedReejereyShedsACounterEachTimeItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	r := ecl5Cast(t, g, "Encumbered Reejerey", "Creature — Merfolk Soldier", "{1}{W}", ecl5ReejereyOracle, []string{"W"}, 5, 4)
	if got := countersOn(g, r, game.CounterMinusOne); got != 3 {
		t.Fatalf("entered with %d -1/-1 counters, want 3", got)
	}
	if p, tt := ptOf(t, g, r); p != 2 || tt != 1 {
		t.Errorf("the Reejerey is %d/%d, want 2/1", p, tt)
	}
	for want := 2; want >= 0; want-- {
		untapForTest(g, r)
		tapForTest(t, g, r)
		passPriorityAroundTable(t, g)
		if got := countersOn(g, r, game.CounterMinusOne); got != want {
			t.Fatalf("after a tap: %d counters, want %d", got, want)
		}
	}
	if p, tt := ptOf(t, g, r); p != 5 || tt != 4 {
		t.Errorf("with no counters the Reejerey is %d/%d, want 5/4", p, tt)
	}
	// No -1/-1 counter left: becoming tapped does not trigger.
	untapForTest(g, r)
	tapForTest(t, g, r)
	if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("PassPriority: %v", err)
	}
	if triggerOnStack(g, r) != nil {
		t.Error("the Reejerey triggered with no -1/-1 counter on it")
	}
}

// Attacking taps it, which is "becomes tapped".
func TestEcl5EncumberedReejereyShedsACounterWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	r := b12Push(g, me.ID, "Encumbered Reejerey", "Creature — Merfolk Soldier", ecl5ReejereyOracle, 5, 4)
	setCountersOn(g, r, map[string]int{game.CounterMinusOne: 2})
	declareAttack(t, g, opp.ID, r)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, r, game.CounterMinusOne); got != 1 {
		t.Errorf("after attacking: %d counters, want 1", got)
	}
}

// --- Enraged Flamecaster --------------------------------------------------

func TestEcl5EnragedFlamecasterBurnsEachOpponentForBigSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fc := b27Push(g, me.ID, "Enraged Flamecaster", "Creature — Elemental Sorcerer", ecl5FlamecasterOracle, "{2}{R}", 3, 2, "R")
	assertKeywords(t, g, fc, "reach")
	life, opps := me.Life, lifeOfOpponents(g)
	mscM2CastArtifact(t, g, "{3}", "{C}{C}{C}")
	if got := lifeOfOpponents(g); !slices.Equal(got, opps) {
		t.Fatalf("a mana value 3 spell changed the opponents' life: %v → %v", opps, got)
	}
	mscM2CastArtifact(t, g, "{4}", "{C}{C}{C}{C}")
	got := lifeOfOpponents(g)
	for i := range got {
		if got[i] != opps[i]-2 {
			t.Errorf("opponent %d: life %d after a mana value 4 spell, want %d", i, got[i], opps[i]-2)
		}
	}
	if me.Life != life {
		t.Errorf("my life changed: %d → %d", life, me.Life)
	}
}

// --- Explosive Prodigy ----------------------------------------------------

func TestEcl5ExplosiveProdigyDealsDamageForEachColorAmongYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b27Push(g, me.ID, "White Thing", "Artifact", "", "{W}", 0, 0, "W")
	b27Push(g, me.ID, "Blue Thing", "Artifact", "", "{U}", 0, 0, "U")
	b27Push(g, me.ID, "Second White Thing", "Artifact", "", "{W}", 0, 0, "W")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Beast", 5, 5)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	prodigy := ecl5Cast(t, g, "Explosive Prodigy", "Creature — Elemental Sorcerer", "{1}{R}", ecl5ProdigyOracle, []string{"R"}, 1, 1)

	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if hasID(pick.PickTargetCards, mine) || hasID(pick.PickTargetCards, prodigy) {
		t.Error("a creature you control is not a legal target")
	}
	if !hasID(pick.PickTargetCards, big) || !hasID(pick.PickTargetCards, small) {
		t.Error("each creature an opponent controls is a legal target")
	}
	pickCard(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	// White, blue and red: three colours, the second white one adds none.
	if got := damageMarkedOn(g, big); got != 3 {
		t.Errorf("damage on the target = %d, want 3", got)
	}
	if got := damageMarkedOn(g, small); got != 0 {
		t.Errorf("damage on the other creature = %d, want 0", got)
	}
}

// --- Feisty Spikeling -----------------------------------------------------

func TestEcl5FeistySpikelingHasFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Push(g, me.ID, "Feisty Spikeling", "Creature — Shapeshifter", ecl5SpikelingOracle, 2, 1)
	theirs := b12Push(g, opp.ID, "Feisty Spikeling", "Creature — Shapeshifter", ecl5SpikelingOracle, 2, 1)
	if !hasEffectiveKeyword(t, g, mine, "first strike") {
		t.Error("on its controller's turn the Spikeling has no first strike")
	}
	if hasEffectiveKeyword(t, g, theirs, "first strike") {
		t.Error("on another player's turn the Spikeling has first strike")
	}
	c := ecl5Card(t, g, mine)
	for _, sub := range []string{"Goblin", "Kithkin", "Merfolk", "Elf"} {
		if !c.HasSubtype(sub) {
			t.Errorf("changeling: the Spikeling is not a %s", sub)
		}
	}
}

// --- Flame-Chain Mauler ---------------------------------------------------

func TestEcl5FlameChainMaulerGetsPlusOneAndMenaceEachActivation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	m := b12Push(g, me.ID, "Flame-Chain Mauler", "Creature — Elemental Warrior", ecl5MaulerOracle, 2, 2)
	if hasEffectiveKeyword(t, g, m, "menace") {
		t.Fatal("menace before any activation")
	}
	for want := 3; want <= 4; want++ {
		if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		if p, tt := ptOf(t, g, m); p != want || tt != 2 {
			t.Errorf("the Mauler is %d/%d, want %d/2", p, tt, want)
		}
		if !hasEffectiveKeyword(t, g, m, "menace") {
			t.Error("no menace after the activation")
		}
	}
}

// --- Flamekin Gildweaver --------------------------------------------------

func TestEcl5FlamekinGildweaverMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gw := ecl5Cast(t, g, "Flamekin Gildweaver", "Creature — Elemental Sorcerer", "{3}{R}", ecl5GildweaverOracle, []string{"R"}, 4, 3)
	if !hasEffectiveKeyword(t, g, gw, "trample") {
		t.Error("no trample")
	}
	if got := ecl1Named(g, me.ID, "Treasure"); got != 1 {
		t.Fatalf("Treasure tokens = %d, want 1", got)
	}
	tr := ecl5Card(t, g, findBattlefieldByName(g, "Treasure"))
	if !tr.IsToken() || !tr.IsArtifact() || !tr.HasSubtype("Treasure") {
		t.Errorf("the token is %q, want a Treasure artifact token", tr.TypeLine)
	}
}

// --- Flaring Cinder -------------------------------------------------------

func TestEcl5FlaringCinderRummagesOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Instant"})
	fodder := handCard(me, "Fodder", "Instant")
	ecl5Cast(t, g, "Flaring Cinder", "Creature — Elemental Sorcerer", "{1}{U/R}{U/R}", ecl5CinderOracle, []string{"U", "R"}, 3, 2)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt on entry")
	}
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(fodder) {
		t.Error("the discarded card is still in hand")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d: one discarded and one drawn should leave it the same", hand, me.Hand.Size())
	}
}

func TestEcl5FlaringCinderRummagesOnlyForSpellsOfManaValueFourOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Flaring Cinder", "Creature — Elemental Sorcerer", ecl5CinderOracle, "{1}{U/R}{U/R}", 3, 2, "U", "R")
	handCard(me, "Fodder", "Instant")
	mscM2CastArtifact(t, g, "{3}", "{C}{C}{C}")
	if discardChoiceFor(g, me.ID) != nil {
		t.Fatal("a mana value 3 spell asked for a discard")
	}
	mscM2CastArtifact(t, g, "{4}", "{C}{C}{C}{C}")
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("a mana value 4 spell asked for no discard")
	}
	// Declining the discard draws nothing.
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d after declining: no card is drawn without a discard", hand, me.Hand.Size())
	}
}

// --- Flock Impostor -------------------------------------------------------

func TestEcl5FlockImpostorReturnsAnotherCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	elk := b12Creature(g, me.ID, "My Elk", "Creature — Elk", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	imp := ecl5Cast(t, g, "Flock Impostor", "Creature — Shapeshifter", "{2}{W}", ecl5ImpostorOracle, []string{"W"}, 2, 2)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if hasID(pick.PickTargetCards, imp) || hasID(pick.PickTargetCards, theirs) || hasID(pick.PickTargetCards, rock) {
		t.Error("the Impostor itself, an opponent's creature and a noncreature are not legal targets")
	}
	if !hasID(pick.PickTargetCards, bear) || !hasID(pick.PickTargetCards, elk) {
		t.Error("another creature you control is a legal target")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the chosen creature did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(elk) || !g.Battlefield.Contains(imp) {
		t.Error("only the target leaves")
	}
	assertKeywords(t, g, imp, "flash", "flying", game.KeywordChangeling)
}

// With no other creature to return, it enters and nothing happens.
func TestEcl5FlockImpostorWithNoOtherCreatureJustEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	imp := ecl5Cast(t, g, "Flock Impostor", "Creature — Shapeshifter", "{2}{W}", ecl5ImpostorOracle, []string{"W"}, 2, 2)
	if pick := latestPickTarget(g, me.ID); pick != nil && hasID(pick.PickTargetCards, imp) {
		t.Error("the Impostor was offered as its own target")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(imp) {
		t.Error("the Impostor returned itself")
	}
}
