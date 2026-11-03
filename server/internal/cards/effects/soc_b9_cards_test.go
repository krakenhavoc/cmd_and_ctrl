package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b9_cards_test.go — Secrets of Strixhaven Commander, batch 9
// (#4834). Each test plays the real catalog entry and asserts what its
// text says.

const (
	socB9MoonshakerOracle = "b3f16f34-d78c-4b61-821b-05fc6bd0da37"
	socB9MuddleOracle     = "9a8f6e5e-6cde-49df-861c-a26cb09c9b12"
	socB9MycolothOracle   = "d7fd16ce-282d-49cd-b2b4-0d25935e7a72"
	socB9NaktamunOracle   = "c78783e5-868d-4a8b-a4f8-95a92853cf0a"
	socB9NevOracle        = "42750e30-d022-4623-a290-d91fa80721c9"
	socB9NexusOracle      = "6bdbe36f-b46d-4746-baf2-04f0977bb532"
	socB9NilsOracle       = "ff6b52b7-ae3c-465b-b10e-901160e5230e"
	socB9HarvestOracle    = "76bc6518-dea4-4a5c-a8a7-352cd297197b"
	socB9OversimplifyOrc  = "25e1d5fe-c2d8-4a11-ad54-bf1626f9ab45"
	socB9OwlinOracle      = "36d6d6c0-721c-40c3-b81f-4f027fcf847e"
	socB9StrokeOracle     = "0cc6d683-366f-4ae4-be60-20ad9621fdaf"
)

func TestSocB9EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB9MoonshakerOracle:      "Moonshaker Cavalry",
		socB9MuddleOracle:          "Muddle, the Ever-Changing",
		socB9MycolothOracle:        "Mycoloth",
		socB9NaktamunOracle:        "Naktamun Lorespinner",
		socB9NaktamunOracle + "#1": "Wheel of Fortune",
		socB9NevOracle:             "Nev, the Practical Dean",
		socB9NexusOracle:           "Nexus Mentality",
		socB9NilsOracle:            "Nils, Discipline Enforcer",
		socB9HarvestOracle:         "Ominous Harvest",
		socB9OversimplifyOrc:       "Oversimplify",
		socB9OwlinOracle:           "Owlin Spiralmancer",
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

// socB9Seed puts a catalog creature on the battlefield.
func socB9Seed(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

// socB9CastX casts a spell from the active seat's hand with a mana cost
// and an X, leaving it on the stack.
func socB9CastX(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, x int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// socB9Counter puts n counters of kind on a permanent.
func socB9Counter(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddCounterForEffect(id, kind, n) })
	if err != nil {
		t.Fatalf("add %d %s counters: %v", n, kind, err)
	}
}

// socB9Library gives each seat ten cards to draw.
func socB9Library(g *game.Game) {
	for _, p := range g.Seats {
		for i := 0; i < 10; i++ {
			p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
		}
	}
}

// --- Moonshaker Cavalry ---------------------------------------------------

func TestSocB9MoonshakerCavalryGivesTheTeamFlyingAndPlusX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	cav := castHoldCreature(t, g, "Moonshaker Cavalry", "Creature — Spirit Knight", socB9MoonshakerOracle, 6, 6)
	passPriorityAroundTable(t, g)
	// X = 2: the Bear and the Cavalry itself.
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("the Bear is %d power, want 4 (+2/+2)", p)
	}
	if p := effectivePower(t, g, cav); p != 8 {
		t.Errorf("the Cavalry is %d power, want 8", p)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear did not gain flying")
	}
	if effectivePower(t, g, theirs) != 2 || hasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("an opponent's creature was pumped")
	}
	// A creature that arrives afterwards gets nothing (the ruling).
	late := b12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if effectivePower(t, g, late) != 2 || hasEffectiveKeyword(t, g, late, "flying") {
		t.Error("a creature that entered after the trigger resolved was pumped")
	}
}

// --- Mycoloth ---------------------------------------------------------------

func TestSocB9MycolothDevoursAndMakesSaprolingsPerCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushDevourFodder(g, me.ID, "Goblin A")
	b := pushDevourFodder(g, me.ID, "Goblin B")
	id := castDevourer(t, g, "Mycoloth", socB9MycolothOracle)
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b)
	if got := devourPlusOnes(g, id); got != 4 {
		t.Fatalf("counters = %d, want 4 (devour 2 x 2)", got)
	}
	// Another counter from elsewhere counts too (the ruling).
	socB9Counter(t, g, id, game.CounterPlusOne, 1)
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Saproling", me.ID); got != 5 {
		t.Errorf("%d Saprolings, want 5 (one per +1/+1 counter)", got)
	}
}

func TestSocB9MycolothMakesNothingOnAnotherPlayersUpkeepOrWithNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	socB9Seed(g, me.ID, "Mycoloth", "Creature — Fungus", socB9MycolothOracle, 4, 4)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Saproling", me.ID); got != 0 {
		t.Errorf("%d Saprolings, want 0", got)
	}
}

// --- Naktamun Lorespinner // Wheel of Fortune ------------------------------

func socB9Lorespinner(g *game.Game, owner uuid.UUID) game.Card {
	c := sos1Prep(owner, socB9NaktamunOracle, nil, "Naktamun Lorespinner", "Creature — Jackal Wizard", "{2}{R}", 3, 3,
		"Wheel of Fortune", "Sorcery", "{2}{R}")
	pushBattlefieldCardWithTimestamp(g, c)
	return c
}

func socB9FillHands(g *game.Game, n int) {
	for _, p := range g.Seats {
		for p.Hand.Size() < n {
			p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Held", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
		}
	}
}

func TestSocB9NaktamunPreparesOnlyWhenAPlayerIsNearlyEmptyHanded(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c := socB9Lorespinner(g, me.ID)
	advanceToEndStepOf(t, g, 3)
	socB9FillHands(g, 3)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("prepared while every player held three cards")
	}
	// An opponent down to one card: the next upkeep prepares it.
	advanceToEndStepOf(t, g, 3)
	socB9FillHands(g, 3)
	opp := g.Seats[2]
	for opp.Hand.Size() > 1 {
		_, _ = opp.Hand.Remove(opp.Hand.Cards[0].InstanceID)
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("not prepared with an opponent on one card")
	}
	// Wheel of Fortune: every hand is discarded, then seven each.
	socB9Library(g)
	for _, p := range g.Seats {
		for i := 0; i < 5; i++ {
			p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
		}
	}
	advanceToMain(t, g)
	fraBCastCopy(t, g, me, "Wheel of Fortune")
	for i, p := range g.Seats {
		if p.Hand.Size() != 7 {
			t.Errorf("seat %d holds %d cards, want 7", i, p.Hand.Size())
		}
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after the copy was cast")
	}
}

func TestSocB9NaktamunRechecksItsIfAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c := socB9Lorespinner(g, me.ID)
	advanceToEndStepOf(t, g, 3)
	socB9FillHands(g, 3)
	opp := g.Seats[1]
	for opp.Hand.Size() > 0 {
		_, _ = opp.Hand.Remove(opp.Hand.Cards[0].InstanceID)
	}
	advanceToUpkeepOf(t, g, 0)
	// The trigger waits; the opponent's hand refills in response.
	for opp.Hand.Size() < 3 {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Held", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	passPriorityAroundTable(t, g)
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("prepared although no player had one or fewer cards as it resolved (CR 603.4)")
	}
}

// --- Nev, the Practical Dean -------------------------------------------------

func TestSocB9NevGivesTrampleToCreaturesWithAnyCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	socB9Seed(g, me.ID, "Nev, the Practical Dean", "Legendary Creature — Merfolk Wizard", socB9NevOracle, 2, 2)
	oiled := b12Creature(g, me.ID, "Oiled", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	socB9Counter(t, g, oiled, "oil", 1)
	socB9Counter(t, g, theirs, game.CounterPlusOne, 1)
	if !hasEffectiveKeyword(t, g, oiled, "trample") {
		t.Error("a creature with an oil counter lacks trample")
	}
	if hasEffectiveKeyword(t, g, plain, "trample") {
		t.Error("a creature with no counter has trample")
	}
	if hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature has trample")
	}
	socB9Counter(t, g, oiled, "oil", -1)
	if hasEffectiveKeyword(t, g, oiled, "trample") {
		t.Error("trample stayed after the last counter came off")
	}
}

func TestSocB9NevGrowsByXOnTheFirstXSpellEachTurnOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nev := socB9Seed(g, me.ID, "Nev, the Practical Dean", "Legendary Creature — Merfolk Wizard", socB9NevOracle, 2, 2)
	advanceToMain(t, g)
	socB9CastX(t, g, "Plain Sorcery", "Sorcery", "{1}{G}", "", 0, nil)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, nev); got != 0 {
		t.Fatalf("a spell with no {X} put %d counters", got)
	}
	socB9CastX(t, g, "X Sorcery", "Sorcery", "{X}{G}", "", 3, nil)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, nev); got != 3 {
		t.Fatalf("Nev has %d counters, want 3", got)
	}
	if !hasEffectiveKeyword(t, g, nev, "trample") {
		t.Error("Nev with counters lacks trample")
	}
	socB9CastX(t, g, "Second X Sorcery", "Sorcery", "{X}{G}", "", 4, nil)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, nev); got != 3 {
		t.Errorf("the second X spell this turn grew Nev to %d", got)
	}
}

func TestSocB9NevIgnoresAnXSpellAfterOneCastBeforeItArrived(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	socB9CastX(t, g, "Early X Sorcery", "Sorcery", "{X}{G}", "", 1, nil)
	passPriorityAroundTable(t, g)
	nev := socB9Seed(g, me.ID, "Nev, the Practical Dean", "Legendary Creature — Merfolk Wizard", socB9NevOracle, 2, 2)
	socB9CastX(t, g, "Late X Sorcery", "Sorcery", "{X}{G}", "", 2, nil)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, nev); got != 0 {
		t.Errorf("Nev has %d counters, want 0 (the turn's first X spell came before it, the ruling)", got)
	}
}

// --- Owlin Spiralmancer ------------------------------------------------------

func TestSocB9OwlinCopiesTheFirstXSpellWithItsX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	socB9Library(g)
	socB9Seed(g, me.ID, "Owlin Spiralmancer", "Creature — Bird Wizard", socB9OwlinOracle, 3, 4)
	advanceToMain(t, g)
	hand := opp.Hand.Size()
	socB9CastX(t, g, "Stroke of Genius", "Instant", "{X}{2}{U}", socB9StrokeOracle, 2, b16TargetPlayer(opp.ID))
	answerLatestTriggerPrompt(t, g, me.ID, true)
	answerCopyPrompts(t, g, me.ID, opp.ID, 1)
	if got := opp.Hand.Size() - hand; got != 4 {
		t.Errorf("the opponent drew %d, want 4 (the spell and its copy, X = 2 each)", got)
	}
	// The second X spell this turn is not offered.
	socB9CastX(t, g, "Stroke of Genius", "Instant", "{X}{2}{U}", socB9StrokeOracle, 1, b16TargetPlayer(opp.ID))
	for i := 0; i < 8 && !stackFullyEmpty(g); i++ {
		if socB9HasConfirm(g, me.ID) {
			t.Fatal("the second X spell of the turn offered a copy")
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

func TestSocB9OwlinDeclinedMakesNoCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	socB9Library(g)
	socB9Seed(g, me.ID, "Owlin Spiralmancer", "Creature — Bird Wizard", socB9OwlinOracle, 3, 4)
	advanceToMain(t, g)
	hand := opp.Hand.Size()
	socB9CastX(t, g, "Stroke of Genius", "Instant", "{X}{2}{U}", socB9StrokeOracle, 2, b16TargetPlayer(opp.ID))
	answerLatestTriggerPrompt(t, g, me.ID, false)
	answerCopyPrompts(t, g, me.ID, opp.ID, 0)
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Errorf("the opponent drew %d, want 2 (no copy)", got)
	}
}

func socB9HasConfirm(g *game.Game, chooser uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			return true
		}
	}
	return false
}

// --- Ominous Harvest -----------------------------------------------------------

func TestSocB9OminousHarvestCopiesForEachPermanentPutIntoAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	socB9Library(g)
	advanceToMain(t, g)
	// Two creatures die (one of them an opponent's) and a token dies.
	killCreature(t, g, me.ID, b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2))
	killCreature(t, g, opp.ID, b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2))
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Token", TypeLine: "Token Creature — Bear", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	killCreature(t, g, me.ID, tok)
	// A bounced permanent did not go to a graveyard.
	bounced := b12Creature(g, me.ID, "Bounced", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.BounceCardsToHandForEffect([]uuid.UUID{bounced}) })

	hand, life := opp.Hand.Size(), opp.Life
	castCatalogSpell(t, g, "Ominous Harvest", "Sorcery", socB9HarvestOracle, b16TargetPlayer(opp.ID))
	answerCopyPrompts(t, g, me.ID, opp.ID, 3)
	if got := opp.Hand.Size() - hand; got != 4 {
		t.Errorf("the opponent drew %d, want 4 (the spell and three copies)", got)
	}
	if got := life - opp.Life; got != 4 {
		t.Errorf("the opponent lost %d life, want 4", got)
	}
}

func TestSocB9OminousHarvestWithNothingDeadIsOneSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	socB9Library(g)
	hand := opp.Hand.Size()
	castCatalogSpell(t, g, "Ominous Harvest", "Sorcery", socB9HarvestOracle, b16TargetPlayer(opp.ID))
	answerCopyPrompts(t, g, me.ID, opp.ID, 0)
	if got := opp.Hand.Size() - hand; got != 1 {
		t.Errorf("the opponent drew %d, want 1", got)
	}
}

// --- Oversimplify ------------------------------------------------------------

func TestSocB9OversimplifyExilesAllAndMakesFractalsOfTheExiledPower(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	three := b12Creature(g, me.ID, "Three", "Creature — Bear", 3, 3)
	socB9Counter(t, g, three, game.CounterPlusOne, 1)
	theirs := b12Creature(g, a.ID, "Theirs", "Creature — Bear", 4, 4)
	castCatalogSpell(t, g, "Oversimplify", "Sorcery", socB9OversimplifyOrc, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{two, three, theirs} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
	}
	check := func(p *game.Player, want int) {
		t.Helper()
		var fractals []uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Controller == p.ID && c.Name == "Fractal" {
				fractals = append(fractals, c.InstanceID)
			}
		}
		if want == 0 {
			if len(fractals) != 0 {
				t.Errorf("%s keeps a Fractal; a 0/0 dies", p.Name)
			}
			return
		}
		if len(fractals) != 1 {
			t.Fatalf("%s has %d Fractals, want 1", p.Name, len(fractals))
		}
		if got := plusOnes(g, fractals[0]); got != want {
			t.Errorf("%s's Fractal has %d counters, want %d", p.Name, got, want)
		}
	}
	check(me, 6) // 2 + (3 + 1): power as it last existed, counters included
	check(a, 4)
	check(b, 0)
}

// --- Nexus Mentality ---------------------------------------------------------

func socB9CastNexus(g *game.Game, modes []int, targets []game.TargetRef) error {
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Nexus Mentality", TypeLine: "Instant", OracleID: socB9NexusOracle, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
	if err != nil {
		_, _ = me.Hand.Remove(id)
	}
	return err
}

func TestSocB9NexusMentalityMovesEveryCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	from := b12Creature(g, me.ID, "From", "Creature — Bear", 2, 2)
	to := b12Creature(g, me.ID, "To", "Creature — Bear", 2, 2)
	socB9Counter(t, g, from, game.CounterPlusOne, 2)
	socB9Counter(t, g, from, "oil", 1)
	if err := socB9CastNexus(g, []int{0}, []game.TargetRef{
		{Kind: game.TargetCard, ID: from, Mode: 0, Slot: 0},
		{Kind: game.TargetCard, ID: to, Mode: 0, Slot: 1},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, from) != 0 || countersOn(g, from, "oil") != 0 {
		t.Error("counters were left on the first permanent")
	}
	if plusOnes(g, to) != 2 || countersOn(g, to, "oil") != 1 {
		t.Errorf("the second permanent has %d +1/+1 and %d oil, want 2 and 1", plusOnes(g, to), countersOn(g, to, "oil"))
	}
}

func TestSocB9NexusMentalityRefusesAnOpponentsPermanentAndTheSameOneTwice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	if err := socB9CastNexus(g, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: theirs, Mode: 0, Slot: 0}}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("an opponent's permanent: err = %v, want ErrIllegalTarget", err)
	}
	if err := socB9CastNexus(g, []int{0}, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Mode: 0, Slot: 0},
		{Kind: game.TargetCard, ID: mine, Mode: 0, Slot: 1},
	}); err == nil {
		t.Error("the same permanent was accepted as both targets")
	}
}

func TestSocB9NexusMentalityChoosesBothOnlyWithACommander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	socB9Library(g)
	advanceToMain(t, g)
	from := b12Creature(g, me.ID, "From", "Creature — Bear", 2, 2)
	to := b12Creature(g, me.ID, "To", "Creature — Bear", 2, 2)
	socB9Counter(t, g, from, game.CounterPlusOne, 2)
	socB9Counter(t, g, to, "oil", 1)
	targets := []game.TargetRef{
		{Kind: game.TargetCard, ID: from, Mode: 0, Slot: 0},
		{Kind: game.TargetCard, ID: to, Mode: 0, Slot: 1},
		{Kind: game.TargetCard, ID: to, Mode: 1, Slot: 0},
	}
	if err := socB9CastNexus(g, []int{0, 1}, targets); err == nil {
		t.Fatal("both modes were allowed with no commander")
	}
	b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
	hand := me.Hand.Size()
	if err := socB9CastNexus(g, []int{0, 1}, targets); err != nil {
		t.Fatalf("both modes with a commander: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The move lands first (printed order), then all three counters on
	// the second permanent come off and draw three.
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
	if plusOnes(g, to) != 0 || countersOn(g, to, "oil") != 0 || plusOnes(g, from) != 0 {
		t.Error("counters remain")
	}
}

// --- Nils, Discipline Enforcer ----------------------------------------------

func TestSocB9NilsPutsACounterOnUpToOneCreaturePerPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	socB9Seed(g, me.ID, "Nils, Discipline Enforcer", "Legendary Creature — Human Cleric", socB9NilsOracle, 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, a.ID, "Theirs", "Creature — Bear", 2, 2)
	other := b12Creature(g, b.ID, "Other", "Creature — Bear", 2, 2)
	advanceToEndStepOf(t, g, 0)
	answerPickTarget(t, g, mine)
	answerPickTarget(t, g, theirs)
	// No pick for the third player: every clause is "up to one".
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("decline the third clause: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, mine) != 1 || plusOnes(g, theirs) != 1 {
		t.Errorf("counters: mine %d, theirs %d, want 1 each", plusOnes(g, mine), plusOnes(g, theirs))
	}
	if plusOnes(g, other) != 0 {
		t.Error("the declined clause still put a counter")
	}
}

func TestSocB9NilsTaxesAttackersByTheirCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	socB9Seed(g, opp.ID, "Nils, Discipline Enforcer", "Legendary Creature — Human Cleric", socB9NilsOracle, 2, 2)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	marked := pushVanillaCreature(g, me.ID, "Marked", 2, 2)
	socB9Counter(t, g, marked, game.CounterPlusOne, 2)
	socB9Counter(t, g, marked, "stun", 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	walker := pushCatalogPermanent(g, opp.ID, "Test Walker", "Legendary Planeswalker — Test", "test-walker-tax", false)
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: plain, Target: opp.ID}}); !p.IsFree() {
		t.Errorf("a creature with no counters was charged %q", p.Cost)
	}
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: marked, Target: opp.ID}}); p.Total.Generic != 3 {
		t.Errorf("a creature with three counters (of two kinds) was charged %d, want 3", p.Total.Generic)
	}
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: marked, Target: walker}}); p.Total.Generic != 3 {
		t.Errorf("attacking Nils's planeswalker cost %d, want 3", p.Total.Generic)
	}
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: marked, Target: g.Seats[2].ID}}); !p.IsFree() {
		t.Errorf("attacking another player was charged %q", p.Cost)
	}
	if err := g.DeclareAttacker(marked, opp.ID); !errors.Is(err, game.ErrAttackTaxUnpaid) {
		t.Fatalf("attack with no mana: %v, want ErrAttackTaxUnpaid", err)
	}
	fillPool(me, 3)
	if err := g.DeclareAttacker(marked, opp.ID); err != nil {
		t.Fatalf("attack paying {3}: %v", err)
	}
}

// --- Muddle, the Ever-Changing ----------------------------------------------

func TestSocB9MuddleBecomesACopyWithMyriadUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	muddle := socB9Seed(g, me.ID, "Muddle, the Ever-Changing", "Legendary Creature — Elemental Otter Shapeshifter", socB9MuddleOracle, 3, 3)
	// A real catalog creature: a copy keeps its granted myriad only over an
	// oracle ID, which every real card has.
	ogre := socB9Seed(g, me.ID, "Gyre Sage", "Creature — Elf Druid", "e3aa16cf-079b-4737-9ffd-7bfdffef0cb2", 1, 2)
	legend := b12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 1, 1)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range pick.PickTargetCards {
		if id == legend || id == muddle {
			t.Errorf("a legendary creature (%s) was offered", id)
		}
	}
	pickCard(t, g, me.ID, ogre)
	passPriorityAroundTable(t, g)
	c := dcCard(t, g, muddle)
	if !c.IsCopy() || c.Name != "Gyre Sage" || c.Effective().Power != 1 {
		t.Fatalf("Muddle is %q %d/%d (copy %v), want a 1/2 Gyre Sage copy", c.Name, c.Effective().Power, c.Effective().Toughness, c.IsCopy())
	}
	// It has myriad: attacking one opponent offers copies at the others.
	declareAttack(t, g, opp.ID, muddle)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	answerMayChoice(t, g, me.ID, true)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Gyre Sage" && c.InstanceID != muddle && c.InstanceID != ogre && IsToken(c) && c.AttackingTarget != opp.ID {
			copies++
		}
	}
	if copies != 2 {
		t.Errorf("%d myriad token copies, want one at each other opponent", copies)
	}
	dcCleanup(g)
	if c := dcCard(t, g, muddle); c.IsCopy() || c.Name != "Muddle, the Ever-Changing" {
		t.Errorf("after cleanup Muddle is %q (copy %v)", c.Name, c.IsCopy())
	}
}

func TestSocB9MuddleWithNoTargetStaysItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	muddle := socB9Seed(g, me.ID, "Muddle, the Ever-Changing", "Legendary Creature — Elemental Otter Shapeshifter", socB9MuddleOracle, 3, 3)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if c := dcCard(t, g, muddle); c.IsCopy() {
		t.Error("Muddle became a copy of nothing")
	}
	// A creature spell does not trigger it.
	b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 5, 4)
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a creature spell triggered Muddle")
	}
}
