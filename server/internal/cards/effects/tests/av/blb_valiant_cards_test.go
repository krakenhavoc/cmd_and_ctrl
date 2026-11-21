package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		blbValiantEmberheartOracle:             "Emberheart Challenger",
		blbValiantFlowerfootOracle:             "Flowerfoot Swordmaster",
		effectstest.ValiantHeartfireHeroOracle: "Heartfire Hero",
		effectstest.ValiantMouseTrapperOracle:  "Mouse Trapper",
		blbValiantNettleGuardOracle:            "Nettle Guard",
		blbValiantSeedglaiveOracle:             "Seedglaive Mentor",
		blbValiantGuardmouseOracle:             "Veteran Guardmouse",
		blbValiantWhiskerquillOracle:           "Whiskerquill Scribe",
		blbValiantWhiskervaleOracle:            "Whiskervale Forerunner",
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

// newBlbValiantTable puts `name` under the active player in their main
// phase.
func newBlbValiantTable(t *testing.T, name, typeLine, oracle string, power, toughness int) (effectstest.ValiantTable, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	tb := effectstest.ValiantTable{G: g, Seat: seat, Me: g.Seats[seat], Them: g.Seats[(seat+1)%len(g.Seats)]}
	tb.Hero = effectstest.B12Push(g, tb.Me.ID, name, typeLine, oracle, power, toughness)
	return tb, tb.Hero
}

// --- Heartfire Hero -----------------------------------------------------

func TestBlbValiantHeartfireHeroDealsItsPowerToEachOpponentWhenItDies(t *testing.T) {
	tb := effectstest.NewValiantTable(t)
	effectstest.ValiantTarget(t, tb.G, tb.Me, tb.Hero) // a +1/+1 counter: power 2
	effectstest.PassPriorityAroundTable(t, tb.G)
	before := map[uuid.UUID]int{}
	for _, p := range tb.G.Seats {
		before[p.ID] = p.Life
	}
	tb.G.WithWriteLock(func() {
		if err := tb.G.DestroyPermanentForEffect(tb.Hero); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, tb.G)
	for _, p := range tb.G.Seats {
		want := before[p.ID] - 2
		if p.ID == tb.Me.ID {
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
		if !effectstest.HasKeywordOf(tb.G, mentor, kw) {
			t.Errorf("Seedglaive Mentor lacks %s", kw)
		}
	}
	effectstest.ValiantTarget(t, tb.G, tb.Me, mentor)
	effectstest.PassPriorityAroundTable(t, tb.G)
	if n := effectstest.PlusOneCounters(tb.G, mentor); n != 1 {
		t.Fatalf("Seedglaive Mentor has %d +1/+1 counters, want 1", n)
	}
}

// --- Nettle Guard -------------------------------------------------------

func TestBlbValiantNettleGuardGetsPlusZeroPlusTwo(t *testing.T) {
	tb, guard := newBlbValiantTable(t, "Nettle Guard", "Creature — Mouse Soldier", blbValiantNettleGuardOracle, 3, 1)
	effectstest.ValiantTarget(t, tb.G, tb.Me, guard)
	effectstest.PassPriorityAroundTable(t, tb.G)
	if p, tg := effectstest.EffectivePower(t, tb.G, guard), effectstest.EffectiveToughness(t, tb.G, guard); p != 3 || tg != 3 {
		t.Fatalf("Nettle Guard is %d/%d after valiant, want 3/3", p, tg)
	}
}

func TestBlbValiantNettleGuardHasTheDisenchantAbility(t *testing.T) {
	spec, ok := effects.Lookup(blbValiantNettleGuardOracle)
	if !ok {
		t.Fatal("Nettle Guard is not registered")
	}
	if len(spec.Activated) != 1 || spec.Activated[0].Label != "{1}, Sacrifice this creature: Destroy target artifact or enchantment." {
		t.Fatalf("Nettle Guard's activated abilities: %+v", spec.Activated)
	}
}

// --- Mouse Trapper ------------------------------------------------------

func TestBlbValiantMouseTrapperHasFlash(t *testing.T) {
	tb, trapper := effectstest.NewMouseTrapperTable(t)
	if !effectstest.HasKeywordOf(tb.G, trapper, "flash") {
		t.Error("Mouse Trapper lacks flash")
	}
}

// --- Veteran Guardmouse -------------------------------------------------

func TestBlbValiantVeteranGuardmouseGetsPowerFirstStrikeAndScries(t *testing.T) {
	tb, mouse := newBlbValiantTable(t, "Veteran Guardmouse", "Creature — Mouse Soldier", blbValiantGuardmouseOracle, 3, 4)
	effectstest.ValiantTarget(t, tb.G, tb.Me, mouse)
	effectstest.PassPriorityAroundTable(t, tb.G)
	if effectstest.ScryChoiceFor(tb.G, tb.Me.ID) == nil {
		t.Fatal("no scry 1 prompt")
	}
	if c := effectstest.ScryChoiceFor(tb.G, tb.Me.ID); len(c.ScryCards) != 1 {
		t.Fatalf("scry looks at %d cards, want 1", len(c.ScryCards))
	}
	effectstest.AnswerScryKeepAll(t, tb.G, tb.Me.ID)
	effectstest.PassPriorityAroundTable(t, tb.G)
	if p, tg := effectstest.EffectivePower(t, tb.G, mouse), effectstest.EffectiveToughness(t, tb.G, mouse); p != 4 || tg != 4 {
		t.Errorf("Veteran Guardmouse is %d/%d, want 4/4", p, tg)
	}
	if !effectstest.HasKeywordOf(tb.G, mouse, "first strike") {
		t.Error("Veteran Guardmouse did not gain first strike")
	}
}

// --- Flowerfoot Swordmaster ---------------------------------------------

func TestBlbValiantFlowerfootSwordmasterPumpsOnlyYourMice(t *testing.T) {
	tb, sword := newBlbValiantTable(t, "Flowerfoot Swordmaster", "Creature — Mouse Soldier", blbValiantFlowerfootOracle, 1, 2)
	myMouse := effectstest.B12Creature(tb.G, tb.Me.ID, "My Mouse", "Creature — Mouse Warrior", 2, 2)
	myBear := effectstest.B12Creature(tb.G, tb.Me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirMouse := effectstest.B12Creature(tb.G, tb.Them.ID, "Their Mouse", "Creature — Mouse", 2, 2)
	effectstest.ValiantTarget(t, tb.G, tb.Me, sword)
	effectstest.PassPriorityAroundTable(t, tb.G)
	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{
		"the Swordmaster": {sword, 2},
		"another Mouse":   {myMouse, 3},
		"a non-Mouse":     {myBear, 2},
		"their Mouse":     {theirMouse, 2},
	} {
		if p := effectstest.EffectivePower(t, tb.G, c.id); p != c.want {
			t.Errorf("%s has power %d, want %d", name, p, c.want)
		}
	}
	// A Mouse that arrives after the pump resolved is not pumped (CR 611.2c).
	late := effectstest.B12Creature(tb.G, tb.Me.ID, "Late Mouse", "Creature — Mouse", 1, 1)
	if p := effectstest.EffectivePower(t, tb.G, late); p != 1 {
		t.Errorf("a Mouse that entered later has power %d, want 1", p)
	}
}

func TestBlbValiantFlowerfootSwordmasterHasOffspring(t *testing.T) {
	spec, ok := effects.Lookup(blbValiantFlowerfootOracle)
	if !ok {
		t.Fatal("Flowerfoot Swordmaster is not registered")
	}
	if len(spec.OptionalCosts) != 1 || spec.OptionalCosts[0].Key != effects.OffspringKey || spec.OptionalCosts[0].ManaCost != "{2}" {
		t.Fatalf("optional costs %+v, want Offspring {2}", spec.OptionalCosts)
	}
}

// --- Whiskerquill Scribe ------------------------------------------------

func TestBlbValiantWhiskerquillScribeRummages(t *testing.T) {
	tb, scribe := newBlbValiantTable(t, "Whiskerquill Scribe", "Creature — Mouse Citizen", blbValiantWhiskerquillOracle, 2, 2)
	fodder := uuid.New()
	tb.Me.Hand.PushTop(game.Card{InstanceID: fodder, Name: "Fodder", TypeLine: "Sorcery", Owner: tb.Me.ID, Controller: tb.Me.ID})
	drawn := effectstest.SeedLibraryTop(tb.Me, "Drawn", "Instant", "{U}")
	effectstest.ValiantTarget(t, tb.G, tb.Me, scribe)
	effectstest.PassPriorityAroundTable(t, tb.G)
	effectstest.AnswerDiscard(t, tb.G, tb.Me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, tb.G)
	if !tb.Me.Graveyard.Contains(fodder) {
		t.Error("the discarded card is not in the graveyard")
	}
	if !tb.Me.Hand.Contains(drawn) {
		t.Error("no card was drawn for the discard")
	}
}

func TestBlbValiantWhiskerquillScribeDrawsNothingWithoutADiscard(t *testing.T) {
	tb, scribe := newBlbValiantTable(t, "Whiskerquill Scribe", "Creature — Mouse Citizen", blbValiantWhiskerquillOracle, 2, 2)
	keep := uuid.New()
	tb.Me.Hand.PushTop(game.Card{InstanceID: keep, Name: "Keep", TypeLine: "Sorcery", Owner: tb.Me.ID, Controller: tb.Me.ID})
	drawn := effectstest.SeedLibraryTop(tb.Me, "Not Drawn", "Instant", "{U}")
	effectstest.ValiantTarget(t, tb.G, tb.Me, scribe)
	effectstest.PassPriorityAroundTable(t, tb.G)
	effectstest.AnswerDiscard(t, tb.G, tb.Me.ID)
	effectstest.PassPriorityAroundTable(t, tb.G)
	if !tb.Me.Hand.Contains(keep) || tb.Me.Hand.Contains(drawn) {
		t.Error("declining the discard still drew a card or lost the kept one")
	}
}

// --- Emberheart Challenger ----------------------------------------------

func TestBlbValiantEmberheartChallengerImpulsesTheTopCard(t *testing.T) {
	tb, challenger := newBlbValiantTable(t, "Emberheart Challenger", "Creature — Mouse Warrior", blbValiantEmberheartOracle, 2, 2)
	for _, kw := range []string{"haste", "prowess"} {
		if !effectstest.HasKeywordOf(tb.G, challenger, kw) {
			t.Errorf("Emberheart Challenger lacks %s", kw)
		}
	}
	top := effectstest.SeedLibrary(tb.Me, "Top Card")[0]
	effectstest.ValiantTarget(t, tb.G, tb.Me, challenger)
	// The instant is a noncreature spell, so prowess triggers beside valiant
	// and the two are ordered (CR 603.3b).
	effectstest.PassPriorityAroundTable(t, tb.G)
	if effectstest.TriggerOrderPrompt(tb.G) != nil {
		effectstest.AnswerTriggerOrderInOfferedOrder(t, tb.G)
	}
	effectstest.PassPriorityAroundTable(t, tb.G)
	if p := effectstest.EffectivePower(t, tb.G, challenger); p != 3 {
		t.Errorf("Emberheart Challenger has power %d, want 3 from prowess", p)
	}
	if !tb.G.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	perm := effectstest.ExiledPermission(tb.G, top)
	if perm.Player != tb.Me.ID || perm.CastOnly {
		t.Errorf("permission %+v, want a play grant for the controller", perm)
	}
	if !effectstest.PermissionLive(tb.G, perm, tb.Me.ID) {
		t.Error("the exiled card is not playable this turn")
	}
}

// --- Whiskervale Forerunner ---------------------------------------------

// blbValiantSeedForerunnerLibrary puts a mana value 2 creature, a mana
// value 5 creature and three sorceries on top of `p`'s library.
func blbValiantSeedForerunnerLibrary(p *game.Player) (small, big uuid.UUID) {
	effectstest.SeedLibrary(p, "One", "Two", "Three")
	small, big = uuid.New(), uuid.New()
	p.Library.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{5}", Power: 5, Toughness: 5, Owner: p.ID, Controller: p.ID})
	p.Library.PushTop(game.Card{InstanceID: small, Name: "Small", TypeLine: "Creature — Elf", ManaCost: "{2}", Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	return small, big
}

func TestBlbValiantWhiskervaleForerunnerPutsTheCreatureOntoTheBattlefieldOnYourTurn(t *testing.T) {
	tb, forerunner := newBlbValiantTable(t, "Whiskervale Forerunner", "Creature — Mouse Bard", blbValiantWhiskervaleOracle, 3, 4)
	small, big := blbValiantSeedForerunnerLibrary(tb.Me)
	effectstest.ValiantTarget(t, tb.G, tb.Me, forerunner)
	effectstest.PassPriorityAroundTable(t, tb.G)
	pick := effectstest.LatestChooseCards(tb.G, tb.Me.ID)
	if pick == nil {
		t.Fatal("no battlefield prompt")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != small {
		t.Fatalf("offered %v, want only the mana value 2 creature", pick.ChooseCards)
	}
	if err := tb.G.ResolveChooseCards(pick.ID, tb.Me.ID, []uuid.UUID{small}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, tb.G)
	if !effectstest.OnBoard(tb.G, small) {
		t.Fatal("the creature did not enter the battlefield")
	}
	if effectstest.LatestChooseCards(tb.G, tb.Me.ID) != nil {
		t.Error("a hand prompt followed a creature put onto the battlefield")
	}
	if !tb.Me.Library.Contains(big) {
		t.Error("an unpicked card left the library")
	}
	if top := tb.Me.Library.Cards[len(tb.Me.Library.Cards)-1].InstanceID; top == big {
		t.Error("the rest stayed on top instead of going to the bottom")
	}
}

func TestBlbValiantWhiskervaleForerunnerOffersTheHandWhenYouDeclineTheBattlefield(t *testing.T) {
	tb, forerunner := newBlbValiantTable(t, "Whiskervale Forerunner", "Creature — Mouse Bard", blbValiantWhiskervaleOracle, 3, 4)
	small, _ := blbValiantSeedForerunnerLibrary(tb.Me)
	effectstest.ValiantTarget(t, tb.G, tb.Me, forerunner)
	effectstest.PassPriorityAroundTable(t, tb.G)
	pick := effectstest.LatestChooseCards(tb.G, tb.Me.ID)
	if pick == nil {
		t.Fatal("no battlefield prompt")
	}
	if err := tb.G.ResolveChooseCards(pick.ID, tb.Me.ID, nil); err != nil {
		t.Fatal(err)
	}
	hand := effectstest.LatestChooseCards(tb.G, tb.Me.ID)
	if hand == nil {
		t.Fatal("no hand prompt after declining the battlefield")
	}
	if err := tb.G.ResolveChooseCards(hand.ID, tb.Me.ID, []uuid.UUID{small}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, tb.G)
	if !tb.Me.Hand.Contains(small) || effectstest.OnBoard(tb.G, small) {
		t.Error("the creature did not go to the hand")
	}
}

func TestBlbValiantWhiskervaleForerunnerOnlyOffersTheHandOnAnotherPlayersTurn(t *testing.T) {
	tb, forerunner := newBlbValiantTable(t, "Whiskervale Forerunner", "Creature — Mouse Bard", blbValiantWhiskervaleOracle, 3, 4)
	effectstest.ToNextTurnMainOf(t, tb.G, (tb.Seat+1)%len(tb.G.Seats))
	small, _ := blbValiantSeedForerunnerLibrary(tb.Me)
	effectstest.ValiantTarget(t, tb.G, tb.Me, forerunner)
	effectstest.PassPriorityAroundTable(t, tb.G)
	pick := effectstest.LatestChooseCards(tb.G, tb.Me.ID)
	if pick == nil {
		t.Fatal("no prompt")
	}
	if err := tb.G.ResolveChooseCards(pick.ID, tb.Me.ID, []uuid.UUID{small}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, tb.G)
	if effectstest.OnBoard(tb.G, small) {
		t.Error("the creature entered the battlefield on another player's turn")
	}
	if !tb.Me.Hand.Contains(small) {
		t.Error("the creature did not go to the hand")
	}
	if effectstest.LatestChooseCards(tb.G, tb.Me.ID) != nil {
		t.Error("a second prompt followed the hand pick")
	}
}
