package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b23_more_test.go — the rest of S82 batch 23: Arcavios, Alara,
// Ikoria and Begin the Invasion.

// --- Invasion of Arcavios // Invocation of the Founders -------------------------

func arcaviosRow() cards.Card {
	return b23Row(effects.InvasionOfArcaviosOracleID, "Invasion of Arcavios", "Invocation of the Founders", "{3}{U}{U}", "7", "Enchantment", "", "")
}

// b23ArcaviosAsk enters the Siege and answers the zone question with `option`.
func b23ArcaviosAsk(t *testing.T, g *game.Game, option int) *game.Player {
	t.Helper()
	_, owner, _ := b23Enter(t, g, arcaviosRow())
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestOptionPickFor(g, owner.ID)
	if c == nil {
		t.Fatal("the enters trigger did not ask where to search")
	}
	if err := g.ResolveOptionPick(c.ID, owner.ID, option); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	return owner
}

func TestInvasionOfArcaviosSearchesTheLibraryAndGraveyardForAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.B23Library(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
	)
	bear, bolt := ids[0], ids[1]
	burn := effectstest.PushGraveyardCardForTest(me, "Burn")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == burn {
				me.Graveyard.Cards[i].TypeLine = "Sorcery"
			}
		}
	})
	b23ArcaviosAsk(t, g, effects.ArcaviosLibraryOnly)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	c = effectstest.SearchChoiceFor(g, me.ID)
	if !slices.Contains(c.SearchCards, bolt) || !slices.Contains(c.SearchCards, burn) || slices.Contains(c.SearchCards, bear) {
		t.Fatalf("candidates %v, want the instant and the graveyard sorcery but not the creature", c.SearchCards)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{burn}); err != nil {
		t.Fatalf("take the graveyard card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(burn) {
		t.Error("the card did not go into the hand")
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("the outside-the-game half was asked though the player searched only the library and graveyard")
	}
}

func TestInvasionOfArcaviosSearchesOutsideTheGame(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 4, effectstest.OutsideCard("Lightning Bolt", "Instant"), effectstest.OutsideCard("Bear", "Creature — Bear"))
	me := g.Seats[g.Turn.ActiveSeat]
	if me != g.Seats[0] {
		t.Skip("the sideboard belongs to seat 0")
	}
	b23ArcaviosAsk(t, g, effects.ArcaviosOutsideOnly)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no outside pick: %+v", g.PendingChoices)
	}
	if got := effectstest.CandidateNames(g, c); len(got) != 1 || got[0] != "Lightning Bolt" {
		t.Fatalf("candidates %v, want only the instant", got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, "Lightning Bolt")}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Lightning Bolt") {
		t.Error("the card did not come into the hand")
	}
}

func TestInvasionOfArcaviosBothHalvesAsksOutsideWhenTheLibraryHasNothing(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 4, effectstest.OutsideCard("Lightning Bolt", "Instant"))
	me := g.Seats[g.Turn.ActiveSeat]
	if me != g.Seats[0] {
		t.Skip("the sideboard belongs to seat 0")
	}
	me.Graveyard.Cards = nil
	effectstest.B23Library(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	b23ArcaviosAsk(t, g, effects.ArcaviosBoth)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		effectstest.BeginSearchIfAsked(t, g, c)
		if c = effectstest.SearchChoiceFor(g, me.ID); c != nil {
			if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
				t.Fatalf("find nothing: %v", err)
			}
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("the library found nothing and the outside half was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, "Lightning Bolt")}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !effectstest.OutsideHandHas(me, "Lightning Bolt") {
		t.Error("the outside card did not come into the hand")
	}
}

func TestInvocationOfTheFoundersCopiesAnInstantOrSorceryCastFromYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	invocation := effectstest.PushCatalogPermanent(g, me.ID, "Invocation of the Founders", "Enchantment", effects.InvasionOfArcaviosOracleID+"#1", false)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if effectstest.TriggerOnStack(g, invocation) == nil {
		t.Fatal("casting an instant from hand did not trigger Invocation of the Founders")
	}
	before := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	// The copy asks for new targets; keep the same one.
	for i := 0; i < 8 && !effectstest.StackFullyEmpty(g); i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePickTarget {
				pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
				if err := g.ResolvePickTarget(c.ID, c.Chooser, pick); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
				}
				break
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := before - opp.Life; got != 6 {
		t.Errorf("the original and its copy dealt %d damage, want 3 + 3", got)
	}
}

func TestInvocationOfTheFoundersIgnoresASpellNotCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	invocation := effectstest.PushCatalogPermanent(g, me.ID, "Invocation of the Founders", "Enchantment", effects.InvasionOfArcaviosOracleID+"#1", false)
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID})
	})
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Duration: g.UntilEndOfTurnDuration(), Cost: "{0}", CastOnly: true})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	if effectstest.TriggerOnStack(g, invocation) != nil {
		t.Error("a spell cast from exile triggered Invocation of the Founders")
	}
}

// --- Invasion of Alara // Awaken the Maelstrom ----------------------------------

func alaraRow() cards.Card {
	return b23Row(effects.InvasionOfAlaraOracleID, "Invasion of Alara", "Awaken the Maelstrom", "{W}{U}{B}{R}{G}", "7", "Sorcery", "", "")
}

// alaraLibrary stacks the walk: LandA, Big, A (hit), LandB, B (hit), C.
func alaraLibrary(me *game.Player) (landA, big, a, landB, b, c uuid.UUID) {
	ids := effectstest.B23Library(me,
		game.Card{Name: "LandA", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{5}{G}"},
		game.Card{Name: "A", TypeLine: "Sorcery", ManaCost: "{1}{R}"},
		game.Card{Name: "LandB", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "B", TypeLine: "Sorcery", ManaCost: "{2}{G}"},
		game.Card{Name: "C", TypeLine: "Sorcery", ManaCost: "{1}"},
		game.Card{Name: "D", TypeLine: "Sorcery", ManaCost: "{1}"},
	)
	return ids[0], ids[1], ids[2], ids[3], ids[4], ids[5]
}

func b23CastOffered(t *testing.T, g *game.Game, chooser, card uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if effectstest.IsKeywordCastPrompt(c) && c.Chooser == chooser {
			if err := g.CastForResolutionChoice(chooser, c.ID, card, game.CastSpellParams{FromZone: string(effectstest.OfferedZone(g, chooser, card))}); err != nil {
				t.Fatalf("casting the offered card: %v", err)
			}
			return
		}
	}
	t.Fatal("no free-cast offer is open")
}

func TestInvasionOfAlaraDecliningTheCastPutsOneOfTheTwoInHandAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	landA, big, a, landB, b, c := alaraLibrary(me)
	_, _, _ = b23Enter(t, g, alaraRow())
	effectstest.PassPriorityAroundTable(t, g)

	// Both hits are on offer and the passed-over cards are in exile.
	for _, id := range []uuid.UUID{landA, big, landB} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s was passed over but is not in exile", id)
		}
	}
	if me.Library.Contains(a) || me.Library.Contains(b) {
		t.Error("the two hits are still in the library")
	}
	if !me.Library.Contains(c) {
		t.Error("the walk went past the second hit")
	}
	effectstest.AnswerMayCastPrompt(t, g, me.ID, false)
	choice := effectstest.PendingChooseCards(t, g, me.ID)
	if len(choice.ChooseCards) != 2 || !slices.Contains(choice.ChooseCards, a) || !slices.Contains(choice.ChooseCards, b) {
		t.Fatalf("offered %v, want the two hits", choice.ChooseCards)
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("put one into hand: %v", err)
	}
	if !me.Hand.Contains(b) {
		t.Error("the chosen card is not in the hand")
	}
	// The other hit and the three passed-over cards are the bottom four.
	bottom := effectstest.LibraryBottom(me, 4)
	if !effectstest.SameSet(bottom, []uuid.UUID{a, landA, big, landB}) {
		t.Errorf("bottom four = %v, want the uncast hit and the passed-over cards", bottom)
	}
	if g.Exile.Contains(a) || g.Exile.Contains(landA) {
		t.Error("cards are still in exile")
	}
}

func TestInvasionOfAlaraCastingOneFreePutsTheOtherInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	landA, big, a, landB, b, _ := alaraLibrary(me)
	_, _, _ = b23Enter(t, g, alaraRow())
	effectstest.PassPriorityAroundTable(t, g)
	b23CastOffered(t, g, me.ID, a)

	if !me.Hand.Contains(b) {
		t.Error("the other hit was not put into the hand")
	}
	if me.Hand.Contains(a) {
		t.Error("the cast card is in the hand as well")
	}
	if !g.Stack.Contains(a) {
		t.Error("the cast card is not on the stack")
	}
	if !effectstest.SameSet(effectstest.LibraryBottom(me, 3), []uuid.UUID{landA, big, landB}) {
		t.Errorf("bottom three = %v, want the passed-over cards", effectstest.LibraryBottom(me, 3))
	}
}

func TestInvasionOfAlaraWithOnlyOneHitInTheLibraryPutsItInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.B23Library(me,
		game.Card{Name: "LandA", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Only", TypeLine: "Sorcery", ManaCost: "{1}"},
		game.Card{Name: "LandB", TypeLine: "Basic Land — Forest"},
	)
	_, _, _ = b23Enter(t, g, alaraRow())
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCastPrompt(t, g, me.ID, false)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the only hit was not put into the hand")
	}
	if me.Library.Size() != 2 || !effectstest.SameSet(effectstest.LibraryBottom(me, 2), []uuid.UUID{ids[0], ids[2]}) {
		t.Errorf("library = %d cards, want the two lands on the bottom", me.Library.Size())
	}
}

func TestAwakenTheMaelstromDoesAllFiveThings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Bear A", effectstest.WthBearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", effectstest.WthBearType, 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", effectstest.WthBearType, 2, 2)
	rock := effectstest.PushCatalogHandCard(me, "Rock", "Artifact", "")
	me.Hand.Cards = []game.Card{me.Hand.Cards[len(me.Hand.Cards)-1]}
	handBefore := me.Hand.Size()
	effectstest.B23Library(me, game.Card{Name: "X1", TypeLine: "Sorcery"}, game.Card{Name: "X2", TypeLine: "Sorcery"}, game.Card{Name: "X3", TypeLine: "Sorcery"})

	effectstest.CastCatalogSpell(t, g, "Awaken the Maelstrom", "Sorcery", effects.InvasionOfAlaraOracleID+"#1", []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
		{Kind: game.TargetCard, ID: victim, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	// 1. the target player drew two.
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand grew by %d, want 2 drawn", got)
	}
	// 2. an artifact card may be put onto the battlefield.
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was not put onto the battlefield")
	}
	// 3. a token that's a copy of a permanent you control.
	before := effectstest.CountNamed(g, "Bear A")
	effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {a}})
	// 4. three counters among one to three creatures.
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.AnswerNumber(t, g, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountNamed(g, "Bear A"); got != before+1 {
		t.Errorf("Bear A count %d -> %d, want one more (the token copy)", before, got)
	}
	if n := effectstest.CounterCount(g, a, "+1/+1"); n != 2 {
		t.Errorf("Bear A has %d counters, want the 2 named", n)
	}
	if n := effectstest.CounterCount(g, b, "+1/+1"); n != 1 {
		t.Errorf("Bear B has %d counters, want the 1 left over", n)
	}
	// 5. destroy target permanent an opponent controls.
	if g.Battlefield.Contains(victim) {
		t.Error("the opponent's permanent was not destroyed")
	}
}

// --- Invasion of Ikoria // Zilortha, Apex of Ikoria -----------------------------

func ikoriaRow() cards.Card {
	return b23Row(effects.InvasionOfIkoriaOracleID, "Invasion of Ikoria", "Zilortha, Apex of Ikoria", "{X}{G}{G}", "6", "Legendary Creature — Dinosaur", "8", "8")
}

func TestInvasionOfIkoriaFindsANonHumanCreatureWithManaValueXOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.B23Library(me,
		game.Card{Name: "Human", TypeLine: "Creature — Human", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}", Power: 5, Toughness: 5},
		game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{1}"},
	)
	human, bear, dragon := ids[0], ids[1], ids[2]
	elf := effectstest.PushGraveyardCardForTest(me, "Elf")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == elf {
				me.Graveyard.Cards[i].ManaCost = "{G}"
				me.Graveyard.Cards[i].TypeLine = "Creature — Elf"
			}
		}
	})

	effectstest.ToMainPhaseSameTurn(t, g)
	id := effectstest.ImportToHand(ikoriaRow(), me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast the Siege for X = 3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProtectorPrompt(t, g, me.ID, g.Seats[(g.Turn.ActiveSeat+1)%4].ID)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	c = effectstest.SearchChoiceFor(g, me.ID)
	for _, want := range []uuid.UUID{bear, elf} {
		if !slices.Contains(c.SearchCards, want) {
			t.Errorf("%s is not offered", want)
		}
	}
	for _, bad := range []uuid.UUID{human, dragon} {
		if slices.Contains(c.SearchCards, bad) {
			t.Errorf("%s is offered: a Human, or mana value above X", bad)
		}
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("take the Bear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was not put onto the battlefield")
	}
}

func TestZilorthaGivesYourNonHumanCreaturesTheUnblockedOption(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Zilortha, Apex of Ikoria", TypeLine: "Legendary Creature — Dinosaur", OracleID: effects.InvasionOfIkoriaOracleID + "#1", Power: 8, Toughness: 8})
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	human := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Soldier", TypeLine: "Creature — Human Soldier", Power: 2, Toughness: 2})
	wallA := effectstest.MscWlBlocker(g, "Wall A", 1, 4)
	wallB := effectstest.MscWlBlocker(g, "Wall B", 1, 4)
	effectstest.MscWlBlockedCombat(t, g, []uuid.UUID{bear, human}, map[uuid.UUID]uuid.UUID{wallA: bear, wallB: human})

	c := effectstest.MscWlDamagePrompt(t, g)
	f := c.DamageAssignment
	if !f.AsThoughUnblocked || len(f.BlockerIDs) != 1 || f.BlockerIDs[0] != wallA {
		t.Fatalf("the prompt is %+v, want the Bear's, as-though-unblocked", f)
	}
	if err := effectstest.MscWlAnswer(t, g, c, 2, game.DamageAssignmentEntry{BlockerID: wallA, Amount: 0}); err != nil {
		t.Fatalf("all of it to the player: %v", err)
	}
	if opp.Life != 38 {
		t.Errorf("the defending player has %d life, want 38", opp.Life)
	}
	if got := effectstest.Pr6Marked(g, wallB); got != 2 {
		t.Errorf("the Human's blocker was dealt %d, want its damage assigned to the blocker as usual", got)
	}
}

// --- Begin the Invasion ---------------------------------------------------------

func b23Battle(name string) game.Card {
	row := effectstest.SiegeRow(uuid.NewString(), name, name+" Back", "{2}{R}", "4", "Creature — Elemental", "3", "3")
	return deck.ToGameCard(row, false)
}

func TestBeginTheInvasionFetchesXBattlesWithDifferentNames(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	protector := g.Seats[(g.Turn.ActiveSeat+1)%4]
	one, two, three := b23Battle("Invasion of One"), b23Battle("Invasion of Two"), b23Battle("Invasion of One")
	ids := effectstest.B23Library(me, one, two, three, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	// b23Library keeps the cards' own ids when they have them.
	oneID, twoID, threeID := ids[0], ids[1], ids[2]

	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Begin the Invasion", TypeLine: "Sorcery", ManaCost: "{X}{W}{U}{B}{R}{G}", OracleID: beginTheInvasionOracleID, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast for X = 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	c = effectstest.SearchChoiceFor(g, me.ID)
	if c.SearchMax != 2 {
		t.Errorf("search max = %d, want X = 2", c.SearchMax)
	}
	if slices.ContainsFunc(c.SearchCards, func(id uuid.UUID) bool { return id != oneID && id != twoID && id != threeID }) {
		t.Errorf("candidates %v include a card that is not a battle", c.SearchCards)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{oneID, threeID}); err == nil {
		t.Error("two battles with the same name were accepted")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{oneID, twoID}); err != nil {
		t.Fatalf("take two differently named battles: %v", err)
	}
	// One protector prompt per battle, each answered in turn.
	for _, battle := range []uuid.UUID{oneID, twoID} {
		answered := false
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Kind == game.PendingChoiceChooseProtector && pc.Source == battle {
				if err := g.ResolveChooseProtector(pc.ID, me.ID, protector.ID); err != nil {
					t.Fatalf("protector for %s: %v", battle, err)
				}
				answered = true
				break
			}
		}
		if !answered {
			t.Errorf("no protector prompt for %s", battle)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, battle := range []uuid.UUID{oneID, twoID} {
		card := effectstest.BattlefieldCardFor(g, battle)
		if card == nil {
			t.Errorf("%s is not on the battlefield", battle)
			continue
		}
		if card.Counters[game.CounterDefense] != 4 {
			t.Errorf("%s has %d defense counters, want the printed 4", card.Name, card.Counters[game.CounterDefense])
		}
		if card.ProtectorPlayerID != protector.ID {
			t.Errorf("%s protector = %v, want the chosen opponent", card.Name, card.ProtectorPlayerID)
		}
	}
	if !me.Library.Contains(threeID) {
		t.Error("the unchosen battle left the library")
	}
}

func TestBeginTheInvasionForZeroOnlyShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B23Library(me, b23Battle("Invasion of One"))
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Begin the Invasion", TypeLine: "Sorcery", ManaCost: "{X}{W}{U}{B}{R}{G}", OracleID: beginTheInvasionOracleID, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 0}); err != nil {
		t.Fatalf("cast for X = 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil || me.Library.Size() != 1 {
		t.Error("X = 0 found something")
	}
}
