package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jodi_free_cast_test.go — S58 PR 5: Ugin's Binding, Sunbird's
// Invocation, Emergent Ultimatum and Nicol Bolas, God-Pharaoh.

// castFromExile casts `id` from exile with the free-cast grant,
// non-strict as the harness does everywhere.
func castFromExile(t *testing.T, g *game.Game, caster, id uuid.UUID) error {
	t.Helper()
	return g.CastSpell(caster, id, game.CastSpellParams{FromZone: "exile"})
}

// --- Ugin's Binding ----------------------------------------------------

func TestUginsBindingBouncesATargetNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushCatalogPermanent(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", false)
	land := b31Push(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	if err := castCatalogSpellErr(t, g, "Ugin's Binding", "Instant", uginsBindingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}}); err == nil {
		t.Error("a land is not a legal target")
	}
	castCatalogSpell(t, g, "Ugin's Binding", "Instant", uginsBindingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the target is returned to its owner's hand")
	}
}

// pushGraveyardBinding puts Ugin's Binding in `p`'s graveyard.
func pushGraveyardBinding(p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Ugin's Binding", TypeLine: "Instant",
		OracleID: uginsBindingOracle, ManaCost: "{2}{U}", Colors: nil, Owner: p.ID, Controller: p.ID})
	return id
}

func TestUginsBindingFromTheGraveyardReturnsEveryNonlandPermanentYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	binding := pushGraveyardBinding(me)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	rock := pushCatalogPermanent(g, third.ID, "Rock", "Artifact", "", false)
	land := b31Push(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	mine := pushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)

	castSpellWithCost(t, g, "Colorless Titan", "Creature — Eldrazi", "", "{8}", nil)
	passPriorityUntilChoice(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("casting a colorless spell with mana value 8 should offer the exile")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if !inExile(g, binding) {
		t.Fatal("Ugin's Binding should be exiled from the graveyard")
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(rock) {
		t.Error("each nonland permanent you don't control is returned")
	}
	if !opp.Hand.Contains(bear) || !third.Hand.Contains(rock) {
		t.Error("they go to their owners' hands")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("lands are not returned")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own permanents stay")
	}
}

func TestUginsBindingFromTheGraveyardMayBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	binding := pushGraveyardBinding(me)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	castSpellWithCost(t, g, "Colorless Titan", "Creature — Eldrazi", "", "{7}", nil)
	passPriorityUntilChoice(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("mana value 7 is enough")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(binding) || !g.Battlefield.Contains(bear) {
		t.Error("declining leaves the card in the graveyard and bounces nothing")
	}
}

func TestUginsBindingFromTheGraveyardNeedsAColorlessSpellOfValueSevenOrMore(t *testing.T) {
	for _, tc := range []struct {
		name, cost string
		colors     []string
	}{
		{"colorless but only six", "{6}", nil},
		{"seven but green", "{6}{G}", []string{"G"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			pushGraveyardBinding(me)
			id := putInHand(me, game.Card{Name: "A Spell", TypeLine: "Creature — Bear", ManaCost: tc.cost, Colors: tc.colors})
			toMain(t, g)
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			if len(g.PendingTriggers) != 0 || triggerOnStackFrom(g, "Ugin's Binding") {
				t.Error("the graveyard ability must not trigger")
			}
		})
	}
}

// triggerOnStackFrom reports whether a triggered ability whose source
// is a card named `name` is waiting on the stack.
func triggerOnStackFrom(g *game.Game, name string) bool {
	for _, item := range g.StackMeta {
		if item == nil || item.Kind != game.StackItemTriggered {
			continue
		}
		if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && c.Name == name {
			return true
		}
	}
	return false
}

// --- Sunbird's Invocation ----------------------------------------------

// sunbirdLibrary stacks the top of `p`'s library, top card last in the
// argument order given (the first argument ends up on top).
func sunbirdLibrary(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		ids[i] = c.InstanceID
		p.Library.PushTop(c)
	}
	return ids
}

func TestSunbirdsInvocationRevealsTopXAndCastsOneFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Sunbird's Invocation", "Enchantment", sunbirdsInvocationOracle, false)
	ids := sunbirdLibrary(me,
		game.Card{Name: "A Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Costly Wurm", TypeLine: "Creature — Wurm", ManaCost: "{3}{G}", Power: 4, Toughness: 4},
	)
	forest, bear, wurm := ids[0], ids[1], ids[2]
	castSpellWithCost(t, g, "Three Drop", "Sorcery", "", "{2}{R}", nil)
	passPriorityUntilChoice(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the reveal should offer a free cast")
	}
	if !hasID(c.ChooseCards, bear) || hasID(c.ChooseCards, forest) || hasID(c.ChooseCards, wurm) {
		t.Errorf("offered %v: only the nonland card with mana value 3 or less may be cast", c.ChooseCards)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !inExile(g, bear) {
		t.Fatal("the chosen card is held in exile for its free cast")
	}
	for _, id := range []uuid.UUID{forest, wurm} {
		if i := libraryIndex(me, id); i < 0 || i > 1 {
			t.Errorf("the rest should be on the bottom of the library, found at index %d", i)
		}
	}
	// The window closes on the caster's next pass, so the Bear is cast
	// now, with the Three Drop still on the stack beneath it.
	if err := castFromExile(t, g, me.ID, bear); err != nil {
		t.Fatalf("free cast of the revealed card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the free spell should resolve")
	}
}

func TestSunbirdsInvocationDeclineBottomsEverything(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Sunbird's Invocation", "Enchantment", sunbirdsInvocationOracle, false)
	ids := sunbirdLibrary(me,
		game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Cheap Elk", TypeLine: "Creature — Elk", ManaCost: "{G}"},
	)
	castSpellWithCost(t, g, "Two Drop", "Sorcery", "", "{1}{R}", nil)
	passPriorityUntilChoice(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the reveal should offer a free cast")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	for _, id := range ids {
		if i := libraryIndex(me, id); i < 0 || i > 1 {
			t.Errorf("a declined card is bottomed with the rest, found at index %d", i)
		}
	}
}

// A spell cast from anywhere but your hand does not trigger it.
func TestSunbirdsInvocationIgnoresSpellsNotCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Sunbird's Invocation", "Enchantment", sunbirdsInvocationOracle, false)
	sunbirdLibrary(me, game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"})
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: id, Name: "Exiled Spell", TypeLine: "Sorcery", ManaCost: "{2}",
			Owner: me.ID, Controller: me.ID})
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player: me.ID, Zone: game.ZoneExile, Cost: "{0}", CastOnly: true,
			Duration: g.UntilEndOfTurnDuration(),
		}, []game.Card{{InstanceID: id, Owner: me.ID}})
	})
	toMain(t, g)
	if err := castFromExile(t, g, me.ID, id); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a spell cast from exile does not trigger Sunbird's Invocation")
	}
}

// --- Emergent Ultimatum -------------------------------------------------

func TestEmergentUltimatumExilesChoosesShufflesAndCasts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := sunbirdLibrary(me,
		game.Card{Name: "White Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{W}"},
		game.Card{Name: "Blue Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{U}"},
		game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Gold Bear", TypeLine: "Creature — Bear", ManaCost: "{W}{U}"},
		game.Card{Name: "White Bear", TypeLine: "Creature — Bear", ManaCost: "{W}"},
		game.Card{Name: "A Land", TypeLine: "Basic Land — Forest"},
	)
	white, blue, green, gold, whiteDup := ids[0], ids[1], ids[2], ids[3], ids[4]
	spell := castCatalogSpell(t, g, "Emergent Ultimatum", "Sorcery", emergentUltimatumOracle, nil)
	passPriorityUntilChoice(t, g)
	s := searchChoiceFor(g, me.ID)
	if s == nil {
		t.Fatal("no search prompt")
	}
	for _, bad := range []uuid.UUID{gold, ids[5]} {
		if hasID(s.SearchCards, bad) {
			t.Error("only monocolored cards may be found")
		}
	}
	if err := g.ResolveSearchLibrary(s.ID, me.ID, []uuid.UUID{white, whiteDup, blue}); err == nil {
		t.Error("two cards with the same name are refused")
	}
	if err := g.ResolveSearchLibrary(s.ID, me.ID, []uuid.UUID{white, blue, green}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	for _, id := range []uuid.UUID{white, blue, green} {
		if !inExile(g, id) {
			t.Fatalf("a found card is not in exile")
		}
	}
	// "An opponent chooses": the caster picks which one.
	oc := latestChoiceOfKind(g, game.PendingChoiceOptionPick)
	if oc == nil {
		t.Fatal("the caster should choose which opponent picks")
	}
	if err := g.ResolveOptionPick(oc.ID, me.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	rp := latestChoiceOfKind(g, game.PendingChoiceRevealPick)
	if rp == nil || rp.Chooser != opp.ID {
		t.Fatalf("the chosen opponent should pick a card, got %+v", rp)
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{white}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if inExile(g, white) || !libraryHolds(me, white) {
		t.Error("the chosen card is shuffled into the library")
	}
	if !g.Exile.Contains(spell) {
		t.Error("Emergent Ultimatum exiles itself")
	}
	// The other two may be cast for free; a sorcery-speed cast during
	// the resolution is what the flash-timed grant allows.
	for _, id := range []uuid.UUID{blue, green} {
		if err := castFromExile(t, g, me.ID, id); err != nil {
			t.Fatalf("free cast: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(blue) || !g.Battlefield.Contains(green) {
		t.Error("both free spells resolve")
	}
}

// --- Nicol Bolas, God-Pharaoh -------------------------------------------

func activateBolas(t *testing.T, g *game.Game, me *game.Player, bolas uuid.UUID, idx int, targets ...game.TargetRef) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, bolas, idx, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("activate ability %d: %v", idx, err)
	}
}

func TestNicolBolasPlusTwoExilesUntilANonlandAndLetsYouCastIt(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bolas := pushCatalogWalker(g, me.ID, "Nicol Bolas, God-Pharaoh", nicolBolasOracle, 7)
	ids := sunbirdLibrary(opp,
		game.Card{Name: "Land One", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Land Two", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Their Spell", TypeLine: "Sorcery", ManaCost: "{4}{R}"},
		game.Card{Name: "Next Card", TypeLine: "Creature — Bear", ManaCost: "{2}"},
	)
	activateBolas(t, g, me, bolas, 0, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	passPriorityAroundTable(t, g)

	if loyaltyCount(g, bolas) != 9 {
		t.Errorf("+2 from 7 should be 9 loyalty, got %d", loyaltyCount(g, bolas))
	}
	for _, id := range ids[:3] {
		if !inExile(g, id) {
			t.Errorf("a card up to and including the first nonland is exiled")
		}
	}
	if inExile(g, ids[3]) {
		t.Error("exiling stops at the first nonland card")
	}
	if err := castFromExile(t, g, me.ID, ids[2]); err != nil {
		t.Fatalf("free cast of the exiled spell: %v", err)
	}
	if err := castFromExile(t, g, me.ID, ids[0]); err == nil {
		t.Error("an exiled land is not castable")
	}
}

func TestNicolBolasPlusOneExilesTwoCardsFromEachOpponentsHand(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bolas := pushCatalogWalker(g, me.ID, "Nicol Bolas, God-Pharaoh", nicolBolasOracle, 7)
	activateBolas(t, g, me, bolas, 1)
	passPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		before := p.Hand.Size()
		c := chooseCardsChoiceFor(g, p.ID)
		if c == nil {
			t.Fatalf("seat %v should be asked to exile two cards", p.ID)
		}
		if c.ChooseMin != 2 || c.ChooseMax != 2 {
			t.Errorf("bounds %d..%d, want exactly two", c.ChooseMin, c.ChooseMax)
		}
		picks := []uuid.UUID{c.ChooseCards[0], c.ChooseCards[1]}
		if err := g.ResolveChooseCards(c.ID, p.ID, picks); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		if got := p.Hand.Size(); got != before-2 {
			t.Errorf("hand %d -> %d, want 2 fewer", before, got)
		}
		for _, id := range picks {
			if !inExile(g, id) || p.Graveyard.Contains(id) {
				t.Error("the cards are exiled, not discarded")
			}
		}
	}
}

func TestNicolBolasMinusFourDealsSevenToAnOpponentCreatureOrWalker(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bolas := pushCatalogWalker(g, me.ID, "Nicol Bolas, God-Pharaoh", nicolBolasOracle, 7)
	mine := pushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Giant",
		TypeLine: "Creature — Giant", Power: 8, Toughness: 7, Owner: opp.ID, Controller: opp.ID})

	if err := g.ActivateCatalogAbility(me.ID, bolas, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}}); err == nil {
		t.Error("the -4 cannot target your own creature")
	}
	life := opp.Life
	activateBolas(t, g, me, bolas, 2, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("7 damage kills a 8/7")
	}
	if loyaltyCount(g, bolas) != 3 {
		t.Errorf("loyalty %d, want 3", loyaltyCount(g, bolas))
	}
	if opp.Life != life {
		t.Error("the creature was the target, not the player")
	}
}

func TestNicolBolasMinusFourCanHitAnOpponentButNotYou(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bolas := pushCatalogWalker(g, me.ID, "Nicol Bolas, God-Pharaoh", nicolBolasOracle, 7)
	if err := g.ActivateCatalogAbility(me.ID, bolas, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("the -4 cannot target you")
	}
	life := opp.Life
	activateBolas(t, g, me, bolas, 2, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	passPriorityAroundTable(t, g)
	if opp.Life != life-7 {
		t.Errorf("life %d -> %d, want 7 lost", life, opp.Life)
	}
}

func TestNicolBolasMinusTwelveExilesEveryOpponentNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	bolas := pushCatalogWalker(g, me.ID, "Nicol Bolas, God-Pharaoh", nicolBolasOracle, 13)
	a := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	b := pushCatalogPermanent(g, third.ID, "Rock", "Artifact", "", false)
	land := b31Push(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	mine := pushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)
	activateBolas(t, g, me, bolas, 3)
	passPriorityAroundTable(t, g)
	if !inExile(g, a) || !inExile(g, b) {
		t.Error("each nonland permanent your opponents control is exiled")
	}
	if !g.Battlefield.Contains(land) || !g.Battlefield.Contains(mine) {
		t.Error("lands and your own permanents stay")
	}
}
