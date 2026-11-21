package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB88WormOracle      = "8fb39f48-10a1-4d9c-870a-eeda070bf12b"
	gapB88RyanOracle      = "6c06e2d5-1473-4ce6-bcf6-10dfbbe5f801"
	gapB88SoothsayerOracl = "70718e4a-c3c2-4fcd-96f8-98cd7fee257c"
	gapB88GeneralOracle   = "76a416fa-d97b-4c24-af41-b2382f1fd546"
	gapB88YasminOracle    = "93637599-d9b4-4f7e-a5c0-5a92656e941e"
	gapB88LureFishOracle  = "51e0a61d-69f0-475d-b40b-50d69252800d"
	gapB88JennyOracle     = "d57adf99-972a-43fe-9ba2-2b01e673252f"
	gapB88TenthOracle     = "23af0a0a-1d12-47c7-b191-2bd3f84eea93"
)

func TestGapB88EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB88WormOracle:      "Memory Worm",
		gapB88RyanOracle:      "Ryan Sinclair",
		gapB88SoothsayerOracl: "Sibylline Soothsayer",
		gapB88GeneralOracle:   "Sontaran General",
		gapB88YasminOracle:    "Yasmin Khan",
		gapB88LureFishOracle:  "Frost Fair Lure Fish",
		gapB88JennyOracle:     "Jenny Flint",
		gapB88TenthOracle:     "The Tenth Doctor",
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

// gapB88CastFromExile casts a spell from exile with a permission for p.
func gapB88CastFromExile(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) {
	t.Helper()
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.Cards = append(g.Exile.Cards, game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		_ = g.ExileCardWithPermissionForEffect(id, game.CastPermission{Player: p.ID, Duration: g.UntilEndOfTurnDuration()})
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast %s from exile: %v", name, err)
	}
}

func gapB88InExile(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

func TestGapB88MemoryWormBurnsLootsAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	worm := apaPush(g, me.ID, me.ID, game.Card{Name: "Memory Worm", OracleID: gapB88WormOracle,
		TypeLine: "Creature — Alien Worm", Power: 1, Toughness: 1})
	fodder := handCardOf(g, opp, "Fodder", "Instant", "{R}", "")
	life, hand := opp.Life, opp.Hand.Size()
	gapB88CastFromExile(t, g, me, "Exile Spell", "Instant")
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("target player: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2 damage", life, opp.Life)
	}
	answerDiscard(t, g, opp.ID, fodder)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(fodder) {
		t.Error("the chosen card was not discarded")
	}
	if opp.Hand.Size() != hand {
		t.Errorf("hand %d, want %d (a discard, then a draw)", opp.Hand.Size(), hand)
	}
	if n := countersOn(g, worm, game.CounterPlusOne); n != 1 {
		t.Errorf("the Worm has %d counters, want 1", n)
	}
	// A hand cast does nothing.
	castCatalogSpell(t, g, "Hand Spell", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || countersOn(g, worm, game.CounterPlusOne) != 1 {
		t.Error("a cast from the hand triggered Paradox")
	}
}

func TestGapB88JennyFlintGrowsAnotherCreatureWhenYouSacrificeAClueOrFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jenny := apaPush(g, me.ID, me.ID, game.Card{Name: "Jenny Flint", OracleID: gapB88JennyOracle,
		TypeLine: "Legendary Creature — Human Detective", Power: 2, Toughness: 2, Keywords: []string{"first strike"}})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	clue := b12Permanent(g, me.ID, "Clue", "Token Artifact — Clue")
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(clue) })
	b04WaitForPickOrSkip(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("the other creature has %d counters, want 1", n)
	}
	if n := countersOn(g, jenny, game.CounterPlusOne); n != 0 {
		t.Errorf("Jenny put a counter on herself: %d", n)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rock) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || countersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Error("sacrificing a plain artifact triggered Jenny")
	}
}

func TestGapB88YasminExilesTheTopCardAndYouMayPlayItUntilYourNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	yasmin := apaPush(g, me.ID, me.ID, game.Card{Name: "Yasmin Khan", OracleID: gapB88YasminOracle,
		TypeLine: "Legendary Creature — Human Detective", Power: 3, Toughness: 3})
	top := seedLibraryTop(me, "Top Land", "Basic Land — Forest", "")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, yasmin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := gapB88InExile(g, top); !ok {
		t.Fatal("the top card was not exiled")
	}
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("playing the exiled land: %v", err)
	}
	if !g.Battlefield.Contains(top) {
		t.Error("the land was not played")
	}
}

func TestGapB88FrostFairLureFishMakesFishAndTreasuresAndGrantsHasteAndEvasion(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Frost Fair Lure Fish", "Creature — Fish", gapB88LureFishOracle, nil)
	passPriorityAroundTable(t, g)
	fish, treasures := 0, 0
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller != me.ID {
			continue
		}
		if c.HasSubtype("Fish") && c.IsToken() {
			fish++
			token = c.InstanceID
		}
		if c.HasSubtype("Treasure") {
			treasures++
			if !c.Tapped {
				t.Error("a Treasure entered untapped")
			}
		}
	}
	if fish != 2 || treasures != 2 {
		t.Fatalf("%d Fish tokens and %d Treasures, want 2 and 2", fish, treasures)
	}
	if !hasEffectiveKeyword(t, g, token, "haste") {
		t.Error("a Fish token has no haste")
	}
	human := b12Creature(g, opp.ID, "Human", "Creature — Human", 1, 1)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, token).SummonedThisTurn = false })
	declareAttack(t, g, opp.ID, token)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(human, token); err == nil {
		t.Error("a Human blocked a Fish")
	}
}

func TestGapB88SibyllineSoothsayerExilesTheFirstBigNonlandCardSuspendedAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// Top to bottom: a land, a cheap spell, a three-drop, a four-drop.
	four := seedLibraryTop(me, "Four", "Sorcery", "{3}{R}")
	three := seedLibraryTop(me, "Three", "Sorcery", "{2}{R}")
	cheap := seedLibraryTop(me, "Cheap", "Instant", "{R}")
	land := seedLibraryTop(me, "Land", "Basic Land — Mountain", "")
	castCatalogSpell(t, g, "Sibylline Soothsayer", "Creature — Human Warlock", gapB88SoothsayerOracl, nil)
	passPriorityAroundTable(t, g)
	c, ok := gapB88InExile(g, three)
	if !ok {
		t.Fatal("the three-mana-value card was not exiled")
	}
	if c.Counters[game.CounterTime] != 3 || !game.CardHasSuspend(c) {
		t.Errorf("the exiled card has %d time counters, suspend %v; want 3 and suspend", c.Counters[game.CounterTime], game.CardHasSuspend(c))
	}
	if _, ok := gapB88InExile(g, four); ok {
		t.Error("the four-mana-value card was exiled too")
	}
	cards := me.Library.Cards
	if n := len(cards); n < 3 {
		t.Fatalf("library has %d cards", n)
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range cards[:3] {
		bottom[c.InstanceID] = true
	}
	for _, id := range []uuid.UUID{land, cheap} {
		if !bottom[id] {
			t.Errorf("revealed card %s is not at the bottom of the library", id)
		}
	}
}

func TestGapB88TenthDoctorExilesASuspendedCardWhenYouAttackAndTimeTravelsThreeTimes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doc := apaPush(g, me.ID, me.ID, game.Card{Name: "The Tenth Doctor", OracleID: gapB88TenthOracle,
		TypeLine: "Legendary Creature — Time Lord Doctor", Power: 3, Toughness: 5})
	nonland := seedLibraryTop(me, "Spell", "Sorcery", "{1}{R}")
	land := seedLibraryTop(me, "Land", "Basic Land — Mountain", "")
	declareAttack(t, g, opp.ID, doc)
	passPriorityAroundTable(t, g)
	c, ok := gapB88InExile(g, nonland)
	if !ok || c.Counters[game.CounterTime] != 3 || !game.CardHasSuspend(c) {
		t.Fatalf("the nonland card is %+v in exile, want 3 time counters and suspend", c)
	}
	if lc, ok := gapB88InExile(g, land); !ok || lc.Counters[game.CounterTime] != 0 {
		t.Errorf("the land exiled on the way: %+v", lc)
	}
	// Timey-Wimey: three times, each adding a counter to the suspended card.
	advanceToMain(t, g)
	g.WithWriteLock(func() { fillPool(me, 7) })
	if err := g.ActivateCatalogAbility(me.ID, doc, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("timey-wimey: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		pick := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
		if pick == nil {
			t.Fatalf("time travel walk %d never asked", i+1)
		}
		if err := g.ResolveOptionPick(pick.ID, me.ID, 0); err != nil {
			t.Fatalf("answering walk %d: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	c, _ = gapB88InExile(g, nonland)
	if c.Counters[game.CounterTime] != 6 {
		t.Errorf("after three time travels the card has %d time counters, want 6", c.Counters[game.CounterTime])
	}
}

func TestGapB88SontaranGeneralGoadsAndStopsBlockersOnAnAttackOfThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	general := apaPush(g, me.ID, me.ID, game.Card{Name: "Sontaran General", OracleID: gapB88GeneralOracle,
		TypeLine: "Creature — Alien Soldier", Power: 5, Toughness: 5, Keywords: []string{"trample", "haste"}})
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 1)
	declareAttack(t, g, opp.ID, general, a, b)
	b04WaitForPickOrSkip(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	c := gapB86EffectiveForB88(t, g, victim)
	if !c.IsGoadedBy(me.ID) {
		t.Error("the chosen creature was not goaded")
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(victim, a); err == nil {
		t.Error("the goaded creature blocked; it can't block this turn")
	}
}

func gapB86EffectiveForB88(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() { c, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

func TestGapB88RyanSinclairOffersAFreeCastOfAnAffordableCardAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ryan := apaPush(g, me.ID, me.ID, game.Card{Name: "Ryan Sinclair", OracleID: gapB88RyanOracle,
		TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
	cheap := seedLibraryTop(me, "Cheap Sorcery", "Sorcery", "{1}{R}")
	land := seedLibraryTop(me, "Land", "Basic Land — Mountain", "")
	declareAttack(t, g, opp.ID, ryan)
	passPriorityAroundTable(t, g)
	prompt := keywordCastPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no free-cast offer for a card with mana value 2 and Ryan's power 2")
	}
	answerKeywordCast(t, g, prompt, true)
	passPriorityAroundTable(t, g)
	if _, ok := gapB88InExile(g, cheap); ok {
		t.Error("the cast card is still in exile")
	}
	if _, ok := gapB88InExile(g, land); ok {
		t.Error("the land exiled on the way was not put on the bottom")
	}
	found := false
	for _, c := range me.Library.Cards[:1] {
		found = found || c.InstanceID == land
	}
	if !found {
		t.Error("the land is not at the bottom of the library")
	}
}

func TestGapB88RyanSinclairSkipsTheOfferForACardAboveHisPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ryan := apaPush(g, me.ID, me.ID, game.Card{Name: "Ryan Sinclair", OracleID: gapB88RyanOracle,
		TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
	big := seedLibraryTop(me, "Big Sorcery", "Sorcery", "{4}{R}")
	declareAttack(t, g, opp.ID, ryan)
	passPriorityAroundTable(t, g)
	if keywordCastPromptFor(g, me.ID) != nil {
		t.Fatal("offered a free cast of a mana value 5 card to a 2-power Ryan")
	}
	if _, ok := gapB88InExile(g, big); ok {
		t.Error("the too-expensive card stayed in exile")
	}
	if c := me.Library.Cards[0]; c.InstanceID != big {
		t.Errorf("the bottom card is %q, want the exiled Big Sorcery", c.Name)
	}
}
