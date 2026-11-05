package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b40_cards_test.go — Star Trek Commander, batch 40 (#4974). Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB40KorOracle     = "fc8d8c51-428c-4b2f-965d-c43ec94997c3"
	trekB40LocutusOracle = "d0cff43a-2a22-4e05-9f71-983cb297989f"
	trekB40ArexOracle    = "51a2671b-677e-434c-9f1a-d768fe1190c9"
	trekB40McCoyOracle   = "c71c7e0c-64d8-486a-9e66-1b343868eb0a"
	trekB40OneOracle     = "eecdf18b-7a85-4b29-9a8a-26ee40da9e7a"
	trekB40ChekovOracle  = "473aba5a-5c87-4294-a556-f8aca0e5fe39"
	trekB40BarclayOracle = "adcf571f-c465-44c6-8941-48313810e732"
	trekB40RiteOracle    = "08c6ca45-09ff-4fe4-bc9b-45876e94f430"
	trekB40SinsOracle    = "bce974a2-d9ad-4ef5-bc2e-7aa2833a39f5"
	trekB40ParisOracle   = "327458f1-d3c4-4e66-8abd-a6533bef63f7"
)

func TestTrekB40EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB40KorOracle:     "Kor, Cause of Vengeance",
		trekB40LocutusOracle: "Locutus of Borg",
		trekB40ArexOracle:    "Lt. Arex, Ever-Adapting",
		trekB40McCoyOracle:   "McCoy, Not a Bricklayer",
		trekB40OneOracle:     "One, Borg Evolved",
		trekB40ChekovOracle:  "Pavel Chekov, First Posting",
		trekB40BarclayOracle: "Reg Barclay, Fidgety Fantasist",
		trekB40RiteOracle:    "Rite of Ascension",
		trekB40SinsOracle:    "Sins of the Father",
		trekB40ParisOracle:   "Tom Paris, Risk-Taker",
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

// --- fixtures ------------------------------------------------------------------

// trekB40PT is a permanent's current power and toughness.
func trekB40PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectivePTOf(t, g, id)
}

// trekB40Tokens lists the tokens named `name` that p controls.
func trekB40Tokens(g *game.Game, p uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == p && c.IsToken() && c.Name == name {
				out = append(out, c)
			}
		}
	})
	return out
}

// trekB40Draw has p draw one card and settles what it triggers, picking
// `pick` for a trigger that targets.
func trekB40Draw(t *testing.T, g *game.Game, p *game.Player, pick *game.TargetRef) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, 1) })
	assistCardSettle(t, g, p.ID, pick, true)
}

// trekB40Library replaces p's library with the given cards, the first
// one on top, and returns their IDs in the same order.
func trekB40Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

// trekB40ResetDraws starts p's turn tally of draws at zero, so a test
// counts "the second card each turn" from its own draws.
func trekB40ResetDraws(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	g.WithWriteLock(func() {
		if tally, ok := g.TurnTally.Players[p.ID]; ok {
			tally.CardsDrawn = 0
			g.TurnTally.Players[p.ID] = tally
		}
	})
	if n := g.TurnTallyFor(p.ID).CardsDrawn; n != 0 {
		t.Fatalf("draw tally %d after reset", n)
	}
}

// --- Kor, Cause of Vengeance ------------------------------------------------------

func TestTrekB40KorSacrificesToDrawAndPumpKlingonsAndWarriors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kor := b12Push(g, me.ID, "Kor, Cause of Vengeance", "Legendary Creature — Klingon Warrior Officer", trekB40KorOracle, 4, 4)
	warrior := b12Creature(g, me.ID, "Brave", "Creature — Human Warrior", 2, 2)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Klingon", "Creature — Klingon", 2, 2)
	hand := handSize(me)

	declareAttack(t, g, opp.ID, kor, warrior)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{kor}); err == nil {
		t.Fatal("sacrificed Kor itself")
	}
	answerOwnPermanents(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Fatal("a second sacrifice prompt: one attack declaration is one trigger")
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the sacrificed creature is still on the battlefield")
	}
	if n := handSize(me); n != hand+1 {
		t.Errorf("hand %d, want %d", n, hand+1)
	}
	if p, tt := trekB40PT(t, g, kor); p != 6 || tt != 6 {
		t.Errorf("Kor %d/%d, want 6/6", p, tt)
	}
	if p, tt := trekB40PT(t, g, warrior); p != 4 || tt != 4 {
		t.Errorf("the Warrior %d/%d, want 4/4", p, tt)
	}
	if p, tt := trekB40PT(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("an opponent's Klingon %d/%d, want 2/2", p, tt)
	}
}

func TestTrekB40KorDeclinedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kor := b12Push(g, me.ID, "Kor, Cause of Vengeance", "Legendary Creature — Klingon Warrior Officer", trekB40KorOracle, 4, 4)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	hand := handSize(me)
	declareAttack(t, g, opp.ID, kor)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) || handSize(me) != hand {
		t.Error("declining still sacrificed or drew")
	}
	if p, tt := trekB40PT(t, g, kor); p != 4 || tt != 4 {
		t.Errorf("Kor %d/%d after declining, want 4/4", p, tt)
	}
}

// --- Locutus of Borg -----------------------------------------------------------------

func TestTrekB40LocutusArtifactCreaturesMakeBorgOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	locutus := b12Push(g, me.ID, "Locutus of Borg", "Legendary Artifact Creature — Borg", trekB40LocutusOracle, 3, 3)
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Construct", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, locutus, "lifelink") {
		t.Error("Locutus has no lifelink")
	}
	life := me.Life
	trekB12HiveMindSwing(t, g, locutus, drone, bear)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if n := len(trekB40Tokens(g, me.ID, "Borg")); n != 2 {
		t.Errorf("%d Borg tokens, want 2 (Locutus and the Drone; the Bear has no such ability)", n)
	}
	for _, b := range trekB40Tokens(g, me.ID, "Borg") {
		if !b.IsArtifact() || !b.IsCreature() || b.Power != 1 || b.Toughness != 1 {
			t.Errorf("the Borg token is %s %d/%d", b.TypeLine, b.Power, b.Toughness)
		}
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d (lifelink)", me.Life, life+3)
	}
}

func TestTrekB40LocutusGivesNothingToAnOpponentsArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Locutus of Borg", "Legendary Artifact Creature — Borg", trekB40LocutusOracle, 3, 3)
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Construct", 1, 1)
	trekB12HiveMindSwing(t, g, drone)
	passPriorityAroundTable(t, g)
	if n := len(trekB40Tokens(g, me.ID, "Borg")); n != 0 {
		t.Errorf("%d Borg tokens from an opponent's Locutus", n)
	}
}

// --- Lt. Arex, Ever-Adapting ---------------------------------------------------------

func TestTrekB40ArexDrawsThreeAndPutsUpToThreeLandsTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	ids := trekB40Library(me,
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"},
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
	)
	mountain := handCardForTest(me, "Mountain", "Basic Land — Mountain", "")
	plains := handCardForTest(me, "Plains", "Basic Land — Plains", "")
	arex := castHoldCreature(t, g, "Lt. Arex, Ever-Adapting", "Legendary Creature — Edosian Officer", trekB40ArexOracle, 0, 0)
	if !hasEffectiveKeyword(t, g, arex, "trample") {
		t.Error("Lt. Arex has no trample")
	}
	for _, id := range ids[:3] {
		if !me.Hand.Contains(id) {
			t.Fatalf("the top three cards were not drawn")
		}
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land prompt")
	}
	for _, id := range pick.ChooseCards {
		if id == ids[1] {
			t.Fatal("a nonland card was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0], ids[2], mountain, plains}); err == nil {
		t.Fatal("put four lands")
	}
	answerChooseCards(t, g, me.ID, ids[0], ids[2], mountain)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{ids[0], ids[2], mountain} {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !g.Battlefield.Contains(id) || !c.Tapped {
			t.Errorf("%s did not enter tapped", c.Name)
		}
	}
	if !me.Hand.Contains(plains) {
		t.Error("the fourth land left the hand")
	}
	// Forest, Island, Swamp, Mountain: four names.
	if p, tt := trekB40PT(t, g, arex); p != 4 || tt != 4 {
		t.Errorf("Lt. Arex %d/%d, want 4/4", p, tt)
	}
}

// --- McCoy, Not a Bricklayer ---------------------------------------------------------

func TestTrekB40McCoyCountersOnEntryAndLifelinkWithCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mccoy := castHoldCreature(t, g, "McCoy, Not a Bricklayer", "Legendary Creature — Human Doctor", trekB40McCoyOracle, 0, 4)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, mccoy} {
		c, _ := g.LookupCardForEffect(id)
		if c.Counters[game.CounterPlusOne] != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", c.Name, c.Counters[game.CounterPlusOne])
		}
		if !hasEffectiveKeyword(t, g, id, "lifelink") {
			t.Errorf("%s has a counter and no lifelink", c.Name)
		}
	}
	plain := b12Creature(g, me.ID, "Plain", "Creature — Elf", 1, 1)
	if hasEffectiveKeyword(t, g, plain, "lifelink") {
		t.Error("a creature with no counter has lifelink")
	}

	// Another Doctor entering triggers it again; a non-Doctor does not.
	enterCard(t, g, me.ID, game.Card{Name: "Medic", TypeLine: "Creature — Human Doctor", Power: 1, Toughness: 1})
	passPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(bear); c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("after a Doctor entered the Bear has %d counters, want 2", c.Counters[game.CounterPlusOne])
	}
	enterCreature(t, g, me.ID, "Not a Doctor", 1, 1)
	passPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(bear); c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("a non-Doctor triggered it: the Bear has %d counters", c.Counters[game.CounterPlusOne])
	}
	// An opponent's Doctor is not "you control".
	enterCard(t, g, g.Seats[1].ID, game.Card{Name: "Their Medic", TypeLine: "Creature — Human Doctor", Power: 1, Toughness: 1})
	passPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(bear); c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("an opponent's Doctor triggered it: the Bear has %d counters", c.Counters[game.CounterPlusOne])
	}
}

// --- One, Borg Evolved -----------------------------------------------------------------

func TestTrekB40OneGrantsArtifactCreaturesItsKeywordsAndWard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	one := b12Push(g, me.ID, "One, Borg Evolved", "Legendary Artifact Creature — Borg Artificer", trekB40OneOracle, 3, 3)
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Construct", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Drone", "Artifact Creature — Construct", 1, 1)
	for _, kw := range []string{"flying", "double strike", "vigilance", "menace", "lifelink"} {
		for _, id := range []uuid.UUID{one, drone} {
			if !hasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("an artifact creature I control has no %s", kw)
			}
		}
		if hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("a non-artifact creature has %s", kw)
		}
		if hasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("an opponent's artifact creature has %s", kw)
		}
	}

	// Ward {2}: an opponent's spell at the Drone is countered unless paid.
	advanceToMain(t, g)
	blade := castAtWardedCreature(t, g, opp, drone)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(drone) {
		t.Error("the warded Drone died; the spell should have been countered")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
}

// --- Pavel Chekov, First Posting ---------------------------------------------------------

func TestTrekB40ChekovRevealsUntilTheChosenKindInsteadOfDrawing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	bfCardWithOracle(g, me, "Pavel Chekov, First Posting", "Legendary Creature — Human Pilot", trekB40ChekovOracle)
	ids := trekB40Library(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
	)
	_ = g.DrawCard(me.ID)
	answerOptional(t, g, me, true)
	answerOptionPick(t, g, me.ID, 0) // land
	if !me.Hand.Contains(ids[2]) {
		t.Fatal("the first land revealed is not in hand")
	}
	for _, id := range ids[:2] {
		if me.Hand.Contains(id) {
			t.Error("a nonland revealed card reached the hand")
		}
		if !me.Library.Contains(id) {
			t.Error("a revealed card left the library")
		}
	}
	// The two others went under; the unrevealed card is now on top.
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != ids[3] {
		t.Error("the unrevealed card is not on top")
	}
	if g.TurnTallyFor(me.ID).CardsDrawn != 0 {
		t.Error("the replaced draw counted as a draw")
	}
}

func TestTrekB40ChekovNonlandAndDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	bfCardWithOracle(g, me, "Pavel Chekov, First Posting", "Legendary Creature — Human Pilot", trekB40ChekovOracle)
	ids := trekB40Library(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"},
	)
	_ = g.DrawCard(me.ID)
	answerOptional(t, g, me, true)
	answerOptionPick(t, g, me.ID, 1) // nonland
	if !me.Hand.Contains(ids[2]) {
		t.Fatal("the nonland card is not in hand")
	}
	if me.Hand.Contains(ids[0]) || me.Hand.Contains(ids[1]) {
		t.Error("a land revealed card reached the hand")
	}
	// Declined: an ordinary draw of the top card.
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	_ = g.DrawCard(me.ID)
	answerOptional(t, g, me, false)
	if !me.Hand.Contains(top) {
		t.Error("declining did not draw the top card")
	}
	if latestOptionPickFor(g, me.ID) != nil {
		t.Error("declining still asked land or nonland")
	}
}

func TestTrekB40ChekovIgnoresAnOpponentsDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	bfCardWithOracle(g, me, "Pavel Chekov, First Posting", "Legendary Creature — Human Pilot", trekB40ChekovOracle)
	_ = g.DrawCard(other.ID)
	if optionalOffer(g, other) != nil || optionalOffer(g, me) != nil {
		t.Error("an opponent's draw was offered the replacement")
	}
}

// --- Reg Barclay, Fidgety Fantasist -------------------------------------------------------

func TestTrekB40BarclayCopiesOnTheSecondDrawAndExilesTheTokenAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	reg := b12Push(g, me.ID, "Reg Barclay, Fidgety Fantasist", "Legendary Creature — Human Artificer", trekB40BarclayOracle, 1, 5)
	hero := b12Creature(g, me.ID, "Hero", "Legendary Creature — Human Knight", 3, 3)
	advanceToMain(t, g)
	trekB40ResetDraws(t, g, me)
	pick := game.TargetRef{Kind: game.TargetCard, ID: hero}
	trekB40Draw(t, g, me, &pick)
	if n := len(trekB40Tokens(g, me.ID, "Hero")); n != 0 {
		t.Fatalf("%d copies after the first draw", n)
	}
	// Reg can't target itself.
	trekB40DrawExpectingTargets(t, g, me, reg, hero)
	copies := trekB40Tokens(g, me.ID, "Hero")
	if len(copies) != 1 {
		t.Fatalf("%d copies after the second draw, want 1", len(copies))
	}
	tok := copies[0]
	if strings.Contains(tok.TypeLine, "Legendary") || !tok.IsArtifact() || !tok.HasSubtype("Illusion") ||
		!tok.HasSubtype("Knight") || !tok.IsCreature() {
		t.Errorf("the copy's type line is %q", tok.TypeLine)
	}
	if !hasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
		t.Error("the copy has no haste")
	}
	if p, tt := trekB40PT(t, g, tok.InstanceID); p != 3 || tt != 3 {
		t.Errorf("the copy is %d/%d", p, tt)
	}
	trekB40Draw(t, g, me, nil)
	if n := len(trekB40Tokens(g, me.ID, "Hero")); n != 1 {
		t.Errorf("%d copies after the third draw, want it to stay 1", n)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tok.InstanceID) {
		t.Error("the token was not exiled at the beginning of the end step")
	}
	if !g.Battlefield.Contains(hero) {
		t.Error("the original left with its copy")
	}
}

// trekB40DrawExpectingTargets draws for p and answers the target prompt
// with `pick`, after checking `not` is refused.
func trekB40DrawExpectingTargets(t *testing.T, g *game.Game, p *game.Player, not, pick uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, 1) })
	pt := latestPickTarget(g, p.ID)
	if pt == nil {
		t.Fatal("no target prompt on the second draw")
	}
	if err := g.ResolvePickTarget(pt.ID, p.ID, game.TargetRef{Kind: game.TargetCard, ID: not}); err == nil {
		t.Fatal("Reg Barclay targeted itself")
	}
	ref := game.TargetRef{Kind: game.TargetCard, ID: pick}
	assistCardSettle(t, g, p.ID, &ref, true)
}

// --- Rite of Ascension ------------------------------------------------------------------

func TestTrekB40RiteMakesAnAttackingWarriorForEachOpponentAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Permanent(g, me.ID, "Rite of Ascension", "Enchantment")
	rite := findBattlefieldByName(g, "Rite of Ascension")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rite {
				g.Battlefield.Cards[i].OracleID = trekB40RiteOracle
			}
		}
	})
	x := b12Creature(g, me.ID, "X", "Creature — Bear", 2, 2)
	y := b12Creature(g, me.ID, "Y", "Creature — Bear", 2, 2)
	z := b12Creature(g, me.ID, "Z", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []struct{ who, at uuid.UUID }{{x, a.ID}, {y, b.ID}, {z, b.ID}} {
		if err := g.DeclareAttacker(d.who, d.at); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	for i := 0; i < 20 && !stackFullyEmpty(g); i++ {
		passAndSettleAttackSplit(t, g, me.ID)
	}
	warriors := trekB40Tokens(g, me.ID, "Klingon Warrior")
	if len(warriors) != 2 {
		t.Fatalf("%d Klingon Warriors, want 2 (two opponents attacked)", len(warriors))
	}
	for _, w := range warriors {
		if !w.Tapped || w.AttackingTarget == uuid.Nil {
			t.Error("a Warrior is not tapped and attacking")
		}
		if w.AttackingTarget == me.ID {
			t.Error("a Warrior attacks its controller")
		}
		if w.Power != 2 || w.Toughness != 1 || !hasEffectiveKeyword(t, g, w.InstanceID, "haste") {
			t.Errorf("the Warrior is %d/%d", w.Power, w.Toughness)
		}
	}
}

// --- Sins of the Father ---------------------------------------------------------------------

func TestTrekB40SinsOfTheFatherChapters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(other.ID, RedGoblinToken(), 1) })
	goblin := findBattlefieldByName(g, "Goblin")

	castCatalogSpell(t, g, "Sins of the Father", "Enchantment — Saga", trekB40SinsOracle, nil)
	passPriorityAroundTable(t, g)
	// I: each opponent sacrifices a nontoken creature.
	if g.Battlefield.Contains(theirs) {
		t.Error("chapter I: the opponent's nontoken creature survived")
	}
	if !g.Battlefield.Contains(goblin) || !g.Battlefield.Contains(mine) {
		t.Error("chapter I took a token or the controller's creature")
	}

	// II: a Klingon Warrior for each opponent.
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	warriors := trekB40Tokens(g, me.ID, "Klingon Warrior")
	if len(warriors) != len(g.Seats)-1 {
		t.Fatalf("chapter II: %d Klingon Warriors, want %d", len(warriors), len(g.Seats)-1)
	}

	// III: exile opp's graveyard; each opponent loses one per creature card.
	gyCardWithOracle(opp, "Dead Bear", "Creature — Bear", "")
	gyCardWithOracle(opp, "Dead Elf", "Creature — Elf", "")
	gyCardWithOracle(opp, "Old Bolt", "Instant", "")
	gyCards := opp.Graveyard.Size()
	creatures := 0
	for _, c := range opp.Graveyard.Cards {
		if c.IsCreature() {
			creatures++
		}
	}
	myLife := me.Life
	start := lifeOfOpponents(g)
	advanceToPrecombatMainOf(t, g, 0)
	pt := latestPickTarget(g, me.ID)
	if pt == nil {
		t.Fatal("chapter III: no target prompt")
	}
	if err := g.ResolvePickTarget(pt.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != 0 {
		t.Errorf("chapter III left %d of %d cards in the graveyard", opp.Graveyard.Size(), gyCards)
	}
	for j, life := range lifeOfOpponents(g) {
		if life != start[j]-creatures {
			t.Errorf("opponent %d at %d, want %d", j, life, start[j]-creatures)
		}
	}
	if me.Life != myLife {
		t.Error("the controller lost life")
	}
}

// --- Tom Paris, Risk-Taker --------------------------------------------------------------------

func TestTrekB40TomParisGrantsKeywordsOnTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	paris := b12Push(g, me.ID, "Tom Paris, Risk-Taker", "Legendary Creature — Human Pilot", trekB40ParisOracle, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	kws := []string{"flying", "double strike", "vigilance", "haste"}
	trekB40ResetDraws(t, g, me)
	trekB40Draw(t, g, opp, nil)
	trekB40Draw(t, g, opp, nil)
	trekB40Draw(t, g, me, nil)
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Fatal("granted on the first draw or on an opponent's draws")
	}
	trekB40Draw(t, g, me, nil)
	for _, kw := range kws {
		for _, id := range []uuid.UUID{paris, bear} {
			if !hasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("a creature I control has no %s after the second draw", kw)
			}
		}
		if hasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("an opponent's creature has %s", kw)
		}
	}
	late := b12Creature(g, me.ID, "Late", "Creature — Bear", 2, 2)
	if hasEffectiveKeyword(t, g, late, "flying") {
		t.Error("a creature that arrived after the effect began gained flying (CR 611.2c)")
	}
}
