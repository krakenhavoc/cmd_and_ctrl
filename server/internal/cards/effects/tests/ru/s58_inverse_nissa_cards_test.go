package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestS58InverseNissaRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.S58nStockUp:        "Stock Up",
		effectstest.S58nMemoryDeluge:   "Memory Deluge",
		effectstest.S58nFatefulAbsence: "Fateful Absence",
		effectstest.S58nJeskaiRev:      "Jeskai Revelation",
		effectstest.S58nWhiteAuracite:  "White Auracite",
		effectstest.S58nTablet:         "Tablet of Discovery",
		effectstest.S58nWeatherMaker:   "Weather Maker",
		effectstest.S58nAvacyn:         "Avacyn, Angel of Horror",
		effectstest.S58nNissa:          "Nissa, Leyline Tamer",
		effectstest.S58nBahamut:        "Summon: Bahamut",
		effectstest.S58nSupremeVerdict: "Supreme Verdict",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s: registered=%v name=%q", name, ok, spec.Name)
			continue
		}
		// Memory Deluge too since #3855: a waived payment asks.
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want %v", name, spec.Completeness, effects.CompletenessFull)
		}
	}
}

// --- Stock Up ---------------------------------------------------------

func TestS58StockUpTakesTwoOfFiveAndOrdersTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "Sixth")
	hand := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Stock Up", "Sorcery", effectstest.S58nStockUp, nil)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("Stock Up raised no take prompt")
	}
	if pick.ChooseMin != 2 || pick.ChooseMax != 2 || len(pick.ChooseCards) != 5 {
		t.Fatalf("take exactly two of five: %d..%d over %d", pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1], ids[3]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[3]) || me.Hand.Size() != hand+2 {
		t.Fatal("the two picked cards go to hand")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 3 {
		t.Fatalf("the other three are ordered onto the bottom: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[4], ids[0], ids[2]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if top, _ := me.Library.Top(); top.InstanceID != ids[5] {
		t.Error("the sixth card is now on top")
	}
}

// --- Memory Deluge ----------------------------------------------------

// X is the mana spent: four from hand, seven for flashback.
func TestS58MemoryDelugeLooksAtTheManaSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "1", "2", "3", "4", "5", "6", "7", "8", "9", "10")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UUUU")
	deluge := effectstest.CastFromHandForTest(t, g, me, "Memory Deluge", "Instant", "{2}{U}{U}", effectstest.S58nMemoryDeluge, effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 4 || pick.ChooseMax != 2 {
		t.Fatalf("four mana spent: take two of four, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if effectstest.PutInLibraryChoiceFor(g, me.ID) != nil {
		t.Error("the rest go to the bottom in a random order, with no ordering prompt")
	}
	if !me.Graveyard.Contains(deluge) {
		t.Fatal("Memory Deluge is in the graveyard after resolving")
	}

	effectstest.FloatForTest(g, me, "UUUUUUU")
	if err := g.CastSpell(me.ID, deluge, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Strict: true,
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick = effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 7 {
		t.Fatalf("seven mana spent on flashback: look at seven, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Exile.Contains(deluge) {
		t.Error("a flashed-back spell is exiled")
	}
}

// --- Fateful Absence --------------------------------------------------

func TestS58FatefulAbsenceDestroysAndTheVictimInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	clues := func(owner uuid.UUID) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Clue" && c.Controller == owner {
				n++
			}
		}
		return n
	}
	effectstest.CastCatalogSpell(t, g, "Fateful Absence", "Instant", effectstest.S58nFatefulAbsence,
		[]game.TargetRef{{Kind: game.TargetCard, ID: walker}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Fatal("the creature is destroyed")
	}
	if clues(opp.ID) != 1 || clues(me.ID) != 0 {
		t.Errorf("its controller investigates: theirs %d, mine %d", clues(opp.ID), clues(me.ID))
	}
}

// --- Jeskai Revelation ------------------------------------------------

func TestS58JeskaiRevelationDoesAllFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life, hand, oppLife := me.Life, me.Hand.Size(), opp.Life

	effectstest.CastCatalogSpell(t, g, "Jeskai Revelation", "Instant", effectstest.S58nJeskaiRev, []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs, Slot: 0}, {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Hand.Contains(theirs) {
		t.Error("the permanent returns to its owner's hand")
	}
	if opp.Life != oppLife-4 {
		t.Errorf("4 damage to the player: %d → %d", oppLife, opp.Life)
	}
	monks := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Monk" && c.Controller == me.ID {
			monks++
			if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "prowess") {
				t.Error("the Monks have prowess")
			}
		}
	}
	if monks != 2 {
		t.Errorf("two Monk tokens, got %d", monks)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("draw two: hand %d → %d", hand, me.Hand.Size())
	}
	if me.Life != life+4 {
		t.Errorf("gain 4: %d → %d", life, me.Life)
	}
}

// A bounce target that left in response doesn't stop the rest.
func TestS58JeskaiRevelationStillResolvesWithoutItsBounceTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 5)
	hand := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Jeskai Revelation", "Instant", effectstest.S58nJeskaiRev, []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs, Slot: 0}, {Kind: game.TargetCard, ID: other, Slot: 1},
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.Pr6Marked(g, other) != 4 {
		t.Errorf("the damage still happens: %d marked, want 4", effectstest.Pr6Marked(g, other))
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("the draw still happens: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- White Auracite ---------------------------------------------------

func TestS58WhiteAuraciteExilesUntilItLeavesAndTapsForWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	threat := effectstest.B12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	rock := effectstest.CastCatalogSpell(t, g, "White Auracite", "Artifact", effectstest.S58nWhiteAuracite, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("only a permanent an opponent controls can be chosen")
	}
	effectstest.PickCard(t, g, me.ID, threat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(threat) {
		t.Fatal("the opposing creature is exiled")
	}

	effectstest.ActivateManaFor(t, g, me.ID, rock, 0, game.ManaAbilityParams{})
	if got := effectstest.GmrPoolCount(me, "W"); got != 1 {
		t.Errorf("{T}: Add {W} — pool has %d {W}", got)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	g.RunStateChecksForTest()
	if effectstest.FindBattlefieldByName(g, "Their Threat") == uuid.Nil {
		t.Error("the exiled card returns when White Auracite leaves")
	}
}

// --- Tablet of Discovery ----------------------------------------------

// The milled card can be played this turn, and a land is played, not
// cast.
func TestS58TabletOfDiscoveryMillsACardYouMayPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.SeedLibraryTop(me, "Mountain", "Basic Land — Mountain", "")

	effectstest.CastCatalogSpell(t, g, "Tablet of Discovery", "Artifact", effectstest.S58nTablet, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(land) {
		t.Fatal("the entry trigger mills the top card")
	}
	if g.CastPermissionOnCardByIDForEffect(land) == nil {
		t.Fatal("the milled card carries a play permission")
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("playing the milled land: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the milled land is on the battlefield")
	}
}

func TestS58TabletOfDiscoveryRestrictedManaIsForInstantsAndSorceries(t *testing.T) {
	spec, _ := effects.Lookup(effectstest.S58nTablet)
	if len(spec.ManaAbilities) != 2 {
		t.Fatalf("two mana abilities, got %d", len(spec.ManaAbilities))
	}
	if spec.ManaAbilities[0].Restrictions != nil {
		t.Error("{T}: Add {R} is unrestricted")
	}
	got := spec.ManaAbilities[1].Restrictions
	want := []string{effects.ManaRestrictCast, effects.ManaRestrictAnyType("Instant", "Sorcery")}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("{R}{R} restrictions = %v, want %v", got, want)
	}
}

// --- Weather Maker ----------------------------------------------------

func TestS58WeatherMakerBanksLandfallAndSpendsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	maker := effectstest.B27Push(g, me.ID, "Weather Maker", "Artifact", effectstest.S58nWeatherMaker, "{3}", 0, 0)
	for i := 0; i < 5; i++ {
		effectstest.PlayLandFromHand(t, g, "Forest", "")
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CounterCount(g, maker, game.CounterCharge); got != 5 {
		t.Fatalf("five landfalls, %d charge counters", got)
	}

	effectstest.ActivateManaFor(t, g, me.ID, maker, 1, game.ManaAbilityParams{})
	if got := effectstest.CounterCount(g, maker, game.CounterCharge); got != 3 {
		t.Errorf("{C}{C} removes two: %d left", got)
	}
	if got := effectstest.GmrPoolCount(me, "C"); got != 2 {
		t.Errorf("pool has %d {C}, want 2", got)
	}

	effectstest.UntapForTest(g, maker)
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, maker, effectstest.S58nWeatherMakerDmg, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("3 damage: %d → %d", life, opp.Life)
	}
	if got := effectstest.CounterCount(g, maker, game.CounterCharge); got != 0 {
		t.Errorf("the damage removes three: %d left", got)
	}
}

// --- Avacyn, Angel of Horror ------------------------------------------

func TestS58AvacynReturnsYourNontokenCreaturesAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	avacyn := effectstest.B12Push(g, me.ID, "Avacyn, Angel of Horror", "Legendary Creature — Angel", effectstest.S58nAvacyn, 6, 6)
	if !effectstest.HasEffectiveKeyword(t, g, avacyn, "flying") || !effectstest.HasEffectiveKeyword(t, g, avacyn, "deathtouch") {
		t.Error("flying, deathtouch")
	}
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")
	effectstest.AdvanceToMain(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if !effectstest.B19TriggerQueuedFrom(g, avacyn) {
		t.Fatal("another nontoken creature you control died")
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(goblin) })
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	if effectstest.B19TriggerQueuedFrom(g, avacyn) {
		t.Error("a token, and an opponent's creature, don't trigger it")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bear) {
		t.Fatal("the card waits in the graveyard until the end step")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the card returns to the battlefield at the beginning of the end step")
	}
	if c := effectstest.FindBattlefieldCardForTest(g, bear); c.Controller != me.ID {
		t.Error("under your control")
	}
}

// Avacyn's own death triggers too, and the card that left the graveyard
// before the end step stays where it went.
func TestS58AvacynReturnsHerselfButNotACardThatLeftTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	avacyn := effectstest.B12Push(g, me.ID, "Avacyn, Angel of Horror", "Legendary Creature — Angel", effectstest.S58nAvacyn, 6, 6)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(avacyn) })
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardUnderControlForEffect(bear, game.ZoneHand, uuid.Nil)
	})
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(avacyn) {
		t.Error("Avacyn's own death returns her")
	}
	if !me.Hand.Contains(bear) {
		t.Error("a card that left the graveyard is a new object and stays in hand")
	}
}

// --- Nissa, Leyline Tamer ---------------------------------------------

func TestS58NissaDrawsEachLandfallAndRevealsOnlyTheFirst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nissa := effectstest.B12Push(g, me.ID, "Nissa, Leyline Tamer", "Legendary Creature — Elf Wizard", effectstest.S58nNissa, 4, 5)
	if !effectstest.HasEffectiveKeyword(t, g, nissa, "deathtouch") || !effectstest.HasEffectiveKeyword(t, g, nissa, "vigilance") {
		t.Error("deathtouch, vigilance")
	}
	// Top to bottom: a card to draw, a sorcery, a creature, and more.
	effectstest.SeedLibraryTop(me, "Filler A", "Sorcery", "")
	beast := effectstest.SeedLibraryTop(me, "Beast", "Creature — Beast", "{3}{G}")
	spell := effectstest.SeedLibraryTop(me, "Spell", "Sorcery", "")
	first := effectstest.SeedLibraryTop(me, "First Draw", "Instant", "")
	hand := me.Hand.Size()

	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(first) || me.Hand.Size() != hand+1 {
		t.Fatal("the first landfall draws a card")
	}
	if !g.Battlefield.Contains(beast) {
		t.Fatal("then reveals until a creature card and puts it onto the battlefield")
	}
	if top, _ := me.Library.Top(); top.InstanceID == spell {
		t.Error("the revealed noncreature goes to the bottom")
	}

	hand = me.Hand.Size()
	battlefield := len(g.Battlefield.Cards)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Error("the second landfall still draws")
	}
	if len(g.Battlefield.Cards) != battlefield+1 {
		t.Error("the second landfall reveals nothing: only the Forest entered")
	}
}

// --- Summon: Bahamut --------------------------------------------------

func TestS58SummonBahamutChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	threat := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	second := effectstest.B12Permanent(g, opp.ID, "Their Idol", "Artifact")
	effectstest.B27Push(g, me.ID, "My Signet", "Artifact", "", "{2}", 0, 0)
	effectstest.B27Push(g, me.ID, "My Bear", "Creature — Bear", "", "{1}{G}", 2, 2)

	saga := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: saga, Name: "Summon: Bahamut", TypeLine: "Enchantment Creature — Saga Dragon",
		OracleID: effectstest.S58nBahamut, ManaCost: "{9}", Power: 9, Toughness: 9, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, saga, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Summon: Bahamut: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, saga, "flying") {
		t.Error("flying")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, threat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(threat) {
		t.Fatal("chapter I destroys the nonland permanent")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // II
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(second) {
		t.Fatal("chapter II destroys the nonland permanent")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // III: the draw step has passed
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("chapter III draws two: %d → %d", hand, me.Hand.Size())
	}

	// IV: other permanents you control, at their mana value, as it
	// resolves.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	want := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.InstanceID != saga && !c.FaceDown {
			want += c.ManaValue()
		}
	}
	if want < 4 {
		t.Fatalf("fixture: the Signet and the Bear should count at least 4, got %d", want)
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		lost := lives[p.ID] - p.Life
		if p.ID == me.ID && lost != 0 {
			t.Errorf("you take no damage, lost %d", lost)
		}
		if p.ID != me.ID && lost != want {
			t.Errorf("each opponent takes %d, %s lost %d", want, p.Name, lost)
		}
	}
	if effectstest.LoreCountersOn(g, saga) >= 0 {
		t.Error("sacrificed after IV")
	}
}
