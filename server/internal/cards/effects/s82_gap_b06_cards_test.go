package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b06_cards_test.go — S82 GAP batch 6 (Mystery Booster Commander
// planeswalkers, #5948).

const (
	gapB06DyfedOracle      = "ab23a9e4-2d62-4b01-9d52-593de40811b0"
	gapB06AshayaOracle     = "0339d25b-9591-49c0-bf31-4ba749ebd0b0"
	gapB06WorzelOracle     = "15084803-8470-469b-8cd9-154c2341b96d"
	gapB06GreensleevesOrac = "20b90808-0279-46eb-ae51-6990ac679a83"
	gapB06FerozOracle      = "724bcd6b-4deb-49c0-aa0e-1245bb40dd56"
	gapB06SifaOracle       = "9a3d148c-e1bb-4e7a-a588-00a24041ebf7"
	gapB06UnluckiestOracle = "c7ea2ce7-f684-44a2-94b2-d4aa7353f6fc"
	gapB06VenserOracle     = "f4a64308-5f24-47c0-83f9-ee16a3b4f143"
)

func gapB06Card(g *game.Game, id uuid.UUID) game.Card {
	var c game.Card
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if p := findBattlefieldCardForTest(g, id); p != nil {
			c = *p
		}
	})
	return c
}

func gapB06Hand(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true}})
	return id
}

func gapB06Named(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func TestGapB06AllAreFull(t *testing.T) {
	for _, o := range []string{gapB06DyfedOracle, gapB06AshayaOracle, gapB06WorzelOracle, gapB06GreensleevesOrac,
		gapB06FerozOracle, gapB06SifaOracle, gapB06UnluckiestOracle, gapB06VenserOracle} {
		spec, ok := Lookup(o)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, completeness %v, caveats %v", o, ok, spec.Completeness, spec.Caveats)
		}
	}
}

// Dyfed: +1 two tapped Powerstones, -X untaps X target artifacts, -6 fetches one.
func TestGapB06DyfedPlusOneMakesTwoTappedPowerstones(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dyfed := pushWalkerForTest(g, me.ID, "Dyfed, the Guiding Hand", gapB06DyfedOracle, 4)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, dyfed, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	stones := gapB06Named(g, "Powerstone")
	if len(stones) != 2 {
		t.Fatalf("%d Powerstones, want 2", len(stones))
	}
	for _, s := range stones {
		if !s.Tapped {
			t.Error("a Powerstone entered untapped")
		}
	}
	if got := loyaltyOf(g, dyfed); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestGapB06DyfedMinusXUntapsXArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dyfed := pushWalkerForTest(g, me.ID, "Dyfed, the Guiding Hand", gapB06DyfedOracle, 4)
	a := b12Permanent(g, me.ID, "Rock A", "Artifact")
	b := b12Permanent(g, me.ID, "Rock B", "Artifact")
	c := b12Permanent(g, me.ID, "Rock C", "Artifact")
	creature := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{a, b, c, creature} {
			findBattlefieldCardForTest(g, id).Tapped = true
		}
	})
	advanceToMain(t, g)
	ref := func(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }
	// X = 2 needs exactly two targets, and a non-artifact is not one.
	if err := g.ActivateCatalogAbility(me.ID, dyfed, 1, game.ActivateAbilityParams{XValue: 2, Targets: []game.TargetRef{ref(a)}}); err == nil {
		t.Fatal("X = 2 with one target")
	}
	if err := g.ActivateCatalogAbility(me.ID, dyfed, 1, game.ActivateAbilityParams{XValue: 2, Targets: []game.TargetRef{ref(a), ref(creature)}}); err == nil {
		t.Fatal("a creature was a legal artifact target")
	}
	b16Activate(t, g, me.ID, dyfed, 1, game.ActivateAbilityParams{XValue: 2, Targets: []game.TargetRef{ref(a), ref(b)}})
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, a) || b16Tapped(t, g, b) {
		t.Error("a targeted artifact is still tapped")
	}
	if !b16Tapped(t, g, c) || !b16Tapped(t, g, creature) {
		t.Error("an untargeted permanent untapped")
	}
	if got := loyaltyOf(g, dyfed); got != 2 {
		t.Errorf("loyalty %d, want 4 - 2 = 2", got)
	}
}

func TestGapB06DyfedMinusSixFetchesAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dyfed := pushWalkerForTest(g, me.ID, "Dyfed, the Guiding Hand", gapB06DyfedOracle, 6)
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Sol Ring", "Artifact"),
		searchTestLand("Mind Stone", "Artifact"),
		searchTestLand("Bear", "Creature — Bear"),
	)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, dyfed, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	prompt := searchChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, prompt, "Forest") != uuid.Nil || searchOptionNamed(g, prompt, "Bear") != uuid.Nil {
		t.Error("a nonartifact card is offered")
	}
	answerSearchNamed(t, g, me.ID, "Sol Ring")
	if len(gapB06Named(g, "Sol Ring")) != 1 {
		t.Error("the artifact did not reach the battlefield")
	}
}

// Ashaya's Enduring Bond: only with a legendary creature or planeswalker, and
// it discovers for the mana spent.
func TestGapB06AshayaDiscoversForTheManaSpent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	spell := handCardOf(g, me, "Ashaya's Enduring Bond", "Legendary Sorcery", "{1}{R}{G}", gapB06AshayaOracle)
	apaMana(me, "R", "G", "C")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no legendary creature or planeswalker")
	}
	pushWalkerForTest(g, me.ID, "Some Walker", "", 3)
	me.Library.Cards = nil
	hit := discoverLibraryCard(me.ID, "Cheap Spell", "Instant", "{1}{G}")
	pushLibraryCardForTest(me, hit)
	pushLibraryCardForTest(me, discoverLibraryCard(me.ID, "Too Big", "Creature — Giant", "{4}{G}"))
	apaMana(me, "R", "G", "C")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	perm := g.CastPermissionOnCardByIDForEffect(hit.InstanceID)
	if perm == nil || perm.Discover == nil || perm.Discover.N != 3 {
		t.Fatalf("the discovered card's permission = %+v, want discover 3", perm)
	}
}

// Worzel: the Cat adds a loyalty counter to every planeswalker, -1 digs, -8 makes
// ten Sprites.
func TestGapB06WorzelCatAddsLoyaltyToEachWalker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	worzel := pushWalkerForTest(g, me.ID, "Worzel, the Protector", gapB06WorzelOracle, 4)
	other := pushWalkerForTest(g, me.ID, "Other Walker", "", 3)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, worzel, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	cats := gapB06Named(g, "Cat")
	if len(cats) != 1 || cats[0].Power != 1 || cats[0].Toughness != 1 {
		t.Fatalf("cats %+v, want one 1/1", cats)
	}
	cat := cats[0].InstanceID
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, cat).SummonedThisTurn = false })
	b16Activate(t, g, me.ID, cat, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if loyaltyOf(g, worzel) != 5 || loyaltyOf(g, other) != 4 {
		t.Errorf("loyalty Worzel %d other %d, want 5 and 4", loyaltyOf(g, worzel), loyaltyOf(g, other))
	}
	if !b16Tapped(t, g, cat) {
		t.Error("the Cat did not tap")
	}
}

func TestGapB06WorzelMinusOneTakesAWalkerOrBasicPlains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	worzel := pushWalkerForTest(g, me.ID, "Worzel, the Protector", gapB06WorzelOracle, 4)
	me.Library.Cards = nil
	// pushLibraryCardForTest puts cards on the bottom: the first pushed is the top.
	walker := pushLibraryCardForTest(me, game.Card{Name: "Library Walker", TypeLine: "Legendary Planeswalker — Test", Owner: me.ID, Controller: me.ID})
	plains := pushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID})
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	for i := 0; i < 8; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, worzel, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt")
	}
	if !hasID(pick.ChooseCards, plains) || !hasID(pick.ChooseCards, walker) || hasID(pick.ChooseCards, forest) {
		t.Errorf("offered %v: want the Plains and the planeswalker, not the Forest", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{plains}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(plains) {
		t.Error("the Plains is not in hand")
	}
}

func TestGapB06WorzelMinusEightMakesTenScrybSprites(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	worzel := pushWalkerForTest(g, me.ID, "Worzel, the Protector", gapB06WorzelOracle, 8)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, worzel, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	sprites := gapB06Named(g, "Scryb Sprites")
	if len(sprites) != 10 {
		t.Fatalf("%d Scryb Sprites, want 10", len(sprites))
	}
	if s := gapB06Card(g, sprites[0].InstanceID); !s.IsToken() || s.Power != 1 || s.Toughness != 1 || !game.HasKeyword(&s, "flying") || len(s.Colors) != 1 || s.Colors[0] != "G" {
		t.Errorf("Scryb Sprites is %+v, want a 1/1 green flying token", s)
	}
}

// Greensleeves.
func TestGapB06GreensleevesMakesAGiantBadgerThatGrowsOnBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushWalkerForTest(g, me.ID, "Greensleeves", gapB06GreensleevesOrac, 5)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	badgers := gapB06Named(g, "Giant Badger")
	if len(badgers) != 1 || badgers[0].Power != 2 || badgers[0].Toughness != 2 || !badgers[0].IsToken() {
		t.Fatalf("badgers %+v, want one 2/2 token", badgers)
	}
	// The Badger blocks on the opponent's turn and grows.
	badger := badgers[0].InstanceID
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, badger).SummonedThisTurn = false })
	advanceToNextSeatsTurn(t, g)
	attacker := b12Creature(g, opp.ID, "Attacker", "Creature — Giant", 3, 3)
	pr7aBlockedCombat(t, g, me.ID, attacker, badger)
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if p, th := gapB06PT(g, badger); p != 4 || th != 4 {
		t.Errorf("the blocking Badger is %d/%d, want 4/4", p, th)
	}
}

func gapB06PT(g *game.Game, id uuid.UUID) (int, int) {
	c := gapB06Card(g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

func TestGapB06GreensleevesMinusThreeMillsAndTakesPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := pushWalkerForTest(g, me.ID, "Greensleeves", gapB06GreensleevesOrac, 5)
	me.Library.Cards = nil
	// The first pushed is the top of the library: the three milled cards.
	spell := pushLibraryCardForTest(me, game.Card{Name: "Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	land := pushLibraryCardForTest(me, game.Card{Name: "Land", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	bear := pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	for i := 0; i < 3; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Deep", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) || !me.Hand.Contains(bear) {
		t.Error("a milled permanent card is not in hand")
	}
	if me.Hand.Contains(spell) || !me.Graveyard.Contains(spell) {
		t.Error("the milled instant should stay in the graveyard")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2", got)
	}
}

func TestGapB06GreensleevesMinusEightMakesCreaturesEightEight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushWalkerForTest(g, me.ID, "Greensleeves", gapB06GreensleevesOrac, 8)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p, th := gapB06PT(g, mine); p != 8 || th != 8 {
		t.Errorf("my Bear is %d/%d, want 8/8", p, th)
	}
	c := gapB06Card(g, mine)
	if !game.HasKeyword(&c, "trample") {
		t.Error("my Bear has no trample")
	}
	if p, th := gapB06PT(g, theirs); p != 2 || th != 2 {
		t.Errorf("their Bear is %d/%d, want 2/2", p, th)
	}
}

// Feroz.
func TestGapB06FerozBirdsDrawAndBan(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	feroz := pushWalkerForTest(g, me.ID, "Feroz, Ulgrotha's Warden", gapB06FerozOracle, 5)
	seedLibrary(me, "A", "B", "C")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, feroz, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	birds := gapB06Named(g, "Bird")
	if len(birds) != 2 || birds[0].Power != 1 || !game.HasKeyword(&birds[0], "flying") {
		t.Fatalf("birds %+v, want two 1/1 flyers", birds)
	}
	if got := loyaltyOf(g, feroz); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestGapB06FerozZeroDrawsAndPutsAPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	feroz := pushWalkerForTest(g, me.ID, "Feroz, Ulgrotha's Warden", gapB06FerozOracle, 5)
	me.Library.Cards = nil
	pushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Instant", ManaCost: "{U}", Owner: me.ID, Controller: me.ID})
	small := gapB06Hand(me, "Small Artifact", "Artifact", "{4}")
	big := gapB06Hand(me, "Big Artifact", "Artifact", "{5}")
	spell := gapB06Hand(me, "A Spell", "Instant", "{1}")
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, feroz, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d → %d, want the draw", hand, got)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no put-from-hand prompt")
	}
	if !hasID(pick.ChooseCards, small) || hasID(pick.ChooseCards, big) || hasID(pick.ChooseCards, spell) {
		t.Errorf("offered %v: want only the mana value 4 permanent", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{small}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the artifact did not enter")
	}
}

func TestGapB06FerozMinusSevenMakesFerozsBanThatTaxesCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	feroz := pushWalkerForTest(g, me.ID, "Feroz, Ulgrotha's Warden", gapB06FerozOracle, 7)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, feroz, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if len(gapB06Named(g, "Feroz's Ban")) != 1 {
		t.Fatal("no Feroz's Ban token")
	}
	bear := handSpell(me, "Bear", "Creature — Bear", "{1}{G}")
	apaMana(me, "G", "G")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a creature spell cost {1}{G} with Feroz's Ban out")
	}
	apaMana(me, "G", "G")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("with four mana: %v", err)
	}
	bolt := handSpell(me, "Opt", "Instant", "{U}")
	apaMana(me, "U")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a noncreature spell is taxed: %v", err)
	}
}

// Sifa.
func TestGapB06SifaDrainsOnAnyCreatureDeathAndGoads(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sifa := pushWalkerForTest(g, me.ID, "Sifa Grent", gapB06SifaOracle, 4)
	victim := b12Creature(g, opp.ID, "Doomed", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	life, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(victim) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 || opp.Life != oppLife-1 {
		t.Errorf("life me %d → %d, opp %d → %d, want +1 and -1", life, me.Life, oppLife, opp.Life)
	}
	// +1 goads up to two.
	a := b12Creature(g, opp.ID, "Goad A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Goad B", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, sifa, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if b33Goaded(t, g, id) != me.ID {
			t.Errorf("%s is not goaded by Sifa's controller", id)
		}
	}
}

func TestGapB06SifaMinusSixExilesTwoFromEachOpponentToPlayFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sifa := pushWalkerForTest(g, me.ID, "Sifa Grent", gapB06SifaOracle, 6)
	for _, p := range g.Seats[1:] {
		p.Library.Cards = nil
		for i := 0; i < 4; i++ {
			pushLibraryCardForTest(p, game.Card{Name: "Their Card", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Owner: p.ID, Controller: p.ID})
		}
	}
	advanceToMain(t, g)
	sizes := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		sizes[p.ID] = p.Library.Size()
	}
	b16Activate(t, g, me.ID, sifa, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if got := sizes[p.ID] - p.Library.Size(); got != 2 {
			t.Errorf("%s exiled %d cards, want 2", p.Name, got)
		}
	}
	exiledCards := 0
	for _, c := range g.Exile.Cards {
		if perm := g.CastPermissionOnCardByIDForEffect(c.InstanceID); perm != nil && perm.Player == me.ID {
			exiledCards++
		}
	}
	if exiledCards != 2*len(g.Seats[1:]) {
		t.Errorf("%d exiled cards playable by Sifa's controller, want %d", exiledCards, 2*len(g.Seats[1:]))
	}
}

// Venser.
func TestGapB06VenserCountersOnCreaturesNotCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushWalkerForTest(g, me.ID, "Venser, Visionary Traveler", gapB06VenserOracle, 4)
	advanceToMain(t, g)
	// Cast from hand: no counters.
	cast := handCardOf(g, me, "Cast Bear", "Creature — Bear", "{1}{G}", "")
	apaMana(me, "G", "G")
	if err := g.CastSpell(me.ID, cast, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, cast); !ok || c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("a creature cast from hand has %d counters, want 0", c.Counters[game.CounterPlusOne])
	}
	// Put from hand: two counters.
	put := handCardOf(g, me, "Put Bear", "Creature — Bear", "{1}{G}", "")
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(put, game.HandEntryOptions{Controller: me.ID}); err != nil {
			t.Fatal(err)
		}
	})
	if c, ok := battlefieldCard(g, put); !ok || c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("a creature put onto the battlefield has %d counters, want 2", c.Counters[game.CounterPlusOne])
	}
	// A token gets none.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 black Zombie"), 1)
	})
	if z := findBattlefieldByName(g, "Zombie"); z != uuid.Nil {
		if c, _ := battlefieldCard(g, z); c.Counters[game.CounterPlusOne] != 0 {
			t.Errorf("a token has %d counters, want 0", c.Counters[game.CounterPlusOne])
		}
	}
}

func TestGapB06VenserPlusOneBlinksUntilTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	venser := pushWalkerForTest(g, me.ID, "Venser, Visionary Traveler", gapB06VenserOracle, 4)
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, venser, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: venser}}}); err == nil {
		t.Fatal("Venser targeted himself")
	}
	b16Activate(t, g, me.ID, venser, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ally}}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ally) || !g.Exile.Contains(ally) {
		t.Fatal("the Ally was not exiled")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := len(gapB06Named(g, "Ally")); n != 1 {
		t.Errorf("%d Allies on the battlefield after the end step, want 1", n)
	}
}

func TestGapB06VenserMinusTwoBouncesOnePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	venser := pushWalkerForTest(g, me.ID, "Venser, Visionary Traveler", gapB06VenserOracle, 4)
	a := b12Permanent(g, o1.ID, "First Thing", "Artifact")
	b := b12Permanent(g, o2.ID, "Second Thing", "Artifact")
	land := b12Permanent(g, o1.ID, "Their Land", "Land")
	advanceToMain(t, g)
	ref := func(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }
	if err := g.ActivateCatalogAbility(me.ID, venser, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{ref(land)}}); err == nil {
		t.Fatal("a land was a legal target")
	}
	b16Activate(t, g, me.ID, venser, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{ref(a), ref(b)}})
	passPriorityAroundTable(t, g)
	if !o1.Hand.Contains(a) || !o2.Hand.Contains(b) {
		t.Error("a targeted permanent was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land left")
	}
}

// The Unluckiest Planeswalker.
func TestGapB06UnluckiestPlusOneMakesAGremlinForAnAuraSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := pushWalkerForTest(g, me.ID, "The Unluckiest Planeswalker", gapB06UnluckiestOracle, 4)
	other := pushWalkerForTest(g, me.ID, "Another Unlucky One", gapB06UnluckiestOracle, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	pickAny := func() {
		pick := pendingOfKind(g, game.PendingChoiceMana)
		if pick == nil {
			t.Fatal("no mana colour prompt")
		}
		if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
			t.Fatal(err)
		}
	}
	b16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	pickAny()
	aura := handSpell(me, "Pacifism", "Enchantment — Aura", "{R}")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("cast the Aura with the mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(gapB06Named(g, "Gremlin")); n != 1 {
		t.Fatalf("%d Gremlins after spending the mana on an Aura, want 1", n)
	}
	// Mana spent on another kind of spell makes none.
	b16Activate(t, g, me.ID, other, 0, game.ActivateAbilityParams{})
	pickAny()
	instant := handSpell(me, "Opt", "Instant", "{R}")
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(gapB06Named(g, "Gremlin")); n != 1 {
		t.Errorf("%d Gremlins after spending the mana on an instant, want still 1", n)
	}
}

func TestGapB06UnluckiestMinusThreeDiscardsDrawsAndMovesCurses(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	walker := pushWalkerForTest(g, me.ID, "The Unluckiest Planeswalker", gapB06UnluckiestOracle, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mineOnBear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Aura", TypeLine: "Enchantment — Aura",
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear}})
	curse := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Curse of Testing", TypeLine: "Enchantment — Aura Curse",
		Owner: o1.ID, Controller: o1.ID, AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}})
	_ = mineOnBear
	me.Hand.Cards = nil
	gapB06Hand(me, "H1", "Instant", "{U}")
	gapB06Hand(me, "H2", "Instant", "{U}")
	gapB06Hand(me, "H3", "Instant", "{U}")
	seedLibrary(me, "A", "B", "C", "D", "E")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	// Two Auras are controlled (My Aura, and the Curse is the opponent's): one, so draw two.
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand %d, want 2 (three discarded, twice one Aura drawn)", got)
	}
	answerChoosePlayer(t, g, me.ID, o2)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, curse)
	if !ok || c.AttachedTo.Kind != game.TargetPlayer || c.AttachedTo.ID != o2.ID {
		t.Errorf("the Curse is attached to %+v, want the chosen opponent %s", c.AttachedTo, o2.ID)
	}
	if c, _ := battlefieldCard(g, mineOnBear); c.AttachedTo.ID != bear {
		t.Error("an Aura on a creature moved")
	}
}
