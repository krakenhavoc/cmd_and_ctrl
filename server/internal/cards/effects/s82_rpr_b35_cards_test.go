package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b35_cards_test.go — S82 reprint pass, batch 35 (#5908).

const (
	rprMoonsilverSpearOracle   = "d99bcf62-db93-40db-8844-3d0af0756a0c"
	rprMurderOfCrowsOracle     = "babab2f5-fd1f-4e0a-a692-6c2ce35df708"
	rprNimbusMazeOracle        = "d7e1d4eb-1d4e-460e-9304-7db9ab50ccb5"
	rprObsessiveStitcherOracle = "e8210145-a6ac-4f84-a74c-a5799c5e9434"
	rprPhthisisOracle          = "530f088d-bfa1-4410-b1aa-ade8abdb2b7a"
	rprPiperOfTheSwarmOracle   = "a6c5c5ea-0bd0-4d5a-bd3f-6217f14c9939"
	rprPsychicImpetusOracle    = "47214acc-08c4-4dcb-9a8b-09196ce41045"
	rprRileOracle              = "0c4e0e46-77f1-4ff6-afbc-1a67feceedfb"
	rprRiverwiseAugurOracle    = "eac79f15-3b63-457a-b6a8-f9237439c93c"
	rprShieldingPlaxOracle     = "e9bdb194-745d-4abb-9312-a65a68c40a18"
	rprShivanHellkiteOracle    = "069be6a0-9545-4fdc-ad40-957fe307db7a"
)

func TestRprB35CardsAreRegisteredFull(t *testing.T) {
	want := map[string]string{
		rprMoonsilverSpearOracle:   "Moonsilver Spear",
		rprMurderOfCrowsOracle:     "Murder of Crows",
		rprNimbusMazeOracle:        "Nimbus Maze",
		rprObsessiveStitcherOracle: "Obsessive Stitcher",
		rprPhthisisOracle:          "Phthisis",
		rprPiperOfTheSwarmOracle:   "Piper of the Swarm",
		rprPsychicImpetusOracle:    "Psychic Impetus",
		rprRileOracle:              "Rile",
		rprRiverwiseAugurOracle:    "Riverwise Augur",
		rprShieldingPlaxOracle:     "Shielding Plax",
		rprShivanHellkiteOracle:    "Shivan Hellkite",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered as %q, completeness %v", name, spec.Name, spec.Completeness)
		}
		if len(spec.Caveats) != 0 {
			t.Errorf("%s ships caveats %v", name, spec.Caveats)
		}
	}
}

// --- Riverwise Augur ---------------------------------------------------------

func TestRprB35RiverwiseAugurDrawsThreeThenPutsTwoBackInOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	handBefore := me.Hand.Size()
	castCatalogSpell(t, g, "Riverwise Augur", "Creature — Merfolk Wizard", rprRiverwiseAugurOracle, nil)
	passPriorityAroundTable(t, g)
	// The creature left the hand; the trigger drew three.
	if me.Hand.Size() != handBefore+3 {
		t.Fatalf("hand = %d after the draw, want %d", me.Hand.Size(), handBefore+3)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to pick the two cards to put back")
	}
	first, second := me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{first, second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil || len(order.ScryCards) != 2 {
		t.Fatalf("\"in any order\" is its own prompt: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{second, first}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand = %d at the end, want %d", me.Hand.Size(), handBefore+1)
	}
	if top, _ := me.Library.Top(); top.InstanceID != second {
		t.Error("the first card of the answer is not on top")
	}
}

// --- Rile --------------------------------------------------------------------

func TestRprB35RileDamagesYourCreatureGivesTrampleAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Rile", "Sorcery", rprRileOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, bear); got != 1 {
		t.Errorf("damage on the bear = %d, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the bear did not gain trample")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d (Rile left the hand, one card drawn)", me.Hand.Size(), hand+1)
	}
}

func TestRprB35RileKillsAOneToughnessCreatureAndStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	mouse := b12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Rile", "Sorcery", rprRileOracle, b16TargetCard(mouse))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mouse) {
		t.Error("the 1/1 survived 1 damage")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestRprB35RileCantTargetAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Rile", "Sorcery", rprRileOracle, b16TargetCard(theirs)); err == nil {
		t.Fatal("Rile accepted a creature you don't control")
	}
}

// --- Shivan Hellkite ---------------------------------------------------------

func TestRprB35ShivanHellkitePingsAnyTargetForOneRed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	dragon := pushCatalogPermanent(g, me.ID, "Shivan Hellkite", "Creature — Dragon", rprShivanHellkiteOracle, true)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Fatal("activated with no mana")
	}
	b06AddMana(me, "R", "C")
	b16Activate(t, g, me.ID, dragon, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want one damage", life, opp.Life)
	}
}

// --- Nimbus Maze -------------------------------------------------------------

func TestRprB35NimbusMazeColoursNeedTheOtherBasicType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	maze := seedPermanentWithOracle(g, me.ID, "Nimbus Maze", "Land", rprNimbusMazeOracle)
	if err := g.ActivateManaAbility(me.ID, maze, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("{W} with no Island")
	}
	if err := g.ActivateManaAbility(me.ID, maze, 2, game.ManaAbilityParams{}); err == nil {
		t.Fatal("{U} with no Plains")
	}
	if err := g.ActivateManaAbility(me.ID, maze, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{C}: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool = %v, want {C}", got)
	}

	me.ManaPool = nil
	maze2 := seedPermanentWithOracle(g, me.ID, "Nimbus Maze", "Land", rprNimbusMazeOracle)
	seedManaLand(g, me.ID, "Island", "Basic Land — Island", "U")
	if err := g.ActivateManaAbility(me.ID, maze2, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{W} beside an Island: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"W"}) {
		t.Errorf("pool = %v, want {W}", got)
	}
	if err := g.ActivateManaAbility(me.ID, maze2, 2, game.ManaAbilityParams{}); err == nil {
		t.Error("{U} beside only an Island (and a tapped Maze)")
	}

	me.ManaPool = nil
	maze3 := seedPermanentWithOracle(g, me.ID, "Nimbus Maze", "Land", rprNimbusMazeOracle)
	seedManaLand(g, me.ID, "Plains", "Basic Land — Plains", "W")
	if err := g.ActivateManaAbility(me.ID, maze3, 2, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{U} beside a Plains: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool = %v, want {U}", got)
	}
}

// --- Shielding Plax ----------------------------------------------------------

func TestRprB35ShieldingPlaxDrawsAndBarsTheAurasControllersOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	// An enchanted creature that belongs to an OPPONENT: still barred to
	// that opponent, because "your opponents" is the Aura controller's.
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Other Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Shielding Plax", "Enchantment — Aura", rprShieldingPlaxOracle, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d (Aura cast, one card drawn)", me.Hand.Size(), hand+1)
	}
	spec := TargetCreature("target creature")
	if a109p6Has(a109p6LegalCards(g, opp.ID, spec), theirs) {
		t.Error("the enchanted creature's own controller (an opponent of the Aura's controller) can target it")
	}
	if !a109p6Has(a109p6LegalCards(g, opp.ID, spec), other) {
		t.Error("an unenchanted creature is barred")
	}
	if !a109p6Has(a109p6LegalCards(g, me.ID, spec), theirs) {
		t.Error("the Aura's controller can't target the enchanted creature")
	}
	if a109p6Has(a109p6LegalCards(g, g.Seats[2].ID, spec), theirs) {
		t.Error("another opponent of the Aura's controller can target it")
	}
}

// --- Psychic Impetus ---------------------------------------------------------

func TestRprB35PsychicImpetusPumpsGoadsAndScriesOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	auraCast(t, g, "Psychic Impetus", rprPsychicImpetusOracle, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("enchanted bear is %d/%d, want 4/4", p, tg)
	}
	if c, ok := g.LookupCardForEffect(bear); !ok || !c.Goaded() {
		t.Error("the enchanted creature is not goaded")
	}
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("attacking with the enchanted creature queued no scry prompt")
	}
}

// --- Murder of Crows ---------------------------------------------------------

func TestRprB35MurderOfCrowsLootsWhenAnotherCreatureDiesAndMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	msc18Library(me, 8)
	pushCatalogPermanent(g, me.ID, "Murder of Crows", "Creature — Bird", rprMurderOfCrowsOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	b27Kill(g, bear)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand = %d after drawing, want %d", me.Hand.Size(), hand+1)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand = %d after the discard, want %d", me.Hand.Size(), hand)
	}

	// Declined: nothing is drawn or discarded.
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	b27Kill(g, wolf)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("declining changed the hand: %d, want %d", me.Hand.Size(), hand)
	}
}

func TestRprB35MurderOfCrowsIgnoresItsOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crows := pushCatalogPermanent(g, me.ID, "Murder of Crows", "Creature — Bird", rprMurderOfCrowsOracle, false)
	b27Kill(g, crows)
	if len(g.PendingChoices) != 0 {
		t.Errorf("its own death raised a prompt: %+v", g.PendingChoices)
	}
}

// --- Phthisis ----------------------------------------------------------------

func TestRprB35PhthisisDestroysAndDrainsPowerPlusToughness(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 4)
	life := opp.Life
	castCatalogSpell(t, g, "Phthisis", "Sorcery", rprPhthisisOracle, b16TargetCard(ogre))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre survived")
	}
	if opp.Life != life-7 {
		t.Errorf("life %d → %d, want 3+4 = 7 lost by the creature's controller", life, opp.Life)
	}
}

func TestRprB35PhthisisDrainsEvenWhenTheCreatureIsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	golem := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 2, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	life := opp.Life
	castCatalogSpell(t, g, "Phthisis", "Sorcery", rprPhthisisOracle, b16TargetCard(golem))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(golem) {
		t.Error("an indestructible Golem was destroyed")
	}
	if opp.Life != life-4 {
		t.Errorf("life %d → %d, want 4 lost", life, opp.Life)
	}
}

func TestRprB35PhthisisCanBeSuspendedForOneBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	id := handCardForTest(me, "Phthisis", "Sorcery", rprPhthisisOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("the suspended card is not in exile")
	}
	c, _ := g.LookupCardForEffect(id)
	if c.Counters[game.CounterTime] != 5 {
		t.Errorf("time counters = %d, want 5", c.Counters[game.CounterTime])
	}
}

// --- Obsessive Stitcher ------------------------------------------------------

func TestRprB35ObsessiveStitcherLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	msc18Library(me, 6)
	stitcher := pushCatalogPermanent(g, me.ID, "Obsessive Stitcher", "Creature — Human Wizard", rprObsessiveStitcherOracle, false)
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, stitcher, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand = %d after the draw, want %d", me.Hand.Size(), hand+1)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand = %d after the discard, want %d", me.Hand.Size(), hand)
	}
}

func TestRprB35ObsessiveStitcherSacrificesItselfToReanimateFromYourYard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	stitcher := pushCatalogPermanent(g, me.ID, "Obsessive Stitcher", "Creature — Human Wizard", rprObsessiveStitcherOracle, false)
	mine := b38Graveyard(g, me.ID, "Dead Bear", "Creature — Bear", "{1}{G}", 2, 2)
	theirs := b38Graveyard(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	b06AddMana(me, "C", "C", "U", "B")
	if err := g.ActivateCatalogAbility(me.ID, stitcher, 1, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(theirs),
	}); err == nil {
		t.Fatal("an opponent's graveyard creature was a legal target")
	}
	b16Activate(t, g, me.ID, stitcher, 1, game.ActivateAbilityParams{Targets: cardRefs(mine)})
	if g.Battlefield.Contains(stitcher) {
		t.Error("the sacrifice is a cost, paid on activation")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("the creature card was not returned")
	}
}

// --- Piper of the Swarm ------------------------------------------------------

func TestRprB35PiperGivesRatsMenaceAndMakesRats(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	piper := pushCatalogPermanent(g, me.ID, "Piper of the Swarm", "Creature — Human Warlock", rprPiperOfTheSwarmOracle, false)
	myRat := b12Creature(g, me.ID, "Rat", "Creature — Rat", 1, 1)
	theirRat := b12Creature(g, opp.ID, "Rat", "Creature — Rat", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, myRat, "menace") {
		t.Error("your Rat lacks menace")
	}
	if hasEffectiveKeyword(t, g, theirRat, "menace") || hasEffectiveKeyword(t, g, bear, "menace") || hasEffectiveKeyword(t, g, piper, "menace") {
		t.Error("menace leaked to an opponent's Rat or a non-Rat")
	}
	b06AddMana(me, "C", "B")
	b16Activate(t, g, me.ID, piper, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Rat"); got != 2 {
		t.Errorf("%d Rats, want 2 (the one already there and the new token)", got)
	}
}

func TestRprB35PiperSacrificesThreeRatsToStealACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	piper := pushCatalogPermanent(g, me.ID, "Piper of the Swarm", "Creature — Human Warlock", rprPiperOfTheSwarmOracle, false)
	r1 := b12Creature(g, me.ID, "Rat", "Creature — Rat", 1, 1)
	r2 := b12Creature(g, me.ID, "Rat", "Creature — Rat", 1, 1)
	r3 := b12Creature(g, me.ID, "Rat", "Creature — Rat", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	target := b12Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 5, 5)
	b06AddMana(me, "C", "C", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, piper, 1, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{r1, r2, bear}, Targets: cardRefs(target),
	}); err == nil {
		t.Fatal("a non-Rat paid the sacrifice cost")
	}
	b06AddMana(me, "C", "C", "B", "B")
	b16Activate(t, g, me.ID, piper, 1, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{r1, r2, r3}, Targets: cardRefs(target),
	})
	for _, id := range []uuid.UUID{r1, r2, r3} {
		if g.Battlefield.Contains(id) {
			t.Error("a Rat was not sacrificed as a cost")
		}
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, target); !ok || c.Controller != me.ID {
		t.Errorf("the dragon is not under your control: %+v %v", c.Controller, ok)
	}
	// Permanent: it is still yours after the turn ends.
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, target); !ok || c.Controller != me.ID {
		t.Error("control did not last past end of turn")
	}
}

// --- Moonsilver Spear --------------------------------------------------------

func TestRprB35MoonsilverSpearGrantsFirstStrikeAndMakesAnAngelOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	spear := pushCatalogPermanent(g, me.ID, "Moonsilver Spear", "Artifact — Equipment", rprMoonsilverSpearOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "C", "C", "C", "C")
	b16Activate(t, g, me.ID, spear, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Fatal("equipped creature lacks first strike")
	}
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	var angel game.Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Angel" && c.Controller == me.ID {
				angel, found = c, true
			}
		}
	})
	if !found {
		t.Fatal("no Angel token was created")
	}
	if angel.Power != 4 || angel.Toughness != 4 {
		t.Errorf("Angel is %d/%d, want 4/4", angel.Power, angel.Toughness)
	}
}
