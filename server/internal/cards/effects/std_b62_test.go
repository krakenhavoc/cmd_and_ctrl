package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b62_test.go — Standard leftovers, batch 62 (#4452): Lorwyn Eclipsed.

const (
	stdB62TrapperOracle   = "915d70ac-46ce-41d0-a137-f17e7aaea87d"
	stdB62ScoundrelOracle = "6d7f7e6e-5076-4af7-afe3-3e500835c4aa"
	stdB62SpinnersOracle  = "b608288e-1755-4738-b828-9df6998bae8d"
	stdB62MysticOracle    = "ed428638-181b-47cb-92f4-379db020e634"
	stdB62DreamsOracle    = "f75568a0-33ed-45a0-a885-48a19a5bb58b"
	stdB62NoggleOracle    = "b49efbe8-e15b-4dc7-82de-970cbce92b79"
	stdB62OkoOracle       = "a0c956dc-d7de-4f11-b76f-7537fbee1731"
	stdB62OmniOracle      = "66d4370e-19b9-4680-9b1e-a192537bc002"
	stdB62WellguardOracle = "b0dbfd40-ecc5-493b-976b-bc3f62137fcf"
	stdB62RimeChillOracle = "bad84deb-3f2c-46fd-9dcb-4e3ba9732769"
)

func TestStdB62EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB62TrapperOracle:    "Wanderbrine Trapper",
		stdB62ScoundrelOracle:  "Gravelgill Scoundrel",
		stdB62SpinnersOracle:   "Illusion Spinners",
		stdB62MysticOracle:     "Kulrath Mystic",
		stdB62DreamsOracle:     "Lofty Dreams",
		stdB62NoggleOracle:     "Noggle the Mind",
		stdB62OkoOracle:        "Oko, Lorwyn Liege",
		stdB62OkoOracle + "#1": "Oko, Shadowmoor Scion",
		stdB62OmniOracle:       "Omni-Changeling",
		stdB62WellguardOracle:  "Pestered Wellguard",
		stdB62RimeChillOracle:  "Rime Chill",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func stdB62Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
		g.EmitEvent(game.Event{Kind: game.EventTapCard, CardID: id})
	})
}

func TestStdB62TrapperTapsAnOpposingCreatorePayingWithAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	trapper := pushCatalogPermanent(g, me.ID, "Wanderbrine Trapper", "Creature — Merfolk Scout", stdB62TrapperOracle, false)
	friend := pushVanillaCreature(g, me.ID, "Friend", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	b16Activate(t, g, me.ID, trapper, 0, game.ActivateAbilityParams{
		TapIDs:  []uuid.UUID{friend},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	})
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirs) || !b16Tapped(t, g, friend) || !b16Tapped(t, g, trapper) {
		t.Errorf("tapped: target=%v friend=%v trapper=%v, want all three", b16Tapped(t, g, theirs), b16Tapped(t, g, friend), b16Tapped(t, g, trapper))
	}
}

func TestStdB62TrapperCannotTargetItsOwnSide(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	trapper := pushCatalogPermanent(g, me.ID, "Wanderbrine Trapper", "Creature — Merfolk Scout", stdB62TrapperOracle, false)
	friend := pushVanillaCreature(g, me.ID, "Friend", 1, 1)
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, trapper, 0, game.ActivateAbilityParams{Strict: true,
		TapIDs:  []uuid.UUID{friend},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}}); err == nil {
		t.Error("my own creature was accepted as the target")
	}
}

func TestStdB62ScoundrelTapsAnotherCreatureToBecomeUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scoundrel := pushCatalogPermanent(g, me.ID, "Gravelgill Scoundrel", "Creature — Merfolk Rogue", stdB62ScoundrelOracle, false)
	friend := pushVanillaCreature(g, me.ID, "Friend", 1, 1)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, scoundrel)
	lockInAttacks(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		passPriorityAroundTable(t, g)
	}
	answerChooseCards(t, g, me.ID, friend)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, friend) {
		t.Error("the other creature was not tapped")
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, scoundrel); err == nil {
		t.Error("the Scoundrel was blocked")
	}
}

func TestStdB62ScoundrelDecliningLeavesItBlockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scoundrel := pushCatalogPermanent(g, me.ID, "Gravelgill Scoundrel", "Creature — Merfolk Rogue", stdB62ScoundrelOracle, false)
	pushVanillaCreature(g, me.ID, "Friend", 1, 1)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, scoundrel)
	lockInAttacks(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		passPriorityAroundTable(t, g)
	}
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, scoundrel); err != nil {
		t.Errorf("the Scoundrel could not be blocked after declining: %v", err)
	}
}

func TestStdB62SpinnersHaveHexproofOnlyWhileUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spinners := pushCatalogPermanent(g, me.ID, "Illusion Spinners", "Creature — Faerie Wizard", stdB62SpinnersOracle, false)
	g.RunStateChecksForTest()
	if !hasKeywordOnBattlefield(t, g, spinners, "hexproof") {
		t.Error("untapped Spinners lack hexproof")
	}
	stdB62Tap(g, spinners)
	g.RunStateChecksForTest()
	if hasKeywordOnBattlefield(t, g, spinners, "hexproof") {
		t.Error("tapped Spinners still have hexproof")
	}
}

func TestStdB62SpinnersCanBeCastAtInstantSpeedOnlyWithAFaerie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	// On the opponent's turn I could only cast it with flash.
	advanceToMain(t, g)
	card := game.NewCard("Illusion Spinners", me.ID)
	card.TypeLine, card.ManaCost, card.OracleID = "Creature — Faerie Wizard", "{4}{U}", stdB62SpinnersOracle
	card.Power, card.Toughness = 4, 3
	me.Hand.PushTop(card)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{U}{U}{U}"); err != nil {
		t.Fatal(err)
	}
	// Main phase of my own turn with an empty stack is sorcery timing: castable regardless.
	// The flash gate is exercised by the registry: the clause exists and tests the board.
	spec, _ := Lookup(stdB62SpinnersOracle)
	if spec.FlashIf == nil {
		t.Fatal("no flash clause")
	}
	_ = opp
	var without, with bool
	g.ReadSnapshot(func() { without = spec.FlashIf.Condition(g, me.ID) })
	b12Creature(g, me.ID, "Pixie", "Creature — Faerie", 1, 1)
	g.ReadSnapshot(func() { with = spec.FlashIf.Condition(g, me.ID) })
	if without || !with {
		t.Errorf("flash condition without a Faerie=%v with one=%v, want false/true", without, with)
	}
}

func TestStdB62MysticGrowsOnAFourDropOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	mystic := pushCatalogPermanent(g, me.ID, "Kulrath Mystic", "Creature — Elemental Wizard", stdB62MysticOracle, false)
	cheap := game.NewCard("Cheap", me.ID)
	cheap.TypeLine, cheap.ManaCost = "Sorcery", "{1}"
	me.Hand.PushTop(cheap)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, cheap.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	before := gaLayered(t, g, mystic).CurrentPower()
	dear := game.NewCard("Dear", me.ID)
	dear.TypeLine, dear.ManaCost = "Sorcery", "{3}{R}"
	me.Hand.PushTop(dear)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}{R}{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, dear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	after := gaLayered(t, g, mystic).CurrentPower()
	if after != before+2 {
		t.Errorf("power %d → %d after a mana value 4 spell, want +2 (the one-drop gave nothing)", before, after)
	}
	if !hasKeywordOnBattlefield(t, g, mystic, "vigilance") {
		t.Error("no vigilance")
	}
}

func TestStdB62LoftyDreamsDrawsAndGivesPlusTwoAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B")
	advanceToMain(t, g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hand := me.Hand.Size()
	aura := game.NewCard("Lofty Dreams", me.ID)
	aura.TypeLine, aura.ManaCost, aura.OracleID = "Enchantment — Aura", "{3}{U}{U}", stdB62DreamsOracle
	me.Hand.PushTop(aura)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{U}{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, aura.InstanceID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Errorf("Bear = %d/%d, want a 4/4 flier", c.CurrentPower(), c.CurrentToughness())
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1 (the Aura left it, a card was drawn)", hand, me.Hand.Size())
	}
}

func TestStdB62NoggleMakesAColorlessOneOneNoggleWithNoAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	dragon := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dragon", TypeLine: "Creature — Dragon", Power: 5, Toughness: 5,
		Colors: []string{"R"}, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	aura := game.NewCard("Noggle the Mind", me.ID)
	aura.TypeLine, aura.ManaCost, aura.OracleID = "Enchantment — Aura", "{1}{U}", stdB62NoggleOracle
	me.Hand.PushTop(aura)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, aura.InstanceID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dragon}}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	d := gaLayered(t, g, dragon)
	if d.CurrentPower() != 1 || d.CurrentToughness() != 1 {
		t.Errorf("Dragon = %d/%d, want 1/1", d.CurrentPower(), d.CurrentToughness())
	}
	if len(d.EffectiveColors()) != 0 {
		t.Errorf("colors %v, want colorless", d.EffectiveColors())
	}
	if !d.HasSubtype("Noggle") || d.HasSubtype("Dragon") {
		t.Errorf("Noggle=%v Dragon=%v, want a Noggle that is no longer a Dragon", d.HasSubtype("Noggle"), d.HasSubtype("Dragon"))
	}
	if !d.IsCreature() {
		t.Error("it stopped being a creature")
	}
	if hasKeywordOnBattlefield(t, g, dragon, "flying") {
		t.Error("it kept flying")
	}
}

func TestStdB62OmniChangelingEntersAsACopyThatHasChangeling(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	angel := seedCopyableCreature(g, me.ID, "Serra Angel", "Creature — Angel", 4, 4)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Omni-Changeling", TypeLine: "Creature — Shapeshifter",
		ManaCost: "{3}{U}{U}", OracleID: stdB62OmniOracle, Owner: me.ID, Controller: me.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{U}{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if copyPrompt(g) == nil {
		t.Fatal("no copy prompt")
	}
	resolveWithCopyChoice(t, g, angel)
	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Serra Angel" || got.CurrentPower() != 4 {
		t.Errorf("entered as %q %d/%d, want a copy of Serra Angel", got.Name, got.CurrentPower(), got.CurrentToughness())
	}
	if !got.HasSubtype("Kithkin") || !got.HasSubtype("Angel") {
		t.Errorf("Kithkin=%v Angel=%v, want every creature type (changeling)", got.HasSubtype("Kithkin"), got.HasSubtype("Angel"))
	}
}

func TestStdB62OmniChangelingDeclinedStaysAZeroZeroChangeling(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCopyableCreature(g, me.ID, "Serra Angel", "Creature — Angel", 4, 4)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Omni-Changeling", TypeLine: "Creature — Shapeshifter",
		ManaCost: "{3}{U}{U}", OracleID: stdB62OmniOracle, Owner: me.ID, Controller: me.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{U}{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	resolveWithCopyChoice(t, g, uuid.Nil)
	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Omni-Changeling" || got.IsCopy() || got.CurrentToughness() != 0 {
		t.Errorf("name=%q copy=%v toughness=%d, want the uncopied 0/0 Omni-Changeling", got.Name, got.IsCopy(), got.CurrentToughness())
	}
	if !got.HasSubtype("Kithkin") {
		t.Error("a declined Omni-Changeling lost its changeling")
	}
}

func TestStdB62WellguardMakesAFaerieWhenItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := pushCatalogPermanent(g, me.ID, "Pestered Wellguard", "Creature — Merfolk Soldier", stdB62WellguardOracle, false)
	declareAttack(t, g, opp.ID, guard)
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Faerie"); got != 1 {
		t.Fatalf("%d Faeries, want 1", got)
	}
	f := mshB1Token(t, g, me.ID, "Faerie")
	if f.Power != 1 || f.Toughness != 1 || len(f.Colors) != 2 || !hasKeywordOnBattlefield(t, g, f.InstanceID, "flying") {
		t.Errorf("Faerie = %d/%d %v, want a 1/1 blue and black flier", f.Power, f.Toughness, f.Colors)
	}
}

func TestStdB62RimeChillTapsStunsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B")
	advanceToMain(t, g)
	a := pushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "B", 2, 2)
	hand := me.Hand.Size()
	card := game.NewCard("Rime Chill", me.ID)
	card.TypeLine, card.ManaCost, card.OracleID = "Instant", "{6}{U}", stdB62RimeChillOracle
	me.Hand.PushTop(card)
	// Colorless permanents only: the full {6}{U} is owed.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{U}{U}{U}{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !b16Tapped(t, g, id) || countersOn(g, id, game.CounterStun) != 1 {
			t.Errorf("creature not tapped with a stun counter (tapped=%v stun=%d)", b16Tapped(t, g, id), countersOn(g, id, game.CounterStun))
		}
	}
	if me.Hand.Size() != hand+1-1+1 && me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d", hand, me.Hand.Size())
	}
}

func TestStdB62RimeChillCostsOneLessPerColorAmongYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	a := pushVanillaCreature(g, opp.ID, "A", 2, 2)
	for _, col := range []string{"W", "R", "G"} {
		c := game.NewCard("Thing "+col, me.ID)
		c.TypeLine, c.Colors, c.Power, c.Toughness = "Creature — Elf", []string{col}, 1, 1
		pushBattlefieldCardWithTimestamp(g, c)
	}
	card := game.NewCard("Rime Chill", me.ID)
	card.TypeLine, card.ManaCost, card.OracleID = "Instant", "{6}{U}", stdB62RimeChillOracle
	me.Hand.PushTop(card)
	// Three colors among my permanents: {6}{U} costs {3}{U}, four mana.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}}}); err != nil {
		t.Fatalf("cast for four mana with three colors out: %v", err)
	}
}

// --- Oko, Lorwyn Liege // Oko, Shadowmoor Scion ---------------------------------

func okoRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Oko, Lorwyn Liege // Oko, Shadowmoor Scion",
		Layout:   "transform",
		TypeLine: "Legendary Planeswalker — Oko // Legendary Planeswalker — Oko",
		OracleID: uuid.MustParse(stdB62OkoOracle),
		Keywords: []string{"Transform"},
		CardFaces: []cards.CardFace{
			{Name: "Oko, Lorwyn Liege", TypeLine: "Legendary Planeswalker — Oko", ManaCost: "{2}{U}", Colors: []string{"U"}, Loyalty: "3",
				OracleText: "At the beginning of your first main phase, you may pay {G}. If you do, transform Oko.\n+2: Up to one target creature gains all creature types. (This effect doesn't end.)\n+1: Target creature gets -2/-0 until your next turn."},
			{Name: "Oko, Shadowmoor Scion", TypeLine: "Legendary Planeswalker — Oko", Colors: []string{"G"}, Loyalty: "3",
				OracleText: "At the beginning of your first main phase, you may pay {U}. If you do, transform Oko.\n−1: Mill three cards. You may put a permanent card from among them into your hand.\n−3: Create two 3/3 green Elk creature tokens.\n−6: Choose a creature type. You get an emblem with \"Creatures you control of the chosen type get +3/+3 and have vigilance and hexproof.\""},
		},
	}
}

// okoPut puts the imported card onto p's battlefield, front face up, with
// `loyalty` loyalty counters.
func okoPut(t *testing.T, g *game.Game, p *game.Player, loyalty int) uuid.UUID {
	t.Helper()
	id := importToHand(okoRow(), p)
	c, ok := cardInZone(p.Hand, id)
	if !ok {
		t.Fatal("Oko did not reach the hand")
	}
	p.Hand.Remove(id)
	c.Counters = map[string]int{game.CounterLoyalty: loyalty}
	pushBattlefieldCardWithTimestamp(g, c)
	return id
}

func TestStdB62OkoPlusTwoGivesAllCreatureTypesForGood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	oko := okoPut(t, g, me, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, oko, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	passPriorityAroundTable(t, g)
	if loyaltyCount(g, oko) != 5 {
		t.Errorf("loyalty %d, want 5", loyaltyCount(g, oko))
	}
	if c := gaLayered(t, g, bear); !c.HasSubtype("Kithkin") || !c.HasSubtype("Bear") {
		t.Error("the Bear is not every creature type")
	}
}

func TestStdB62OkoPlusOneShrinksPowerUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	toMain(t, g)
	oko := okoPut(t, g, me, 3)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 3, 3)
	b16Activate(t, g, me.ID, oko, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}})
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, theirs); c.CurrentPower() != 1 || c.CurrentToughness() != 3 {
		t.Errorf("target = %d/%d, want 1/3", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestStdB62OkoTransformsForGAndMakesElk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oko := okoPut(t, g, me, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	monSettle(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	answerPayUnless(t, g, me.ID, true)
	monSettle(t, g)
	if f := iwFace(g, oko); f != 1 {
		t.Fatalf("Oko is on face %d after paying {G}, want the back face", f)
	}
	if loyaltyCount(g, oko) != 3 {
		t.Errorf("loyalty %d after transforming, want 3 (counters stay)", loyaltyCount(g, oko))
	}
	b16Activate(t, g, me.ID, oko, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Elk"); got != 2 {
		t.Fatalf("%d Elk, want 2", got)
	}
	e := mshB1Token(t, g, me.ID, "Elk")
	if e.Power != 3 || e.Toughness != 3 || len(e.Colors) != 1 || e.Colors[0] != "G" {
		t.Errorf("Elk = %d/%d %v, want a 3/3 green", e.Power, e.Toughness, e.Colors)
	}
}

func TestStdB62OkoMinusSixMakesAnEmblemForTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	toMain(t, g)
	oko := okoPut(t, g, me, 7)
	iwTransform(t, g, oko)
	kithkin := b12Creature(g, me.ID, "Kithkin", "Creature — Kithkin", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Kithkin", "Creature — Kithkin", 1, 1)
	b16Activate(t, g, me.ID, oko, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceCreatureType, me.ID)
	if c == nil {
		t.Fatal("no creature-type prompt")
	}
	if err := g.ResolveCreatureTypeChoice(c.ID, me.ID, "Kithkin"); err != nil {
		t.Fatalf("choose Kithkin: %v", err)
	}
	passPriorityAroundTable(t, g)
	if k := gaLayered(t, g, kithkin); k.CurrentPower() != 4 || k.CurrentToughness() != 4 {
		t.Errorf("my Kithkin = %d/%d, want 4/4", k.CurrentPower(), k.CurrentToughness())
	}
	for _, kw := range []string{"vigilance", "hexproof"} {
		if !hasKeywordOnBattlefield(t, g, kithkin, kw) {
			t.Errorf("my Kithkin lacks %s", kw)
		}
	}
	if b := gaLayered(t, g, bear); b.CurrentPower() != 2 {
		t.Errorf("a Bear got the emblem's bonus (%d/%d)", b.CurrentPower(), b.CurrentToughness())
	}
	if o := gaLayered(t, g, theirs); o.CurrentPower() != 1 {
		t.Errorf("an opponent's Kithkin got the bonus (%d/%d)", o.CurrentPower(), o.CurrentToughness())
	}
}

func TestStdB62OkoMinusOneMillsThreeAndTakesAPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	oko := okoPut(t, g, me, 3)
	iwTransform(t, g, oko)
	top := seedTypedLibrary(me, "Creature — Bear", "Deep", "Third", "Second", "First")
	_ = top
	b16Activate(t, g, me.ID, oko, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	pc := chooseCardsChoiceFor(g, me.ID)
	if pc == nil {
		t.Fatal("no take prompt after the mill")
	}
	if len(pc.ChooseCards) != 3 {
		t.Errorf("%d candidates, want the 3 milled permanent cards", len(pc.ChooseCards))
	}
	hand := me.Hand.Size()
	answerChooseCards(t, g, me.ID, pc.ChooseCards[0])
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}
