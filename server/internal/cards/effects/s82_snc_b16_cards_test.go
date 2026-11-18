package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 — Streets of New Capenna, batch 16: graveyard recursion.

const (
	sncB16PatchUp      = "34a7026f-c6ca-4d89-8364-fd60fb613634"
	sncB16Gallery      = "7ea1b0e8-a577-47e2-b4cb-a6ca973742d0"
	sncB16Extraction   = "4164034a-5e59-4e40-a150-2c1000b0bd0d"
	sncB16Appraiser    = "bd3cdd9b-aa67-4bf5-80c8-53574d2c221e"
	sncB16LaundererOID = "eaca2576-020b-4af4-b348-6d1b56116c41"
	sncB16Cormela      = "3cd9bac9-0abc-43d8-9f84-94f965f7a2e0"
	sncB16Gorger       = "7271c442-2bd1-4d78-a049-b28bd5ee6876"
	sncB16Syrix        = "ec97bb0e-0dc8-4b9b-8762-2f1649cc9044"
	sncB16NextOfKin    = "5b079914-1302-40b5-8060-7c3e5b6e9eae"
	sncB16Jailbreak    = "169cf74a-07bf-4841-83f3-904df8a0a39b"
	sncB16Jalopy       = "469d820d-b7a6-414a-af18-71c1b1e42da5"
	sncB16SorceryType  = "Sorcery"
	sncB16CreatureType = "Creature — Human Rogue"
)

func TestSncB16EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncB16PatchUp:      "Patch Up",
		sncB16Gallery:      "Rogues' Gallery",
		sncB16Extraction:   "Extraction Specialist",
		sncB16Appraiser:    "Corpse Appraiser",
		sncB16LaundererOID: "Body Launderer",
		sncB16Cormela:      "Cormela, Glamour Thief",
		sncB16Gorger:       "Grime Gorger",
		sncB16Syrix:        "Syrix, Carrier of the Flame",
		sncB16NextOfKin:    "Next of Kin",
		sncB16Jailbreak:    "Jailbreak",
		sncB16Jalopy:       "Dodgy Jalopy",
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

// --- Patch Up ------------------------------------------------------------

func TestSncB16PatchUpReturnsThreeCreaturesWithinTheTotal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b38Graveyard(g, me.ID, "A", "Creature — Human", "{W}", 1, 1)
	b := b38Graveyard(g, me.ID, "B", "Creature — Human", "{W}", 1, 1)
	c := b38Graveyard(g, me.ID, "C", "Creature — Human", "{W}", 1, 1)
	castCatalogSpell(t, g, "Patch Up", sncB16SorceryType, sncB16PatchUp, cardRefs(a, b, c))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not return", id)
		}
	}
}

func TestSncB16PatchUpRefusesATotalAboveThreeAndOtherGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	two := b38Graveyard(g, me.ID, "Two", "Creature — Human", "{1}{W}", 2, 2)
	three := b38Graveyard(g, me.ID, "Three", "Creature — Human", "{2}{W}", 3, 3)
	theirs := b38Graveyard(g, opp.ID, "Theirs", "Creature — Human", "{W}", 1, 1)
	if err := castCatalogSpellErrStrict(t, g, "Patch Up", sncB16SorceryType, sncB16PatchUp, cardRefs(two, three)); err == nil {
		t.Error("total mana value 5 was accepted")
	}
	if err := castCatalogSpellErrStrict(t, g, "Patch Up", sncB16SorceryType, sncB16PatchUp, cardRefs(theirs)); err == nil {
		t.Error("a card in an opponent's graveyard was accepted")
	}
	if err := castCatalogSpellErrStrict(t, g, "Patch Up", sncB16SorceryType, sncB16PatchUp, cardRefs(two, theirs)); err == nil {
		t.Error("an opponent's card among legal ones was accepted")
	}
}

func TestSncB16PatchUpIgnoresNoncreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b38Graveyard(g, me.ID, "Rock", "Artifact", "{1}", 0, 0)
	if err := castCatalogSpellErrStrict(t, g, "Patch Up", sncB16SorceryType, sncB16PatchUp, cardRefs(rock)); err == nil {
		t.Error("a noncreature card was accepted")
	}
}

// --- Rogues' Gallery -----------------------------------------------------

func sncB16Dead(g *game.Game, owner uuid.UUID, name string, colors ...string) uuid.UUID {
	p := g.PlayerByID(owner)
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Human", Colors: colors,
		Power: 1, Toughness: 1, Owner: owner, Controller: owner})
	return id
}

func TestSncB16RoguesGalleryReturnsOnePerColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	w := sncB16Dead(g, me.ID, "W", "W")
	b := sncB16Dead(g, me.ID, "B", "B")
	b2 := sncB16Dead(g, me.ID, "B2", "B")
	multi := sncB16Dead(g, me.ID, "UR", "U", "R")
	castCatalogSpell(t, g, "Rogues' Gallery", sncB16SorceryType, sncB16Gallery, []game.TargetRef{
		{Kind: game.TargetCard, ID: w, Slot: 0},
		{Kind: game.TargetCard, ID: multi, Slot: 1},
		{Kind: game.TargetCard, ID: b, Slot: 2},
		{Kind: game.TargetCard, ID: multi, Slot: 3},
	})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{w, b, multi} {
		if !me.Hand.Contains(id) {
			t.Errorf("%s did not return to hand", id)
		}
	}
	if me.Hand.Contains(b2) || !me.Graveyard.Contains(b2) {
		t.Error("an untargeted black card moved")
	}
	n := 0
	for _, c := range me.Hand.Cards {
		if c.InstanceID == multi {
			n++
		}
	}
	if n != 1 {
		t.Errorf("a card targeted for two colours returned %d times", n)
	}
}

func TestSncB16RoguesGalleryAcceptsNoTargetsAndRefusesTheWrongColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	w := sncB16Dead(g, me.ID, "W", "W")
	if err := castCatalogSpellErrStrict(t, g, "Rogues' Gallery", sncB16SorceryType, sncB16Gallery,
		[]game.TargetRef{{Kind: game.TargetCard, ID: w, Slot: 2}}); err == nil {
		t.Error("a white card was accepted for the black clause")
	}
	castCatalogSpell(t, g, "Rogues' Gallery", sncB16SorceryType, sncB16Gallery, nil)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(w) {
		t.Error("a card moved with no targets chosen")
	}
}

// --- Corpse Appraiser ----------------------------------------------------

func TestSncB16CorpseAppraiserExilesACreatureCardThenLooksAtThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := sncB16Dead(g, opp.ID, "Dead", "G")
	ids := seedLibrary(me, "One", "Two", "Three", "Four")
	gyBefore := me.Graveyard.Size()
	_, err := trekB37CastFromHand(t, g, game.Card{Name: "Corpse Appraiser", TypeLine: "Creature — Vampire Rogue",
		OracleID: sncB16Appraiser, Power: 3, Toughness: 3, Colors: []string{"U", "B", "R"}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	cmdr2B07PickIfAsked(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) {
		t.Fatal("the creature card is still in the graveyard")
	}
	if !g.Exile.Contains(dead) {
		t.Error("the creature card was not exiled")
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the top three cards")
	}
	if len(pick.ChooseCards) != 3 || slices.Contains(pick.ChooseCards, ids[3]) {
		t.Fatalf("offered %v, want exactly the top three", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the chosen card is not in hand")
	}
	if got := me.Graveyard.Size() - gyBefore; got != 2 {
		t.Errorf("graveyard grew by %d, want the other two", got)
	}
	if !me.Library.Contains(ids[3]) {
		t.Error("the fourth card left the library")
	}
}

func TestSncB16CorpseAppraiserWithNoTargetDoesNotLook(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	libBefore := me.Library.Size()
	_, err := trekB37CastFromHand(t, g, game.Card{Name: "Corpse Appraiser", TypeLine: "Creature — Vampire Rogue",
		OracleID: sncB16Appraiser, Power: 3, Toughness: 3, Colors: []string{"U", "B", "R"}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != libBefore {
		t.Errorf("library %d → %d with no card exiled", libBefore, me.Library.Size())
	}
}

// --- Extraction Specialist -----------------------------------------------

func TestSncB16ExtractionSpecialistReturnsACardThatCantAttackOrBlock(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cheap := b38Graveyard(g, me.ID, "Cheap", "Creature — Human", "{1}{W}", 2, 2)
	dear := b38Graveyard(g, me.ID, "Dear", "Creature — Human", "{2}{W}", 3, 3)
	spec, err := trekB37CastFromHand(t, g, game.Card{Name: "Extraction Specialist", TypeLine: sncB16CreatureType,
		OracleID: sncB16Extraction, Power: 3, Toughness: 2, Colors: []string{"W"}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	cmdr2B07PickIfAsked(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the cheap creature did not return")
	}
	if !me.Graveyard.Contains(dear) {
		t.Error("a creature with mana value 3 was returned")
	}
	card := layeredCard(t, g, cheap)
	if !game.Restricted(&card, game.CantAttack) || !game.Restricted(&card, game.CantBlock) {
		t.Error("the returned creature can attack or block")
	}
	// Losing the Specialist frees it (CR 611.2b).
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, spec); err != nil {
		t.Fatal(err)
	}
	card = layeredCard(t, g, cheap)
	if game.Restricted(&card, game.CantAttack) || game.Restricted(&card, game.CantBlock) {
		t.Error("the restriction outlived the Specialist")
	}
}

func TestSncB16ExtractionSpecialistRefusesAnExpensiveCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dear := b38Graveyard(g, me.ID, "Dear", "Creature — Human", "{2}{W}", 3, 3)
	trekB37CastFromHand(t, g, game.Card{Name: "Extraction Specialist", TypeLine: sncB16CreatureType,
		OracleID: sncB16Extraction, Power: 3, Toughness: 2}, 0)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dear}); err == nil {
			t.Error("a mana value 3 creature was accepted")
		}
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(dear) {
		t.Error("a creature with mana value 3 returned")
	}
}

// --- Jailbreak -----------------------------------------------------------

func TestSncB16JailbreakReturnsTheirCardThenOneOfYoursOfEqualOrLesserValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b38Graveyard(g, opp.ID, "Theirs", "Creature — Human", "{1}{W}", 2, 2)
	cheap := b38Graveyard(g, me.ID, "Cheap", "Creature — Human", "{W}", 1, 1)
	equal := b38Graveyard(g, me.ID, "Equal", "Artifact", "{2}", 0, 0)
	dear := b38Graveyard(g, me.ID, "Dear", "Creature — Human", "{2}{W}", 3, 3)
	castCatalogSpell(t, g, "Jailbreak", sncB16SorceryType, sncB16Jailbreak, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Fatal("their card did not return")
	}
	if c := cardByID(g, theirs); c.Controller != opp.ID {
		t.Errorf("their card is controlled by %s, want its owner", c.Controller)
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dear}); err == nil {
			t.Error("a card with greater mana value was accepted")
		}
		pickCard(t, g, me.ID, equal)
	} else {
		t.Fatal("no target prompt for the second half")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(equal) {
		t.Error("the equal-value card did not return")
	}
	if !me.Graveyard.Contains(cheap) || !me.Graveyard.Contains(dear) {
		t.Error("an untargeted card left the graveyard")
	}
}

func TestSncB16JailbreakRefusesSpellsYourGraveyardAndNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := b38Graveyard(g, opp.ID, "Bolt", "Instant", "{R}", 0, 0)
	mine := b38Graveyard(g, me.ID, "Mine", "Creature — Human", "{W}", 1, 1)
	if err := castCatalogSpellErrStrict(t, g, "Jailbreak", sncB16SorceryType, sncB16Jailbreak, cardRefs(bolt)); err == nil {
		t.Error("an instant card was accepted")
	}
	if err := castCatalogSpellErrStrict(t, g, "Jailbreak", sncB16SorceryType, sncB16Jailbreak, cardRefs(mine)); err == nil {
		t.Error("a card in your own graveyard was accepted")
	}
}

// --- Body Launderer ------------------------------------------------------

func sncB16Launderer(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Body Launderer", "Creature — Ogre Rogue", sncB16LaundererOID, "{2}{B}{B}", 3, 3, "B")
}

func TestSncB16BodyLaundererConnivesWhenAnotherNontokenCreatureYouControlDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	adv := b27Push(g, me.ID, "Hanger-on", "Creature — Human", "", "{1}{B}", 1, 1, "B")
	sncB16LaundererPush := sncB16Launderer(g, me.ID)
	handBefore := me.Hand.Size()
	killCreature(t, g, me.ID, adv)
	passPriorityAroundTable(t, g)
	// Connive: draw a card, then a discard prompt.
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Fatalf("hand %d → %d, want a card drawn by connive", handBefore, got)
	}
	_ = sncB16LaundererPush
}

func TestSncB16BodyLaundererIgnoresItsOwnDeathOpponentsAndTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sncB16Launderer(g, me.ID)
	theirs := b27Push(g, opp.ID, "Theirs", "Creature — Human", "", "{1}{B}", 1, 1, "B")
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Token Creature — Spirit",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	handBefore := me.Hand.Size()
	killCreature(t, g, opp.ID, theirs)
	killCreature(t, g, me.ID, token)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand %d → %d, but only a nontoken creature of yours dying connives", handBefore, got)
	}
}

func TestSncB16BodyLaundererDiesReturnsANonRogueOfEqualOrLesserPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	launderer := sncB16Launderer(g, me.ID)
	ok := b38Graveyard(g, me.ID, "Fine", "Creature — Human", "{2}", 3, 3)
	rogue := b38Graveyard(g, me.ID, "Sneak", "Creature — Human Rogue", "{1}", 1, 1)
	big := b38Graveyard(g, me.ID, "Big", "Creature — Giant", "{5}", 4, 4)
	killCreature(t, g, me.ID, launderer)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		for _, bad := range []uuid.UUID{rogue, big} {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
				t.Errorf("an illegal card (%s) was accepted", bad)
			}
		}
		pickCard(t, g, me.ID, ok)
	} else {
		t.Fatal("no target prompt")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ok) {
		t.Error("the power-3 non-Rogue did not return")
	}
	if !me.Graveyard.Contains(rogue) || !me.Graveyard.Contains(big) {
		t.Error("an ineligible card left the graveyard")
	}
}

// --- Cormela, Glamour Thief ----------------------------------------------

func TestSncB16CormelaManaIsRestrictedToInstantsAndSorceries(t *testing.T) {
	spec, ok := Lookup(sncB16Cormela)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 || len(spec.ManaAbilities) != 1 {
		t.Fatalf("registered %v, spec %+v", ok, spec)
	}
	ma := spec.ManaAbilities[0]
	if ma.Produced != "{U}{B}{R}" || ma.Cost.Mana != "{1}" || !ma.Cost.Tap {
		t.Errorf("mana ability %+v, want {1},{T}: Add {U}{B}{R}", ma)
	}
	want := map[string]bool{ManaRestrictCast: false}
	for _, r := range ma.Restrictions {
		if _, ok := want[r]; ok {
			want[r] = true
		}
	}
	if !want[ManaRestrictCast] || len(ma.Restrictions) < 2 {
		t.Errorf("restrictions %v, want cast-only and instant/sorcery", ma.Restrictions)
	}
}

func TestSncB16CormelaDiesReturnsAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cormela := b27Push(g, me.ID, "Cormela, Glamour Thief", "Legendary Creature — Vampire Rogue", sncB16Cormela, "{1}{U}{B}{R}", 2, 4, "U", "B", "R")
	bolt := b38Graveyard(g, me.ID, "Bolt", "Instant", "{R}", 0, 0)
	bear := b38Graveyard(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	killCreature(t, g, me.ID, cormela)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Error("a creature card was accepted")
		}
		pickCard(t, g, me.ID, bolt)
	} else {
		t.Fatal("no target prompt")
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant did not return to hand")
	}
}

// --- Grime Gorger --------------------------------------------------------

func sncB16Gorge(t *testing.T, g *game.Game, attacker, defender *game.Player) (uuid.UUID, *game.PendingChoice) {
	t.Helper()
	gorger := b27Push(g, attacker.ID, "Grime Gorger", "Creature — Horror", sncB16Gorger, "{2}{B}{G}", 3, 3, "B", "G")
	declareAttack(t, g, defender.ID, gorger)
	passPriorityAroundTable(t, g)
	return gorger, latestChooseCardsFor(g, attacker.ID)
}

func sncB16Typed(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return b38Graveyard(g, owner, name, typeLine, "{1}", 1, 1)
}

func TestSncB16GrimeGorgerExilesOneCardOfEachTypeAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creature := sncB16Typed(g, opp.ID, "Bear", "Creature — Bear")
	creature2 := sncB16Typed(g, opp.ID, "Bear2", "Creature — Bear")
	instant := sncB16Typed(g, opp.ID, "Bolt", "Instant")
	land := sncB16Typed(g, opp.ID, "Forest", "Basic Land — Forest")
	mine := sncB16Typed(g, me.ID, "Mine", "Instant")
	gorger, pick := sncB16Gorge(t, g, me, opp)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if slices.Contains(pick.ChooseCards, mine) || len(pick.ChooseCards) != 4 {
		t.Fatalf("offered %v, want only the defending player's four cards", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{creature, creature2}); err == nil {
		t.Error("two creature cards were accepted as one of each type")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{creature, instant, land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{creature, instant, land} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
	}
	if !opp.Graveyard.Contains(creature2) || !me.Graveyard.Contains(mine) {
		t.Error("an unchosen card left a graveyard")
	}
	if n := countersOn(g, gorger, game.CounterPlusOne); n != 3 {
		t.Errorf("%d +1/+1 counters, want 3", n)
	}
}

func TestSncB16GrimeGorgerCountsAMultiTypedCardForOneType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := sncB16Typed(g, opp.ID, "Golem", "Artifact Creature — Golem")
	b := sncB16Typed(g, opp.ID, "Golem2", "Artifact Creature — Golem")
	c := sncB16Typed(g, opp.ID, "Golem3", "Artifact Creature — Golem")
	_, pick := sncB16Gorge(t, g, me, opp)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b, c}); err == nil {
		t.Error("three artifact creature cards were accepted for two types")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Errorf("two artifact creature cards (artifact and creature) were refused: %v", err)
	}
}

func TestSncB16GrimeGorgerWithAnEmptyGraveyardAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gorger, pick := sncB16Gorge(t, g, me, opp)
	if pick != nil {
		t.Error("asked to pick from an empty graveyard")
	}
	if n := countersOn(g, gorger, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters with nothing exiled", n)
	}
}

// --- Syrix, Carrier of the Flame -----------------------------------------

func sncB16SyrixPush(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Syrix, Carrier of the Flame", "Legendary Creature — Phoenix", sncB16Syrix, "{2}{B}{R}", 3, 3, "B", "R")
}

func TestSncB16SyrixPhoenixDealsItsPowerAtEndStepAfterACreatureLeftYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	syrix := sncB16SyrixPush(g, me.ID)
	dead := sncB16Dead(g, me.ID, "Dead", "G")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, dead); err != nil {
		t.Fatal(err)
	}
	before := opp.Life
	advanceToEndStep(t, g)
	pickCard(t, g, me.ID, syrix)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life %d → %d, want 3 from Syrix's power", before, opp.Life)
	}
}

func TestSncB16SyrixDoesNothingWithoutACreatureCardLeavingYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sncB16SyrixPush(g, me.ID)
	// A noncreature card leaving, and a creature card leaving an opponent's graveyard, do not count.
	bolt := b38Graveyard(g, me.ID, "Bolt", "Instant", "{R}", 0, 0)
	theirs := sncB16Dead(g, opp.ID, "Theirs", "G")
	for _, id := range []uuid.UUID{bolt, theirs} {
		owner := me.ID
		if id == theirs {
			owner = opp.ID
		}
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
			t.Fatal(err)
		}
	}
	before := opp.Life
	advanceToEndStep(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the end-step trigger asked for targets without its condition")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("life %d → %d", before, opp.Life)
	}
}

func TestSncB16SyrixOffersTheCastWhenAnotherPhoenixYouControlDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	syrix := pushCatalogGraveyardCard(me, "Syrix, Carrier of the Flame", "Legendary Creature — Phoenix", sncB16Syrix, 3, 3)
	phoenix := b27Push(g, me.ID, "Bird", "Creature — Phoenix", "", "{1}{R}", 2, 2, "R")
	bear := b27Push(g, me.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Stack.Size() != 0 || len(g.PendingChoices) != 0 {
		t.Fatal("a Bear dying offered Syrix")
	}
	killCreature(t, g, me.ID, phoenix)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) == 0 {
		t.Fatal("no cast offer when another Phoenix died")
	}
	_ = syrix
}

// --- Next of Kin ---------------------------------------------------------

func TestSncB16NextOfKinPutsACheaperCreatureOntoTheBattlefieldAndReturnsAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := b27Push(g, me.ID, "Host", "Creature — Human", "", "{2}{G}", 2, 2, "G")
	aura := auraCast(t, g, "Next of Kin", sncB16NextOfKin, host)
	cheap := sncB16HandCreature(me, "Cheap", "{1}{G}")
	same := sncB16HandCreature(me, "Same", "{2}{G}")
	killCreature(t, g, me.ID, host)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if !slices.Contains(pick.ChooseCards, cheap) || slices.Contains(pick.ChooseCards, same) {
		t.Fatalf("offered %v, want only the cheaper creature card", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{cheap}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the creature card was not put onto the battlefield")
	}
	if g.Battlefield.Contains(aura) {
		t.Fatal("the Aura came back at once")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not return at the next end step")
	}
	if h := attachmentHostOf(t, g, aura); h.ID != cheap {
		t.Errorf("attached to %v, want the creature put onto the battlefield", h.ID)
	}
}

func TestSncB16NextOfKinDeclinedLeavesTheAuraInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := b27Push(g, me.ID, "Host", "Creature — Human", "", "{2}{G}", 2, 2, "G")
	aura := auraCast(t, g, "Next of Kin", sncB16NextOfKin, host)
	cheap := sncB16HandCreature(me, "Cheap", "{1}{G}")
	killCreature(t, g, me.ID, host)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) || g.Battlefield.Contains(aura) || !me.Graveyard.Contains(aura) {
		t.Error("declining still returned something")
	}
}

func TestSncB16NextOfKinCanPutACommanderFromTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := b27Push(g, me.ID, "Host", "Creature — Human", "", "{3}{G}", 2, 2, "G")
	auraCast(t, g, "Next of Kin", sncB16NextOfKin, host)
	cmdr := uuid.New()
	me.Command.PushTop(game.Card{InstanceID: cmdr, Name: "Boss", TypeLine: "Legendary Creature — Human", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true})
	killCreature(t, g, me.ID, host)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || !slices.Contains(pick.ChooseCards, cmdr) {
		t.Fatalf("the commander was not offered: %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{cmdr}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cmdr) {
		t.Error("the commander did not enter")
	}
}

func sncB16HandCreature(p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Human", ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Dodgy Jalopy --------------------------------------------------------

func TestSncB16DodgyJalopyPowerIsTheGreatestManaValueAmongYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jalopy := b27Push(g, me.ID, "Dodgy Jalopy", "Artifact — Vehicle", sncB16Jalopy, "{2}{G}", 0, 5)
	if p := effectivePower(t, g, jalopy); p != 0 {
		t.Errorf("power %d with no creatures, want 0", p)
	}
	b27Push(g, me.ID, "Two", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	b27Push(g, me.ID, "Five", "Creature — Giant", "", "{3}{G}{G}", 5, 5, "G")
	b27Push(g, opp.ID, "Nine", "Creature — Giant", "", "{9}", 9, 9)
	if p := effectivePower(t, g, jalopy); p != 5 {
		t.Errorf("power %d, want 5 (the opponent's creature does not count)", p)
	}
}

func TestSncB16DodgyJalopyScavengePutsCountersEqualToItsPowerFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := b27Push(g, me.ID, "Four", "Creature — Giant", "", "{2}{G}{G}", 4, 4, "G")
	advanceToMain(t, g)
	id := pushCatalogGraveyardCard(me, "Dodgy Jalopy", "Artifact — Vehicle", sncB16Jalopy, 0, 5)
	fillPoolColored(me, "G", 3)
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	}); err != nil {
		t.Fatalf("scavenge: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("the cost did not exile the Jalopy")
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, big, game.CounterPlusOne); n != 4 {
		t.Errorf("%d +1/+1 counters, want 4 (the greatest mana value among your creatures)", n)
	}
}

func TestSncB16DodgyJalopyScavengeWithNoCreaturesAddsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b27Push(g, opp.ID, "Nine", "Creature — Giant", "", "{9}", 9, 9)
	advanceToMain(t, g)
	id := pushCatalogGraveyardCard(me, "Dodgy Jalopy", "Artifact — Vehicle", sncB16Jalopy, 0, 5)
	fillPoolColored(me, "G", 3)
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("scavenge: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, theirs, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters, want 0 with no creature of yours", n)
	}
}
