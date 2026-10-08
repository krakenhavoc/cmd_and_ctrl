package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s58_inverse_nissa_cards_test.go — the Inverse Nissa deck request
// (#2741): the ten missing cards whose seams already exist, plus the
// review stamp on Supreme Verdict.

const (
	s58nStockUp         = "2251e34e-4ce8-4452-9dc6-d8cce8583996"
	s58nMemoryDeluge    = "e6fd55f2-7e26-469c-a44a-ea2eb90e19a9"
	s58nFatefulAbsence  = "35bba442-1aec-4d33-b502-4c580d61644b"
	s58nJeskaiRev       = "61b6b1d1-4350-41c1-ac47-835b2831f24a"
	s58nWhiteAuracite   = "6022608a-6cf2-45bd-adec-63211710a5ed"
	s58nTablet          = "19cf5798-4600-4a66-b1c7-de77bde157d0"
	s58nWeatherMaker    = "4097fea2-ae34-4b99-ab7e-a276d608b489"
	s58nAvacyn          = "4af61d45-7114-4eb4-a905-9e03caf87454"
	s58nNissa           = "6067b8d0-5b00-402c-967d-e35a1304fafc"
	s58nBahamut         = "fe9f6825-597b-4d80-ab7c-f4ee4e824b6f"
	s58nSupremeVerdict  = "0230de18-8d15-4cfa-9d42-7ccddd9f9570"
	s58nWeatherMakerDmg = 0 // Weather Maker's only activated (non-mana) ability
)

func TestS58InverseNissaRegistered(t *testing.T) {
	want := map[string]string{
		s58nStockUp:        "Stock Up",
		s58nMemoryDeluge:   "Memory Deluge",
		s58nFatefulAbsence: "Fateful Absence",
		s58nJeskaiRev:      "Jeskai Revelation",
		s58nWhiteAuracite:  "White Auracite",
		s58nTablet:         "Tablet of Discovery",
		s58nWeatherMaker:   "Weather Maker",
		s58nAvacyn:         "Avacyn, Angel of Horror",
		s58nNissa:          "Nissa, Leyline Tamer",
		s58nBahamut:        "Summon: Bahamut",
		s58nSupremeVerdict: "Supreme Verdict",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s: registered=%v name=%q", name, ok, spec.Name)
			continue
		}
		wantStamp := CompletenessFull
		if oracle == s58nMemoryDeluge {
			wantStamp = CompletenessCaveats
		}
		if spec.Completeness != wantStamp {
			t.Errorf("%s: completeness %v, want %v", name, spec.Completeness, wantStamp)
		}
	}
}

// --- Stock Up ---------------------------------------------------------

func TestS58StockUpTakesTwoOfFiveAndOrdersTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedLibrary(me, "A", "B", "C", "D", "E", "Sixth")
	hand := me.Hand.Size()

	castCatalogSpell(t, g, "Stock Up", "Sorcery", s58nStockUp, nil)
	passPriorityAroundTable(t, g)

	pick := latestChooseCardsFor(g, me.ID)
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
	order := putInLibraryChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "1", "2", "3", "4", "5", "6", "7", "8", "9", "10")
	advanceToMain(t, g)
	floatForTest(g, me, "UUUU")
	deluge := castFromHandForTest(t, g, me, "Memory Deluge", "Instant", "{2}{U}{U}", s58nMemoryDeluge, strict)
	passPriorityAroundTable(t, g)

	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 4 || pick.ChooseMax != 2 {
		t.Fatalf("four mana spent: take two of four, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if putInLibraryChoiceFor(g, me.ID) != nil {
		t.Error("the rest go to the bottom in a random order, with no ordering prompt")
	}
	if !me.Graveyard.Contains(deluge) {
		t.Fatal("Memory Deluge is in the graveyard after resolving")
	}

	floatForTest(g, me, "UUUUUUU")
	if err := g.CastSpell(me.ID, deluge, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Strict: true,
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick = latestChooseCardsFor(g, me.ID)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	clues := func(owner uuid.UUID) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Clue" && c.Controller == owner {
				n++
			}
		}
		return n
	}
	castCatalogSpell(t, g, "Fateful Absence", "Instant", s58nFatefulAbsence,
		[]game.TargetRef{{Kind: game.TargetCard, ID: walker}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Fatal("the creature is destroyed")
	}
	if clues(opp.ID) != 1 || clues(me.ID) != 0 {
		t.Errorf("its controller investigates: theirs %d, mine %d", clues(opp.ID), clues(me.ID))
	}
}

// --- Jeskai Revelation ------------------------------------------------

func TestS58JeskaiRevelationDoesAllFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life, hand, oppLife := me.Life, me.Hand.Size(), opp.Life

	castCatalogSpell(t, g, "Jeskai Revelation", "Instant", s58nJeskaiRev, []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs, Slot: 0}, {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1},
	})
	passPriorityAroundTable(t, g)

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
			if !hasEffectiveKeyword(t, g, c.InstanceID, "prowess") {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 5)
	hand := me.Hand.Size()

	castCatalogSpell(t, g, "Jeskai Revelation", "Instant", s58nJeskaiRev, []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs, Slot: 0}, {Kind: game.TargetCard, ID: other, Slot: 1},
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)

	if pr6Marked(g, other) != 4 {
		t.Errorf("the damage still happens: %d marked, want 4", pr6Marked(g, other))
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("the draw still happens: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- White Auracite ---------------------------------------------------

func TestS58WhiteAuraciteExilesUntilItLeavesAndTapsForWhite(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	threat := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	rock := castCatalogSpell(t, g, "White Auracite", "Artifact", s58nWhiteAuracite, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); p != nil && hasID(p.PickTargetCards, mine) {
		t.Error("only a permanent an opponent controls can be chosen")
	}
	pickCard(t, g, me.ID, threat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(threat) {
		t.Fatal("the opposing creature is exiled")
	}

	activateManaFor(t, g, me.ID, rock, 0, game.ManaAbilityParams{})
	if got := gmrPoolCount(me, "W"); got != 1 {
		t.Errorf("{T}: Add {W} — pool has %d {W}", got)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	g.RunStateChecksForTest()
	if findBattlefieldByName(g, "Their Threat") == uuid.Nil {
		t.Error("the exiled card returns when White Auracite leaves")
	}
}

// --- Tablet of Discovery ----------------------------------------------

// The milled card can be played this turn, and a land is played, not
// cast.
func TestS58TabletOfDiscoveryMillsACardYouMayPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := seedLibraryTop(me, "Mountain", "Basic Land — Mountain", "")

	castCatalogSpell(t, g, "Tablet of Discovery", "Artifact", s58nTablet, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
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
	spec, _ := Lookup(s58nTablet)
	if len(spec.ManaAbilities) != 2 {
		t.Fatalf("two mana abilities, got %d", len(spec.ManaAbilities))
	}
	if spec.ManaAbilities[0].Restrictions != nil {
		t.Error("{T}: Add {R} is unrestricted")
	}
	got := spec.ManaAbilities[1].Restrictions
	want := []string{ManaRestrictCast, ManaRestrictAnyType("Instant", "Sorcery")}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("{R}{R} restrictions = %v, want %v", got, want)
	}
}

// --- Weather Maker ----------------------------------------------------

func TestS58WeatherMakerBanksLandfallAndSpendsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	maker := b27Push(g, me.ID, "Weather Maker", "Artifact", s58nWeatherMaker, "{3}", 0, 0)
	for i := 0; i < 5; i++ {
		playLandFromHand(t, g, "Forest", "")
		passPriorityAroundTable(t, g)
	}
	if got := counterCount(g, maker, game.CounterCharge); got != 5 {
		t.Fatalf("five landfalls, %d charge counters", got)
	}

	activateManaFor(t, g, me.ID, maker, 1, game.ManaAbilityParams{})
	if got := counterCount(g, maker, game.CounterCharge); got != 3 {
		t.Errorf("{C}{C} removes two: %d left", got)
	}
	if got := gmrPoolCount(me, "C"); got != 2 {
		t.Errorf("pool has %d {C}, want 2", got)
	}

	untapForTest(g, maker)
	life := opp.Life
	b16Activate(t, g, me.ID, maker, s58nWeatherMakerDmg, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("3 damage: %d → %d", life, opp.Life)
	}
	if got := counterCount(g, maker, game.CounterCharge); got != 0 {
		t.Errorf("the damage removes three: %d left", got)
	}
}

// --- Avacyn, Angel of Horror ------------------------------------------

func TestS58AvacynReturnsYourNontokenCreaturesAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	avacyn := b12Push(g, me.ID, "Avacyn, Angel of Horror", "Legendary Creature — Angel", s58nAvacyn, 6, 6)
	if !hasEffectiveKeyword(t, g, avacyn, "flying") || !hasEffectiveKeyword(t, g, avacyn, "deathtouch") {
		t.Error("flying, deathtouch")
	}
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	goblin := findBattlefieldByName(g, "Goblin")
	advanceToMain(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if !b19TriggerQueuedFrom(g, avacyn) {
		t.Fatal("another nontoken creature you control died")
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(goblin) })
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	if b19TriggerQueuedFrom(g, avacyn) {
		t.Error("a token, and an opponent's creature, don't trigger it")
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bear) {
		t.Fatal("the card waits in the graveyard until the end step")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the card returns to the battlefield at the beginning of the end step")
	}
	if c := findBattlefieldCardForTest(g, bear); c.Controller != me.ID {
		t.Error("under your control")
	}
}

// Avacyn's own death triggers too, and the card that left the graveyard
// before the end step stays where it went.
func TestS58AvacynReturnsHerselfButNotACardThatLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	avacyn := b12Push(g, me.ID, "Avacyn, Angel of Horror", "Legendary Creature — Angel", s58nAvacyn, 6, 6)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(avacyn) })
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardUnderControlForEffect(bear, game.ZoneHand, uuid.Nil)
	})
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(avacyn) {
		t.Error("Avacyn's own death returns her")
	}
	if !me.Hand.Contains(bear) {
		t.Error("a card that left the graveyard is a new object and stays in hand")
	}
}

// --- Nissa, Leyline Tamer ---------------------------------------------

func TestS58NissaDrawsEachLandfallAndRevealsOnlyTheFirst(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nissa := b12Push(g, me.ID, "Nissa, Leyline Tamer", "Legendary Creature — Elf Wizard", s58nNissa, 4, 5)
	if !hasEffectiveKeyword(t, g, nissa, "deathtouch") || !hasEffectiveKeyword(t, g, nissa, "vigilance") {
		t.Error("deathtouch, vigilance")
	}
	// Top to bottom: a card to draw, a sorcery, a creature, and more.
	seedLibraryTop(me, "Filler A", "Sorcery", "")
	beast := seedLibraryTop(me, "Beast", "Creature — Beast", "{3}{G}")
	spell := seedLibraryTop(me, "Spell", "Sorcery", "")
	first := seedLibraryTop(me, "First Draw", "Instant", "")
	hand := me.Hand.Size()

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
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
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Error("the second landfall still draws")
	}
	if len(g.Battlefield.Cards) != battlefield+1 {
		t.Error("the second landfall reveals nothing: only the Forest entered")
	}
}

// --- Summon: Bahamut --------------------------------------------------

func TestS58SummonBahamutChapters(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	threat := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	second := b12Permanent(g, opp.ID, "Their Idol", "Artifact")
	b27Push(g, me.ID, "My Signet", "Artifact", "", "{2}", 0, 0)
	b27Push(g, me.ID, "My Bear", "Creature — Bear", "", "{1}{G}", 2, 2)

	saga := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: saga, Name: "Summon: Bahamut", TypeLine: "Enchantment Creature — Saga Dragon",
		OracleID: s58nBahamut, ManaCost: "{9}", Power: 9, Toughness: 9, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, saga, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Summon: Bahamut: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, saga, "flying") {
		t.Error("flying")
	}
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, threat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(threat) {
		t.Fatal("chapter I destroys the nonland permanent")
	}

	advanceToPrecombatMainOf(t, g, seat) // II
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, second)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(second) {
		t.Fatal("chapter II destroys the nonland permanent")
	}

	advanceToPrecombatMainOf(t, g, seat) // III: the draw step has passed
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("chapter III draws two: %d → %d", hand, me.Hand.Size())
	}

	// IV: other permanents you control, at their mana value, as it
	// resolves.
	advanceToPrecombatMainOf(t, g, seat)
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
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		lost := lives[p.ID] - p.Life
		if p.ID == me.ID && lost != 0 {
			t.Errorf("you take no damage, lost %d", lost)
		}
		if p.ID != me.ID && lost != want {
			t.Errorf("each opponent takes %d, %s lost %d", want, p.Name, lost)
		}
	}
	if loreCountersOn(g, saga) >= 0 {
		t.Error("sacrificed after IV")
	}
}
