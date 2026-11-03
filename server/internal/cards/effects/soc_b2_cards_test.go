package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b2_cards_test.go — Secrets of Strixhaven Commander, batch 2
// (#4834). Each test plays the real catalog entry and asserts what its
// text says.

const (
	socB2BitterthornOracle = "5a57fba3-01e3-4852-8c45-7c910502bd03"
	socB2BlightMoundOracle = "c4fe854e-8783-492e-bb2e-9f242ada8a5b"
	socB2BogbeastOracle    = "30f3c3be-0fe9-463d-a245-e44701aec7f2"
	socB2InfiniscopeOracle = "f74d8f53-a239-4641-90f1-bf786d57e253"
	socB2BorrowerOracle    = "c7b044c3-3cfa-407e-bf20-2875e8e04b7b"
	socB2BrudicladOracle   = "a7ac5d98-805e-4212-be94-333a0c0c44a7"
	socB2CasualtiesOracle  = "5c6c8fe7-3520-423b-a224-1c0af516871a"
	socB2CeaselessOracle   = "2f109125-b743-4ecf-84da-690d8e8ddf9c"
	socB2ChainsOracle      = "5acff7a3-bc84-4251-bd37-80231396ac38"
	socB2LoyaltyOracle     = "cdfa9de4-7d7d-4638-8ac4-793ece6372dd"
)

func TestSocB2EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB2BitterthornOracle:     "Bitterthorn, Nissa's Animus",
		socB2BlightMoundOracle:     "Blight Mound",
		socB2BogbeastOracle:        "Blossoming Bogbeast",
		socB2InfiniscopeOracle:     "Brass Infiniscope",
		socB2BorrowerOracle:        "Brazen Borrower",
		socB2BorrowerOracle + "#1": "Petty Theft",
		socB2BrudicladOracle:       "Brudiclad, Telchor Engineer",
		socB2CasualtiesOracle:      "Casualties of War",
		socB2CeaselessOracle:       "Ceaseless Conflict",
		socB2ChainsOracle:          "Chains of Custody",
		socB2LoyaltyOracle:         "Changing Loyalty",
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

// socB2Token seeds a creature token (a "Token" type line) that can attack.
func socB2Token(g *game.Game, owner uuid.UUID, name, subtype string, power, toughness int) uuid.UUID {
	return b12Creature(g, owner, name, "Token Creature — "+subtype, power, toughness)
}

// --- Bitterthorn, Nissa's Animus ---

func TestSocB2BitterthornMakesAGermThatGetsPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	thorn, germ := castLivingWeapon(t, g, "Bitterthorn, Nissa's Animus", socB2BitterthornOracle)
	if host := attachmentHostOf(t, g, thorn); host.ID != germ {
		t.Fatalf("Bitterthorn is attached to %+v, want the Germ", host)
	}
	if p, tg := effectivePower(t, g, germ), effectiveToughness(t, g, germ); p != 1 || tg != 1 {
		t.Errorf("Germ is %d/%d, want 1/1", p, tg)
	}
}

func TestSocB2BitterthornSearchesABasicLandTappedWhenEquippedCreatureAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	idle := seedBear(g, me.ID)
	thorn := seedEquipment(g, me.ID, "Bitterthorn, Nissa's Animus", socB2BitterthornOracle)
	equipTo(t, g, me.ID, thorn, bear)
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("equipped bear power %d, want 3", p)
	}
	seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear Card", TypeLine: "Creature — Bear"},
	)
	declareAttack(t, g, opp.ID, idle, bear)
	if n := triggersOnStackFrom(g, thorn); n != 1 {
		t.Fatalf("attack triggers from Bitterthorn = %d, want 1 (only the equipped creature)", n)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the search is optional, so it asks")
	}
	if searchOptionNamed(g, c, "Bear Card") != uuid.Nil {
		t.Error("a nonland card is offered")
	}
	answerSearchNamed(t, g, me.ID, "Forest")
	forest := findBattlefieldByName(g, "Forest")
	if forest == uuid.Nil {
		t.Fatal("the Forest did not reach the battlefield")
	}
	if !isTapped(g, forest) {
		t.Error("the Forest entered untapped")
	}
}

// --- Blight Mound ---

func TestSocB2BlightMoundPumpsAndGrantsMenaceOnlyToAttackingPests(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Blight Mound", "Enchantment", socB2BlightMoundOracle, 0, 0)
	pest := socB2Token(g, me.ID, "Pest", "Pest", 1, 1)
	idlePest := socB2Token(g, me.ID, "Idle Pest", "Pest", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectivePower(t, g, pest) != 1 || hasEffectiveKeyword(t, g, pest, "menace") {
		t.Fatal("a Pest that is not attacking got the bonus")
	}
	declareAttack(t, g, opp.ID, pest, bear)
	if p := effectivePower(t, g, pest); p != 2 {
		t.Errorf("attacking Pest power %d, want 2", p)
	}
	if !hasEffectiveKeyword(t, g, pest, "menace") {
		t.Error("attacking Pest lacks menace")
	}
	if effectivePower(t, g, bear) != 2 || hasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("an attacking non-Pest got the bonus")
	}
	if effectivePower(t, g, idlePest) != 1 || hasEffectiveKeyword(t, g, idlePest, "menace") {
		t.Error("a Pest that did not attack got the bonus")
	}
}

func TestSocB2BlightMoundMakesAPestWhenANontokenCreatureYouControlDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Blight Mound", "Enchantment", socB2BlightMoundOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	tok := socB2Token(g, me.ID, "Goblin", "Goblin", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b27Kill(g, tok)
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Pest"); n != 0 {
		t.Fatalf("%d Pests from a token and an opponent's creature dying, want 0", n)
	}
	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	pest := findBattlefieldByName(g, "Pest")
	if pest == uuid.Nil || b16CountNamed(g, "Pest") != 1 {
		t.Fatalf("Pests = %d, want 1", b16CountNamed(g, "Pest"))
	}
	life := me.Life
	b27Kill(g, pest)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("the Pest's dies trigger: life %d, want %d", me.Life, life+1)
	}
	if n := b16CountNamed(g, "Pest"); n != 0 {
		t.Errorf("a dying Pest token made another Pest (%d)", n)
	}
}

// --- Blossoming Bogbeast ---

func TestSocB2BogbeastGainsTwoThenPumpsByLifeGainedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	beast := b12Push(g, me.ID, "Blossoming Bogbeast", "Creature — Beast", socB2BogbeastOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// Three life gained earlier in the turn counts too.
	g.WithWriteLock(func() {
		if err := (GainLife{Player: me.ID, Amount: 3}).Apply(NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: beast})); err != nil {
			t.Fatalf("GainLife: %v", err)
		}
	})
	life := me.Life
	declareAttack(t, g, opp.ID, beast)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Fatalf("life %d, want %d", me.Life, life+2)
	}
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 7 || tg != 7 {
		t.Errorf("bear is %d/%d, want 7/7 (+5/+5 for 3 + 2 life)", p, tg)
	}
	if effectivePower(t, g, beast) != 8 {
		t.Errorf("Bogbeast power %d, want 8", effectivePower(t, g, beast))
	}
	for _, id := range []uuid.UUID{beast, bear} {
		if !hasEffectiveKeyword(t, g, id, "trample") {
			t.Error("a creature you control lacks trample")
		}
	}
	if effectivePower(t, g, theirs) != 2 || hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature was pumped")
	}
}

// --- Brass Infiniscope ---

func TestSocB2InfiniscopeDrawsAndGainsHalfXOnTheNextXSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	scope := seedPermanentWithOracle(g, me.ID, "Brass Infiniscope", "Artifact", socB2InfiniscopeOracle)
	activateManaFor(t, g, me.ID, scope, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool %d after the tap, want {C}{C}", got)
	}
	seedLibrary(me, "A", "B", "C")
	hand, life := me.Hand.Size(), me.Life
	// A spell with no {X} in its mana cost neither fires nor spends it.
	castFromHandForTest(t, g, me, "Bear", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatalf("a non-X spell fired the trigger: hand %d (was %d), life %d (was %d)", me.Hand.Size(), hand, me.Life, life)
	}
	castFromHandForTest(t, g, me, "Blaze", "Sorcery", "{X}{R}", "", game.CastSpellParams{XValue: 5})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (one card drawn)", me.Hand.Size(), hand+1)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d (half of 5, rounded down)", me.Life, life+2)
	}
	// The delayed trigger is spent.
	castFromHandForTest(t, g, me, "Blaze Two", "Sorcery", "{X}{R}", "", game.CastSpellParams{XValue: 4})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life+2 {
		t.Errorf("a second X spell fired it again: hand %d, life %d", me.Hand.Size(), me.Life)
	}
}

func TestSocB2InfiniscopeReadsXOfASpellCounteredInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	scope := seedPermanentWithOracle(g, me.ID, "Brass Infiniscope", "Artifact", socB2InfiniscopeOracle)
	activateManaFor(t, g, me.ID, scope, 0, game.ManaAbilityParams{})
	seedLibrary(me, "A", "B")
	hand, life := me.Hand.Size(), me.Life
	spell := castFromHandForTest(t, g, me, "Blaze", "Sorcery", "{X}{R}", "", game.CastSpellParams{XValue: 7})
	if counterByEffect(t, g, spell) {
		t.Fatal("the spell is still on the stack after the counter")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life+3 {
		t.Errorf("hand %d (want %d), life %d (want %d): X read from the countered spell", me.Hand.Size(), hand+1, me.Life, life+3)
	}
}

func TestSocB2InfiniscopeIgnoresAnOpponentsXSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	scope := seedPermanentWithOracle(g, me.ID, "Brass Infiniscope", "Artifact", socB2InfiniscopeOracle)
	activateManaFor(t, g, me.ID, scope, 0, game.ManaAbilityParams{})
	seedLibrary(me, "A", "B")
	hand, life := me.Hand.Size(), me.Life
	castFromHandForTest(t, g, opp, "Their Blaze", "Instant", "{X}{R}", "", game.CastSpellParams{XValue: 6})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("an opponent's X spell fired it: hand %d, life %d", me.Hand.Size(), me.Life)
	}
}

// --- Brazen Borrower // Petty Theft ---

func socB2BrazenBorrower(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   socB2BorrowerOracle,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Brazen Borrower", TypeLine: "Creature — Faerie Rogue", ManaCost: "{1}{U}{U}", Colors: []string{"U"}, Power: 3, Toughness: 1},
			{Name: "Petty Theft", TypeLine: "Instant — Adventure", ManaCost: "{1}{U}", Colors: []string{"U"}},
		},
	}
	c.SetFace(0)
	return c
}

func TestSocB2PettyTheftBouncesAnOpponentsNonlandPermanentThenTheBorrowerIsCastable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	c := socB2BrazenBorrower(me.ID)
	me.Hand.PushTop(c)
	for _, bad := range []uuid.UUID{land, mine} {
		if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad}}}); err == nil {
			t.Fatal("Petty Theft accepted a land or your own permanent")
		}
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err != nil {
		t.Fatalf("cast Petty Theft: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Fatal("the opponent's creature was not returned to its owner's hand")
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Fatal("the resolved Adventure is not in exile (CR 715.3d)")
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast Brazen Borrower from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(c.InstanceID) {
		t.Fatal("Brazen Borrower did not enter")
	}
	for _, kw := range []string{"flash", "flying"} {
		if !hasEffectiveKeyword(t, g, c.InstanceID, kw) {
			t.Errorf("Brazen Borrower lacks %s", kw)
		}
	}
}

func TestSocB2BrazenBorrowerCanBlockOnlyFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	borrower := b12Push(g, opp.ID, "Brazen Borrower", "Creature — Faerie Rogue", socB2BorrowerOracle, 3, 1)
	walker := b12Creature(g, me.ID, "Walker", "Creature — Bear", 2, 2)
	flyer := b12Creature(g, me.ID, "Flyer", "Creature — Bird", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flyer {
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			}
		}
	})
	refused := msc13BlockRefused(t, g, opp.ID, borrower, walker, flyer)
	if !refused[walker] {
		t.Error("Brazen Borrower blocked a creature without flying")
	}
	if refused[flyer] {
		t.Error("Brazen Borrower could not block a flyer")
	}
}

// --- Brudiclad, Telchor Engineer ---

func TestSocB2BrudicladMakesAMyrAndTurnsEveryOtherTokenIntoTheChosenOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	brud := b12Push(g, me.ID, "Brudiclad, Telchor Engineer", "Legendary Artifact Creature — Phyrexian Artificer", socB2BrudicladOracle, 4, 4)
	dragon := socB2Token(g, me.ID, "Dragon", "Dragon", 4, 4)
	goblin := socB2Token(g, me.ID, "Goblin", "Goblin", 1, 1)
	theirs := socB2Token(g, opp.ID, "Their Goblin", "Goblin", 1, 1)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	myr := findBattlefieldByName(g, "Phyrexian Myr")
	if myr == uuid.Nil {
		t.Fatal("no Phyrexian Myr token")
	}
	if p, tg := effectivePower(t, g, myr), effectiveToughness(t, g, myr); p != 2 || tg != 1 {
		t.Errorf("Myr is %d/%d, want 2/1", p, tg)
	}
	if !hasEffectiveKeyword(t, g, myr, "haste") {
		t.Error("a creature token you control lacks haste")
	}
	if hasEffectiveKeyword(t, g, theirs, "haste") || hasEffectiveKeyword(t, g, brud, "haste") {
		t.Error("haste reached an opponent's token or a nontoken creature")
	}
	if legs := answerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {dragon}}); legs != 1 {
		t.Fatalf("token picks = %d, want 1", legs)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{goblin, myr} {
		c, _ := g.LookupCardForEffect(id)
		if effectivePower(t, g, id) != 4 || effectiveToughness(t, g, id) != 4 {
			t.Errorf("%s is %d/%d, want a 4/4 copy of the Dragon", c.Name, effectivePower(t, g, id), effectiveToughness(t, g, id))
		}
	}
	if effectivePower(t, g, theirs) != 1 {
		t.Error("an opponent's token became a copy")
	}
	if effectivePower(t, g, brud) != 4 {
		t.Error("Brudiclad itself changed")
	}
	// The copy has no duration: still a Dragon after the turn ends.
	advanceTo(t, g, game.StepCleanup)
	if effectivePower(t, g, goblin) != 4 {
		t.Error("the copy ended at cleanup")
	}
}

func TestSocB2BrudicladChoosingNoTokenCopiesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Brudiclad, Telchor Engineer", "Legendary Artifact Creature — Phyrexian Artificer", socB2BrudicladOracle, 4, 4)
	dragon := socB2Token(g, me.ID, "Dragon", "Dragon", 4, 4)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	answerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{})
	passPriorityAroundTable(t, g)
	myr := findBattlefieldByName(g, "Phyrexian Myr")
	if myr == uuid.Nil || effectivePower(t, g, myr) != 2 || effectivePower(t, g, dragon) != 4 {
		t.Error("declining the choice changed a token")
	}
}

// --- Casualties of War ---

func TestSocB2CasualtiesOfWarDestroysOneOfEachChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	art := b12Permanent(g, opp.ID, "Signet", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ench := b12Permanent(g, opp.ID, "Pacifism", "Enchantment")
	land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	pw := b12Permanent(g, opp.ID, "Jace", "Legendary Planeswalker — Jace")
	spare := b12Creature(g, opp.ID, "Spare", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Casualties of War", TypeLine: "Sorcery", OracleID: socB2CasualtiesOracle, Owner: me.ID, Controller: me.ID})
	// A creature is not a legal target for the artifact mode.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{modeRef(game.TargetCard, bear, 0, 0)}}); err == nil {
		t.Fatal("the artifact mode accepted a creature")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1, 2, 3, 4}, Targets: []game.TargetRef{
		modeRef(game.TargetCard, art, 0, 0),
		modeRef(game.TargetCard, bear, 1, 0),
		modeRef(game.TargetCard, ench, 2, 0),
		modeRef(game.TargetCard, land, 3, 0),
		modeRef(game.TargetCard, pw, 4, 0),
	}}); err != nil {
		t.Fatalf("cast Casualties of War: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, x := range []uuid.UUID{art, bear, ench, land, pw} {
		if g.Battlefield.Contains(x) {
			c, _ := g.LookupCardForEffect(x)
			t.Errorf("%s was not destroyed", c.Name)
		}
	}
	if !g.Battlefield.Contains(spare) {
		t.Error("an untargeted creature was destroyed")
	}
}

// --- Ceaseless Conflict ---

func TestSocB2CeaselessConflictMakesASpiritPerNontokenCreatureYouLost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	tok := socB2Token(g, me.ID, "Goblin", "Goblin", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Ceaseless Conflict", "Sorcery", socB2CeaselessOracle, nil)
	passPriorityAroundTable(t, g)
	for _, x := range []uuid.UUID{a, b, tok, theirs} {
		if g.Battlefield.Contains(x) {
			t.Error("a creature survived Destroy all creatures")
		}
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.Controller == me.ID {
			n++
			if c.Power != 3 || c.Toughness != 2 {
				t.Errorf("Spirit is %d/%d, want 3/2", c.Power, c.Toughness)
			}
		}
	}
	if n != 2 {
		t.Errorf("%d Spirits, want 2 (two nontoken creatures you controlled)", n)
	}
}

// --- Chains of Custody ---

func TestSocB2ChainsOfCustodyExilesUntilItLeavesAndWardsTheHost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Chains of Custody", "Enchantment — Aura", socB2ChainsOracle, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Fatal("Chains of Custody enchanted a creature you don't control")
	}
	chains := castCatalogSpell(t, g, "Chains of Custody", "Enchantment — Aura", socB2ChainsOracle, []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Error("the exile accepted your own creature")
		}
	}
	spmB12PickIfAsked(t, g, me.ID, theirs)
	if g.Battlefield.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Fatal("the opponent's creature was not exiled")
	}
	// Ward {2} on the enchanted creature.
	blade := castAtWardedCreature(t, g, opp, mine)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the opponent's spell")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || !opp.Graveyard.Contains(blade) {
		t.Fatal("a declined ward did not counter the spell")
	}
	// The Aura leaves: the exiled card returns.
	b27Kill(g, chains)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(chains) {
		t.Fatal("Chains of Custody is still on the battlefield")
	}
	if n := namedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("after Chains of Custody left, %d copies under its owner's control, want 1", n)
	}
}

// --- Changing Loyalty ---

func TestSocB2ChangingLoyaltyReturnsTheDeadCreatureUnderYourControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Changing Loyalty", "Enchantment — Aura", socB2LoyaltyOracle, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	back := findBattlefieldByName(g, "Their Bear")
	if back == uuid.Nil {
		t.Fatal("the dead creature did not return")
	}
	c, _ := g.LookupCardForEffect(back)
	if c.Controller != me.ID || c.Owner != opp.ID {
		t.Errorf("returned under %s (owner %s), want your control", c.Controller, c.Owner)
	}
}

func TestSocB2ChangingLoyaltyReplicatesOntoANewCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Changing Loyalty", TypeLine: "Enchantment — Aura", ManaCost: "{1}{B}",
		OracleID: socB2LoyaltyOracle, Owner: me.ID, Controller: me.ID})
	floatForTest(g, me, "BBBBB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}}, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast with one replicate: %v", err)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool %d after {1}{B} + replicate {2}, want 1 left of 5", got)
	}
	if it := triggerOnStack(g, id); it == nil || it.XValue != 1 {
		t.Fatalf("replicate trigger = %+v, want one counting one payment", it)
	}
	for i := 0; i < 64; i++ {
		if stackFullyEmpty(g) && latestPickTarget(g, me.ID) == nil {
			break
		}
		if p := latestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: b}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	auras := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Changing Loyalty" {
			if host := attachmentHostOf(t, g, c.InstanceID); host.Kind == game.TargetCard {
				auras[host.ID] = IsToken(c)
			}
		}
	}
	if tok, ok := auras[a]; !ok || tok {
		t.Errorf("the original Aura is not on Bear A as a card (%v)", auras)
	}
	if tok, ok := auras[b]; !ok || !tok {
		t.Errorf("the copy is not a token Aura on Bear B (%v)", auras)
	}
}
