package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Ugin's Binding ----------------------------------------------------

func TestUginsBindingBouncesATargetNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", false)
	land := effectstest.B31Push(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	if err := effectstest.CastCatalogSpellErr(t, g, "Ugin's Binding", "Instant", effectstest.UginsBindingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}}); err == nil {
		t.Error("a land is not a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Ugin's Binding", "Instant", effectstest.UginsBindingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the target is returned to its owner's hand")
	}
}

func TestUginsBindingFromTheGraveyardReturnsEveryNonlandPermanentYouDontControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	binding := effectstest.PushGraveyardBinding(me)
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	rock := effectstest.PushCatalogPermanent(g, third.ID, "Rock", "Artifact", "", false)
	land := effectstest.B31Push(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)

	effectstest.CastSpellWithCost(t, g, "Colorless Titan", "Creature — Eldrazi", "", "{8}", nil)
	effectstest.PassPriorityUntilChoice(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("casting a colorless spell with mana value 8 should offer the exile")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if !effectstest.InExile(g, binding) {
		t.Fatal("Ugin's Binding should be exiled from the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	binding := effectstest.PushGraveyardBinding(me)
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	effectstest.CastSpellWithCost(t, g, "Colorless Titan", "Creature — Eldrazi", "", "{7}", nil)
	effectstest.PassPriorityUntilChoice(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("mana value 7 is enough")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.PushGraveyardBinding(me)
			id := effectstest.PutInHand(me, game.Card{Name: "A Spell", TypeLine: "Creature — Bear", ManaCost: tc.cost, Colors: tc.colors})
			effectstest.ToMain(t, g)
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			if len(g.PendingTriggers) != 0 || effectstest.TriggerOnStackFrom(g, "Ugin's Binding") {
				t.Error("the graveyard ability must not trigger")
			}
		})
	}
}

func TestSunbirdsInvocationRevealsTopXAndCastsOneFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Sunbird's Invocation", "Enchantment", effectstest.SunbirdsInvocationOracle, false)
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "A Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Costly Wurm", TypeLine: "Creature — Wurm", ManaCost: "{3}{G}", Power: 4, Toughness: 4},
	)
	forest, bear, wurm := ids[0], ids[1], ids[2]
	effectstest.CastSpellWithCost(t, g, "Three Drop", "Sorcery", "", "{2}{R}", nil)
	effectstest.PassPriorityUntilChoice(t, g)
	// #4250: the cast is offered inside the trigger's resolution.
	c := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, bear)
	if effectstest.HasID(c.DiscardOptions, forest) || effectstest.HasID(c.DiscardOptions, wurm) {
		t.Errorf("offered %v: only the nonland card with mana value 3 or less may be cast", c.DiscardOptions)
	}
	if err := g.CastSpell(me.ID, wurm, game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Error("the offer opens the eligible card and no other card in the library")
	}
	// The Bear is cast from where it sits, with the Three Drop still on
	// the stack beneath it.
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("free cast of the revealed card from the library: %v", err)
	}
	if effectstest.InExile(g, bear) || effectstest.LibraryIndex(me, bear) >= 0 || !g.Stack.Contains(bear) {
		t.Error("the cast takes the card from the library to the stack, not through exile")
	}
	effectstest.CastInsideTheResolution(t, g, bear)
	for _, id := range []uuid.UUID{forest, wurm} {
		if i := effectstest.LibraryIndex(me, id); i < 0 || i > 1 {
			t.Errorf("the rest should be on the bottom of the library, found at index %d", i)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the free spell should resolve")
	}
}

// The cast event names the library as where the spell came from, which
// is what a "cast from your library" reader and Sunbird's own
// hand-only trigger both read.
func TestSunbirdsInvocationCastComesFromTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Sunbird's Invocation", "Enchantment", effectstest.SunbirdsInvocationOracle, false)
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "Cheap Sorcery", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "Filler", TypeLine: "Basic Land — Mountain"},
	)
	sorcery := ids[0]
	effectstest.CastSpellWithCost(t, g, "Two Drop", "Sorcery", "", "{1}{R}", nil)
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, sorcery)
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("free cast from the library: %v", err)
	}
	var from game.ZoneKind
	for _, ev := range g.Events {
		if ev.Kind == game.EventCast && ev.CardID == sorcery {
			from = ev.OldZone
		}
	}
	if from != game.ZoneLibrary {
		t.Errorf("the free spell was cast from %q, want the library", from)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a spell cast from the library does not trigger Sunbird's Invocation again")
	}
}

// Declining (#4250: the resolution_cast's empty answer) bottoms the
// eligible card with the rest.
func TestSunbirdsInvocationUncastPickGoesToTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Sunbird's Invocation", "Enchantment", effectstest.SunbirdsInvocationOracle, false)
	for range 5 {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Deep Card", TypeLine: "Basic Land — Mountain",
			Owner: me.ID, Controller: me.ID})
	}
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Filler", TypeLine: "Basic Land — Forest"},
	)
	bear := ids[0]
	effectstest.CastSpellWithCost(t, g, "Two Drop", "Sorcery", "", "{1}{R}", nil)
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, bear)
	if i := effectstest.LibraryIndex(me, bear); i != len(me.Library.Cards)-1 {
		t.Fatalf("the offered card waits on top of the library, found at index %d", i)
	}
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if i := effectstest.LibraryIndex(me, bear); i < 0 || i > 1 {
		t.Errorf("an uncast card goes to the bottom of the library, found at index %d", i)
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Error("the window closed on the decline")
	}
}

func TestSunbirdsInvocationDeclineBottomsEverything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Sunbird's Invocation", "Enchantment", effectstest.SunbirdsInvocationOracle, false)
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Cheap Elk", TypeLine: "Creature — Elk", ManaCost: "{G}"},
	)
	effectstest.CastSpellWithCost(t, g, "Two Drop", "Sorcery", "", "{1}{R}", nil)
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, ids...)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	for _, id := range ids {
		if i := effectstest.LibraryIndex(me, id); i < 0 || i > 1 {
			t.Errorf("a declined card is bottomed with the rest, found at index %d", i)
		}
	}
}

// A spell cast from anywhere but your hand does not trigger it.
func TestSunbirdsInvocationIgnoresSpellsNotCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Sunbird's Invocation", "Enchantment", effectstest.SunbirdsInvocationOracle, false)
	effectstest.SunbirdLibrary(me, game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"})
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: id, Name: "Exiled Spell", TypeLine: "Sorcery", ManaCost: "{2}",
			Owner: me.ID, Controller: me.ID})
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player: me.ID, Zone: game.ZoneExile, Cost: "{0}", CastOnly: true,
			Duration: g.UntilEndOfTurnDuration(),
		}, []game.Card{{InstanceID: id, Owner: me.ID}})
	})
	effectstest.ToMain(t, g)
	if err := effectstest.CastFromExile(t, g, me.ID, id); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a spell cast from exile does not trigger Sunbird's Invocation")
	}
}

// --- Emergent Ultimatum -------------------------------------------------

func TestEmergentUltimatumExilesChoosesShufflesAndCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "White Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{W}"},
		game.Card{Name: "Blue Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{U}"},
		game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Gold Bear", TypeLine: "Creature — Bear", ManaCost: "{W}{U}"},
		game.Card{Name: "White Bear", TypeLine: "Creature — Bear", ManaCost: "{W}"},
		game.Card{Name: "A Land", TypeLine: "Basic Land — Forest"},
	)
	white, blue, green, gold, whiteDup := ids[0], ids[1], ids[2], ids[3], ids[4]
	spell := effectstest.CastCatalogSpell(t, g, "Emergent Ultimatum", "Sorcery", effectstest.EmergentUltimatumOracle, nil)
	effectstest.PassPriorityUntilChoice(t, g)
	s := effectstest.SearchChoiceFor(g, me.ID)
	if s == nil {
		t.Fatal("no search prompt")
	}
	for _, bad := range []uuid.UUID{gold, ids[5]} {
		if effectstest.HasID(s.SearchCards, bad) {
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
		if !effectstest.InExile(g, id) {
			t.Fatalf("a found card is not in exile")
		}
	}
	// "An opponent chooses": the caster picks which one.
	oc := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick)
	if oc == nil {
		t.Fatal("the caster should choose which opponent picks")
	}
	if err := g.ResolveOptionPick(oc.ID, me.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	rp := effectstest.LatestChoiceOfKind(g, game.PendingChoiceRevealPick)
	if rp == nil || rp.Chooser != opp.ID {
		t.Fatalf("the chosen opponent should pick a card, got %+v", rp)
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{white}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if effectstest.InExile(g, white) || !effectstest.LibraryHolds(me, white) {
		t.Error("the chosen card is shuffled into the library")
	}
	if !g.Exile.Contains(spell) {
		t.Error("Emergent Ultimatum exiles itself")
	}
	// The other two may be cast for free; a sorcery-speed cast during
	// the resolution is what the flash-timed grant allows.
	for _, id := range []uuid.UUID{blue, green} {
		if err := effectstest.CastFromExile(t, g, me.ID, id); err != nil {
			t.Fatalf("free cast: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(blue) || !g.Battlefield.Contains(green) {
		t.Error("both free spells resolve")
	}
}

func TestNicolBolasPlusTwoExilesUntilANonlandAndLetsYouCastIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bolas := effectstest.PushCatalogWalker(g, me.ID, "Nicol Bolas, God-Pharaoh", effectstest.NicolBolasOracle, 7)
	ids := effectstest.SunbirdLibrary(opp,
		game.Card{Name: "Land One", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Land Two", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Their Spell", TypeLine: "Sorcery", ManaCost: "{4}{R}"},
		game.Card{Name: "Next Card", TypeLine: "Creature — Bear", ManaCost: "{2}"},
	)
	effectstest.ActivateBolas(t, g, me, bolas, 0, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.LoyaltyCount(g, bolas) != 9 {
		t.Errorf("+2 from 7 should be 9 loyalty, got %d", effectstest.LoyaltyCount(g, bolas))
	}
	for _, id := range ids[:3] {
		if !effectstest.InExile(g, id) {
			t.Errorf("a card up to and including the first nonland is exiled")
		}
	}
	if effectstest.InExile(g, ids[3]) {
		t.Error("exiling stops at the first nonland card")
	}
	if err := effectstest.CastFromExile(t, g, me.ID, ids[2]); err != nil {
		t.Fatalf("free cast of the exiled spell: %v", err)
	}
	if err := effectstest.CastFromExile(t, g, me.ID, ids[0]); err == nil {
		t.Error("an exiled land is not castable")
	}
}

func TestNicolBolasPlusOneExilesTwoCardsFromEachOpponentsHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bolas := effectstest.PushCatalogWalker(g, me.ID, "Nicol Bolas, God-Pharaoh", effectstest.NicolBolasOracle, 7)
	effectstest.ActivateBolas(t, g, me, bolas, 1)
	effectstest.PassPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		before := p.Hand.Size()
		c := effectstest.ChooseCardsChoiceFor(g, p.ID)
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
			if !effectstest.InExile(g, id) || p.Graveyard.Contains(id) {
				t.Error("the cards are exiled, not discarded")
			}
		}
	}
}

func TestNicolBolasMinusFourDealsSevenToAnOpponentCreatureOrWalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bolas := effectstest.PushCatalogWalker(g, me.ID, "Nicol Bolas, God-Pharaoh", effectstest.NicolBolasOracle, 7)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Giant",
		TypeLine: "Creature — Giant", Power: 8, Toughness: 7, Owner: opp.ID, Controller: opp.ID})

	if err := g.ActivateCatalogAbility(me.ID, bolas, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}}); err == nil {
		t.Error("the -4 cannot target your own creature")
	}
	life := opp.Life
	effectstest.ActivateBolas(t, g, me, bolas, 2, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("7 damage kills a 8/7")
	}
	if effectstest.LoyaltyCount(g, bolas) != 3 {
		t.Errorf("loyalty %d, want 3", effectstest.LoyaltyCount(g, bolas))
	}
	if opp.Life != life {
		t.Error("the creature was the target, not the player")
	}
}

func TestNicolBolasMinusFourCanHitAnOpponentButNotYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bolas := effectstest.PushCatalogWalker(g, me.ID, "Nicol Bolas, God-Pharaoh", effectstest.NicolBolasOracle, 7)
	if err := g.ActivateCatalogAbility(me.ID, bolas, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("the -4 cannot target you")
	}
	life := opp.Life
	effectstest.ActivateBolas(t, g, me, bolas, 2, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-7 {
		t.Errorf("life %d -> %d, want 7 lost", life, opp.Life)
	}
}

func TestNicolBolasMinusTwelveExilesEveryOpponentNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	bolas := effectstest.PushCatalogWalker(g, me.ID, "Nicol Bolas, God-Pharaoh", effectstest.NicolBolasOracle, 13)
	a := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	b := effectstest.PushCatalogPermanent(g, third.ID, "Rock", "Artifact", "", false)
	land := effectstest.B31Push(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)
	effectstest.ActivateBolas(t, g, me, bolas, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, a) || !effectstest.InExile(g, b) {
		t.Error("each nonland permanent your opponents control is exiled")
	}
	if !g.Battlefield.Contains(land) || !g.Battlefield.Contains(mine) {
		t.Error("lands and your own permanents stay")
	}
}
