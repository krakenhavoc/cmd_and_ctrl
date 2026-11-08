package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_valiant_cards_test.go — the Bloomburrow valiant cards (#5230). The
// trigger's own rules (first time, controller, two slots, a new object) are
// in valiant_test.go; these drive each card's effect half.

const (
	blbValiantEmberheartOracle   = "5b4b21ed-c24e-4038-83fb-bd7f3c3426cd"
	blbValiantFlowerfootOracle   = "523d075d-b577-4f24-8c14-9a76761a2eaf"
	blbValiantNettleGuardOracle  = "9ac2dcb6-dacc-4fcb-b820-9fc6744ec503"
	blbValiantSeedglaiveOracle   = "f50bbe5e-a839-40e1-ba79-1e96f91bd218"
	blbValiantGuardmouseOracle   = "2313a3f9-3eac-4684-a25e-6a6ae445a723"
	blbValiantWhiskerquillOracle = "cb1e1acc-3805-418b-bed1-65838ddb1d62"
	blbValiantWhiskervaleOracle  = "861e427f-cbb4-4ed5-8356-747ac638ac34"
)

func TestBlbValiantEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbValiantEmberheartOracle:   "Emberheart Challenger",
		blbValiantFlowerfootOracle:   "Flowerfoot Swordmaster",
		valiantHeartfireHeroOracle:   "Heartfire Hero",
		valiantMouseTrapperOracle:    "Mouse Trapper",
		blbValiantNettleGuardOracle:  "Nettle Guard",
		blbValiantSeedglaiveOracle:   "Seedglaive Mentor",
		blbValiantGuardmouseOracle:   "Veteran Guardmouse",
		blbValiantWhiskerquillOracle: "Whiskerquill Scribe",
		blbValiantWhiskervaleOracle:  "Whiskervale Forerunner",
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

// newBlbValiantTable puts `name` under the active player in their main
// phase.
func newBlbValiantTable(t *testing.T, name, typeLine, oracle string, power, toughness int) (valiantTable, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	tb := valiantTable{g: g, seat: seat, me: g.Seats[seat], them: g.Seats[(seat+1)%len(g.Seats)]}
	tb.hero = b12Push(g, tb.me.ID, name, typeLine, oracle, power, toughness)
	return tb, tb.hero
}

// --- Heartfire Hero -----------------------------------------------------

func TestBlbValiantHeartfireHeroDealsItsPowerToEachOpponentWhenItDies(t *testing.T) {
	tb := newValiantTable(t)
	valiantTarget(t, tb.g, tb.me, tb.hero) // a +1/+1 counter: power 2
	passPriorityAroundTable(t, tb.g)
	before := map[uuid.UUID]int{}
	for _, p := range tb.g.Seats {
		before[p.ID] = p.Life
	}
	tb.g.WithWriteLock(func() {
		if err := tb.g.DestroyPermanentForEffect(tb.hero); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, tb.g)
	for _, p := range tb.g.Seats {
		want := before[p.ID] - 2
		if p.ID == tb.me.ID {
			want = before[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %s life %d, want %d", p.Name, p.Life, want)
		}
	}
}

// --- Seedglaive Mentor --------------------------------------------------

func TestBlbValiantSeedglaiveMentorGetsACounterAndHasItsKeywords(t *testing.T) {
	tb, mentor := newBlbValiantTable(t, "Seedglaive Mentor", "Creature — Mouse Soldier", blbValiantSeedglaiveOracle, 3, 2)
	for _, kw := range []string{"vigilance", "haste"} {
		if !hasKeywordOf(tb.g, mentor, kw) {
			t.Errorf("Seedglaive Mentor lacks %s", kw)
		}
	}
	valiantTarget(t, tb.g, tb.me, mentor)
	passPriorityAroundTable(t, tb.g)
	if n := plusOneCounters(tb.g, mentor); n != 1 {
		t.Fatalf("Seedglaive Mentor has %d +1/+1 counters, want 1", n)
	}
}

// --- Nettle Guard -------------------------------------------------------

func TestBlbValiantNettleGuardGetsPlusZeroPlusTwo(t *testing.T) {
	tb, guard := newBlbValiantTable(t, "Nettle Guard", "Creature — Mouse Soldier", blbValiantNettleGuardOracle, 3, 1)
	valiantTarget(t, tb.g, tb.me, guard)
	passPriorityAroundTable(t, tb.g)
	if p, tg := effectivePower(t, tb.g, guard), effectiveToughness(t, tb.g, guard); p != 3 || tg != 3 {
		t.Fatalf("Nettle Guard is %d/%d after valiant, want 3/3", p, tg)
	}
}

func TestBlbValiantNettleGuardHasTheDisenchantAbility(t *testing.T) {
	spec, ok := Lookup(blbValiantNettleGuardOracle)
	if !ok {
		t.Fatal("Nettle Guard is not registered")
	}
	if len(spec.Activated) != 1 || spec.Activated[0].Label != "{1}, Sacrifice this creature: Destroy target artifact or enchantment." {
		t.Fatalf("Nettle Guard's activated abilities: %+v", spec.Activated)
	}
}

// --- Mouse Trapper ------------------------------------------------------

func TestBlbValiantMouseTrapperHasFlash(t *testing.T) {
	tb, trapper := newMouseTrapperTable(t)
	if !hasKeywordOf(tb.g, trapper, "flash") {
		t.Error("Mouse Trapper lacks flash")
	}
}

// --- Veteran Guardmouse -------------------------------------------------

func TestBlbValiantVeteranGuardmouseGetsPowerFirstStrikeAndScries(t *testing.T) {
	tb, mouse := newBlbValiantTable(t, "Veteran Guardmouse", "Creature — Mouse Soldier", blbValiantGuardmouseOracle, 3, 4)
	valiantTarget(t, tb.g, tb.me, mouse)
	passPriorityAroundTable(t, tb.g)
	if scryChoiceFor(tb.g, tb.me.ID) == nil {
		t.Fatal("no scry 1 prompt")
	}
	if c := scryChoiceFor(tb.g, tb.me.ID); len(c.ScryCards) != 1 {
		t.Fatalf("scry looks at %d cards, want 1", len(c.ScryCards))
	}
	answerScryKeepAll(t, tb.g, tb.me.ID)
	passPriorityAroundTable(t, tb.g)
	if p, tg := effectivePower(t, tb.g, mouse), effectiveToughness(t, tb.g, mouse); p != 4 || tg != 4 {
		t.Errorf("Veteran Guardmouse is %d/%d, want 4/4", p, tg)
	}
	if !hasKeywordOf(tb.g, mouse, "first strike") {
		t.Error("Veteran Guardmouse did not gain first strike")
	}
}

// --- Flowerfoot Swordmaster ---------------------------------------------

func TestBlbValiantFlowerfootSwordmasterPumpsOnlyYourMice(t *testing.T) {
	tb, sword := newBlbValiantTable(t, "Flowerfoot Swordmaster", "Creature — Mouse Soldier", blbValiantFlowerfootOracle, 1, 2)
	myMouse := b12Creature(tb.g, tb.me.ID, "My Mouse", "Creature — Mouse Warrior", 2, 2)
	myBear := b12Creature(tb.g, tb.me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirMouse := b12Creature(tb.g, tb.them.ID, "Their Mouse", "Creature — Mouse", 2, 2)
	valiantTarget(t, tb.g, tb.me, sword)
	passPriorityAroundTable(t, tb.g)
	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{
		"the Swordmaster": {sword, 2},
		"another Mouse":   {myMouse, 3},
		"a non-Mouse":     {myBear, 2},
		"their Mouse":     {theirMouse, 2},
	} {
		if p := effectivePower(t, tb.g, c.id); p != c.want {
			t.Errorf("%s has power %d, want %d", name, p, c.want)
		}
	}
	// A Mouse that arrives after the pump resolved is not pumped (CR 611.2c).
	late := b12Creature(tb.g, tb.me.ID, "Late Mouse", "Creature — Mouse", 1, 1)
	if p := effectivePower(t, tb.g, late); p != 1 {
		t.Errorf("a Mouse that entered later has power %d, want 1", p)
	}
}

func TestBlbValiantFlowerfootSwordmasterHasOffspring(t *testing.T) {
	spec, ok := Lookup(blbValiantFlowerfootOracle)
	if !ok {
		t.Fatal("Flowerfoot Swordmaster is not registered")
	}
	if len(spec.OptionalCosts) != 1 || spec.OptionalCosts[0].Key != OffspringKey || spec.OptionalCosts[0].ManaCost != "{2}" {
		t.Fatalf("optional costs %+v, want Offspring {2}", spec.OptionalCosts)
	}
}

// --- Whiskerquill Scribe ------------------------------------------------

func TestBlbValiantWhiskerquillScribeRummages(t *testing.T) {
	tb, scribe := newBlbValiantTable(t, "Whiskerquill Scribe", "Creature — Mouse Citizen", blbValiantWhiskerquillOracle, 2, 2)
	fodder := uuid.New()
	tb.me.Hand.PushTop(game.Card{InstanceID: fodder, Name: "Fodder", TypeLine: "Sorcery", Owner: tb.me.ID, Controller: tb.me.ID})
	drawn := seedLibraryTop(tb.me, "Drawn", "Instant", "{U}")
	valiantTarget(t, tb.g, tb.me, scribe)
	passPriorityAroundTable(t, tb.g)
	answerDiscard(t, tb.g, tb.me.ID, fodder)
	passPriorityAroundTable(t, tb.g)
	if !tb.me.Graveyard.Contains(fodder) {
		t.Error("the discarded card is not in the graveyard")
	}
	if !tb.me.Hand.Contains(drawn) {
		t.Error("no card was drawn for the discard")
	}
}

func TestBlbValiantWhiskerquillScribeDrawsNothingWithoutADiscard(t *testing.T) {
	tb, scribe := newBlbValiantTable(t, "Whiskerquill Scribe", "Creature — Mouse Citizen", blbValiantWhiskerquillOracle, 2, 2)
	keep := uuid.New()
	tb.me.Hand.PushTop(game.Card{InstanceID: keep, Name: "Keep", TypeLine: "Sorcery", Owner: tb.me.ID, Controller: tb.me.ID})
	drawn := seedLibraryTop(tb.me, "Not Drawn", "Instant", "{U}")
	valiantTarget(t, tb.g, tb.me, scribe)
	passPriorityAroundTable(t, tb.g)
	answerDiscard(t, tb.g, tb.me.ID)
	passPriorityAroundTable(t, tb.g)
	if !tb.me.Hand.Contains(keep) || tb.me.Hand.Contains(drawn) {
		t.Error("declining the discard still drew a card or lost the kept one")
	}
}

// --- Emberheart Challenger ----------------------------------------------

func TestBlbValiantEmberheartChallengerImpulsesTheTopCard(t *testing.T) {
	tb, challenger := newBlbValiantTable(t, "Emberheart Challenger", "Creature — Mouse Warrior", blbValiantEmberheartOracle, 2, 2)
	for _, kw := range []string{"haste", "prowess"} {
		if !hasKeywordOf(tb.g, challenger, kw) {
			t.Errorf("Emberheart Challenger lacks %s", kw)
		}
	}
	top := seedLibrary(tb.me, "Top Card")[0]
	valiantTarget(t, tb.g, tb.me, challenger)
	// The instant is a noncreature spell, so prowess triggers beside valiant
	// and the two are ordered (CR 603.3b).
	passPriorityAroundTable(t, tb.g)
	if triggerOrderPrompt(tb.g) != nil {
		answerTriggerOrderInOfferedOrder(t, tb.g)
	}
	passPriorityAroundTable(t, tb.g)
	if p := effectivePower(t, tb.g, challenger); p != 3 {
		t.Errorf("Emberheart Challenger has power %d, want 3 from prowess", p)
	}
	if !tb.g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	perm := exiledPermission(tb.g, top)
	if perm.Player != tb.me.ID || perm.CastOnly {
		t.Errorf("permission %+v, want a play grant for the controller", perm)
	}
	if !permissionLive(tb.g, perm, tb.me.ID) {
		t.Error("the exiled card is not playable this turn")
	}
}

// --- Whiskervale Forerunner ---------------------------------------------

// blbValiantSeedForerunnerLibrary puts a mana value 2 creature, a mana
// value 5 creature and three sorceries on top of `p`'s library.
func blbValiantSeedForerunnerLibrary(p *game.Player) (small, big uuid.UUID) {
	seedLibrary(p, "One", "Two", "Three")
	small, big = uuid.New(), uuid.New()
	p.Library.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{5}", Power: 5, Toughness: 5, Owner: p.ID, Controller: p.ID})
	p.Library.PushTop(game.Card{InstanceID: small, Name: "Small", TypeLine: "Creature — Elf", ManaCost: "{2}", Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	return small, big
}

func TestBlbValiantWhiskervaleForerunnerPutsTheCreatureOntoTheBattlefieldOnYourTurn(t *testing.T) {
	tb, forerunner := newBlbValiantTable(t, "Whiskervale Forerunner", "Creature — Mouse Bard", blbValiantWhiskervaleOracle, 3, 4)
	small, big := blbValiantSeedForerunnerLibrary(tb.me)
	valiantTarget(t, tb.g, tb.me, forerunner)
	passPriorityAroundTable(t, tb.g)
	pick := latestChooseCards(tb.g, tb.me.ID)
	if pick == nil {
		t.Fatal("no battlefield prompt")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != small {
		t.Fatalf("offered %v, want only the mana value 2 creature", pick.ChooseCards)
	}
	if err := tb.g.ResolveChooseCards(pick.ID, tb.me.ID, []uuid.UUID{small}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, tb.g)
	if !onBoard(tb.g, small) {
		t.Fatal("the creature did not enter the battlefield")
	}
	if latestChooseCards(tb.g, tb.me.ID) != nil {
		t.Error("a hand prompt followed a creature put onto the battlefield")
	}
	if !tb.me.Library.Contains(big) {
		t.Error("an unpicked card left the library")
	}
	if top := tb.me.Library.Cards[len(tb.me.Library.Cards)-1].InstanceID; top == big {
		t.Error("the rest stayed on top instead of going to the bottom")
	}
}

func TestBlbValiantWhiskervaleForerunnerOffersTheHandWhenYouDeclineTheBattlefield(t *testing.T) {
	tb, forerunner := newBlbValiantTable(t, "Whiskervale Forerunner", "Creature — Mouse Bard", blbValiantWhiskervaleOracle, 3, 4)
	small, _ := blbValiantSeedForerunnerLibrary(tb.me)
	valiantTarget(t, tb.g, tb.me, forerunner)
	passPriorityAroundTable(t, tb.g)
	pick := latestChooseCards(tb.g, tb.me.ID)
	if pick == nil {
		t.Fatal("no battlefield prompt")
	}
	if err := tb.g.ResolveChooseCards(pick.ID, tb.me.ID, nil); err != nil {
		t.Fatal(err)
	}
	hand := latestChooseCards(tb.g, tb.me.ID)
	if hand == nil {
		t.Fatal("no hand prompt after declining the battlefield")
	}
	if err := tb.g.ResolveChooseCards(hand.ID, tb.me.ID, []uuid.UUID{small}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, tb.g)
	if !tb.me.Hand.Contains(small) || onBoard(tb.g, small) {
		t.Error("the creature did not go to the hand")
	}
}

func TestBlbValiantWhiskervaleForerunnerOnlyOffersTheHandOnAnotherPlayersTurn(t *testing.T) {
	tb, forerunner := newBlbValiantTable(t, "Whiskervale Forerunner", "Creature — Mouse Bard", blbValiantWhiskervaleOracle, 3, 4)
	toNextTurnMainOf(t, tb.g, (tb.seat+1)%len(tb.g.Seats))
	small, _ := blbValiantSeedForerunnerLibrary(tb.me)
	valiantTarget(t, tb.g, tb.me, forerunner)
	passPriorityAroundTable(t, tb.g)
	pick := latestChooseCards(tb.g, tb.me.ID)
	if pick == nil {
		t.Fatal("no prompt")
	}
	if err := tb.g.ResolveChooseCards(pick.ID, tb.me.ID, []uuid.UUID{small}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, tb.g)
	if onBoard(tb.g, small) {
		t.Error("the creature entered the battlefield on another player's turn")
	}
	if !tb.me.Hand.Contains(small) {
		t.Error("the creature did not go to the hand")
	}
	if latestChooseCards(tb.g, tb.me.ID) != nil {
		t.Error("a second prompt followed the hand pick")
	}
}
