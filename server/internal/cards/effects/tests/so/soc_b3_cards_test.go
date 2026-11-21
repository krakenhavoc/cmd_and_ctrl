package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b3_cards_test.go — Secrets of Strixhaven Commander, batch 3
// (#4834). Each test plays the real catalog entry and asserts what its
// text says.

const (
	socB3CoerciveOracle     = "dbb8a3d9-6b1b-4ad1-b8b9-f92ead03d1d9"
	socB3CalligrapherOracle = "ee7ebbc1-2f9e-4867-a522-424869b6d772"
	socB3InsightOracle      = "54d7d7f8-22cd-4859-b203-924d248b422b"
	socB3TheoristOracle     = "fa81e7ff-0f2c-40df-9169-c0a22f00cd76"
	socB3LiegeOracle        = "1e67b4eb-e5d2-4368-8284-4ba5c4c42455"
	socB3TechniqueOracle    = "c336b735-3e6b-471e-be7a-84d7402ae2f5"
	socB3CrafterOracle      = "0be907b2-54f9-4c8c-b6af-7c0093892e42"
	socB3DanceOracle        = "615b4189-e800-408b-99e0-71ae7089b937"
	socB3BrewOracle         = "b831d3f8-07ce-4172-ba29-213cac414c9a"
	socB3DenialOracle       = "2d0bd780-b754-44a3-894c-a9c8aa711f05"
)

func TestSocB3EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB3CoerciveOracle:     "Coercive Impetus",
		socB3CalligrapherOracle: "Combat Calligrapher",
		socB3InsightOracle:      "Commander's Insight",
		socB3TheoristOracle:     "Conspiracy Theorist",
		socB3LiegeOracle:        "Creakwood Liege",
		socB3TechniqueOracle:    "Creative Technique",
		socB3CrafterOracle:      "Curiosity Crafter",
		socB3DanceOracle:        "Dance with Calamity",
		socB3BrewOracle:         "Deadly Brew",
		socB3DenialOracle:       "Decisive Denial",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// socB3Library replaces p's library with cards of the given type line and
// mana cost, returning their IDs top first.
func socB3Library(p *game.Player, n int, typeLine, cost string) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = uuid.New()
		p.Library.PushTop(game.Card{InstanceID: ids[i], Name: "Library Card", TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	}
	return ids
}

// socB3InExile reports whether the card is in exile.
func socB3InExile(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneExile
}

// --- Coercive Impetus ---

func TestSocB3CoerciveImpetusPumpsGoadsAndDrawsOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AuraCast(t, g, "Coercive Impetus", socB3CoerciveOracle, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("enchanted bear is %d/%d, want 3/3", p, tg)
	}
	if c, ok := g.LookupCardForEffect(bear); !ok || !c.Goaded() {
		t.Error("the enchanted creature is not goaded")
	}
	effectstest.SeedLibrary(me, "A", "B")
	hand, life := me.Hand.Size(), me.Life
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

// --- Combat Calligrapher ---

func TestSocB3CalligrapherGivesTheAttackerAnInklingAttackingThatOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	// The Calligrapher is the second seat's: the first seat attacks it and
	// the third seat at once.
	cal := effectstest.B12Push(g, opp.ID, "Combat Calligrapher", "Creature — Bird Cleric", socB3CalligrapherOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, at := range []struct{ id, def uuid.UUID }{{a, opp.ID}, {b, third.ID}, {c, third.ID}} {
		if err := g.DeclareAttacker(at.id, at.def); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	if n := effectstest.TriggersOnStackFrom(g, cal); n != 1 {
		t.Fatalf("Calligrapher triggers = %d, want 1 (one opponent of its controller attacked, by two creatures)", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var inkling *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Inkling" {
			if inkling != nil {
				t.Fatal("more than one Inkling")
			}
			inkling = &g.Battlefield.Cards[i]
		}
	}
	if inkling == nil {
		t.Fatal("no Inkling was created")
	}
	if inkling.Controller != me.ID {
		t.Error("the Inkling is not the attacking player's")
	}
	if !inkling.Tapped {
		t.Error("the Inkling is not tapped")
	}
	if inkling.AttackingTarget != third.ID {
		t.Errorf("the Inkling attacks %v, want the third seat", inkling.AttackingTarget)
	}
	if !effectstest.HasEffectiveKeyword(t, g, inkling.InstanceID, "flying") || effectstest.EffectivePower(t, g, inkling.InstanceID) != 2 {
		t.Error("the Inkling is not a 2/1 flier")
	}
}

func TestSocB3CalligrapherKeepsInklingsOffItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, opp.ID, "Combat Calligrapher", "Creature — Bird Cleric", socB3CalligrapherOracle, 3, 3)
	ink := effectstest.B12Creature(g, me.ID, "Inkling", "Creature — Inkling", 2, 1)
	ink2 := effectstest.B12Creature(g, me.ID, "Inkling Two", "Creature — Inkling", 2, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ink, opp.ID); err == nil {
		t.Error("an Inkling was declared attacking the Calligrapher's controller")
	}
	if err := g.DeclareAttacker(ink2, third.ID); err != nil {
		t.Errorf("an Inkling may attack another player: %v", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Errorf("a non-Inkling may attack the Calligrapher's controller: %v", err)
	}
}

// --- Commander's Insight ---

func TestSocB3CommandersInsightDrawsXPlusCommanderCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	opp.CommanderCasts = map[uuid.UUID]int{uuid.New(): 2, uuid.New(): 1}
	effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E", "F", "G")
	hand := opp.Hand.Size()
	effectstest.CastFromHandForTest(t, g, me, "Commander's Insight", "Instant", "{X}{U}{U}{U}", socB3InsightOracle,
		game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 5 {
		t.Errorf("drew %d, want 5 (X=2 plus three commander casts)", got)
	}
}

func TestSocB3CommandersInsightWithNoCommanderCastsDrawsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	effectstest.CastFromHandForTest(t, g, me, "Commander's Insight", "Instant", "{X}{U}{U}{U}", socB3InsightOracle,
		game.CastSpellParams{XValue: 3, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3 (X=3, no commander cast)", got)
	}
}

// --- Conspiracy Theorist ---

func TestSocB3TheoristLootsOnAttackAndOffersTheDiscardedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theorist := effectstest.B12Push(g, me.ID, "Conspiracy Theorist", "Creature — Human Shaman", socB3TheoristOracle, 2, 2)
	// An instant, so it can be cast in combat, where the loot happens.
	spell := effectstest.BanonHandCard(me, "Pitched Instant", "Instant")
	effectstest.SeedLibrary(me, "Drawn")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, theorist)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the attack did not offer to pay {1}")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerDiscard(t, g, me.ID, spell)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d (discard one, draw one)", me.Hand.Size(), hand)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, spell) {
		t.Fatalf("the discarded sorcery is not offered: %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, spell)
	if !socB3InExile(g, spell) {
		t.Fatal("the chosen card was not exiled")
	}
	if !effectstest.EnumeratedFor(g, me.ID, spell) {
		t.Error("the exiled card can't be cast this turn")
	}
}

func TestSocB3TheoristTriggersOncePerBatchAndOnlyOnNonlandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	theorist := effectstest.B12Push(g, me.ID, "Conspiracy Theorist", "Creature — Human Shaman", socB3TheoristOracle, 2, 2)
	land := effectstest.BanonHandCard(me, "Mountain", "Basic Land — Mountain")
	g.WithWriteLock(func() {
		if err := g.DiscardCardsKeyedForEffect(me.ID, []uuid.UUID{land}, uuid.Nil, game.KeyedTail{}); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if n := effectstest.TriggersOnStackFrom(g, theorist); n != 0 {
		t.Fatalf("a land discard triggered %d times", n)
	}
	a := effectstest.BanonHandCard(me, "Spell A", "Instant")
	b := effectstest.BanonHandCard(me, "Spell B", "Sorcery")
	land2 := effectstest.BanonHandCard(me, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		if err := g.DiscardCardsKeyedForEffect(me.ID, []uuid.UUID{a, b, land2}, uuid.Nil, game.KeyedTail{}); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	prompts := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
			prompts++
		}
	}
	if prompts != 1 {
		t.Fatalf("three cards discarded together asked %d times, want 1", prompts)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if !effectstest.HasID(pick.ChooseCards, a) || !effectstest.HasID(pick.ChooseCards, b) || effectstest.HasID(pick.ChooseCards, land2) || effectstest.HasID(pick.ChooseCards, land) {
		t.Errorf("offered %v, want only the two nonland cards of this batch", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	if socB3InExile(g, a) || socB3InExile(g, b) {
		t.Error("declining exiled a card")
	}
}

// --- Creakwood Liege ---

func TestSocB3CreakwoodLiegePumpsBlackAndGreenAndMakesAWorm(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	liege := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Creakwood Liege", TypeLine: "Creature — Horror", OracleID: socB3LiegeOracle,
		Colors: []string{"B", "G"}, Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID,
	})
	black := effectstest.B16Creature(g, owner.ID, "Black", "Creature — Zombie", 2, 2, "B")
	golgari := effectstest.B16Creature(g, owner.ID, "Golgari", "Creature — Elf", 2, 2, "B", "G")
	red := effectstest.B16Creature(g, owner.ID, "Red", "Creature — Goblin", 2, 2, "R")
	if effectstest.EffectivePower(t, g, black) != 3 || effectstest.EffectivePower(t, g, golgari) != 4 || effectstest.EffectivePower(t, g, red) != 2 {
		t.Errorf("black %d, golgari %d, red %d; want 3, 4, 2",
			effectstest.EffectivePower(t, g, black), effectstest.EffectivePower(t, g, golgari), effectstest.EffectivePower(t, g, red))
	}
	if effectstest.EffectivePower(t, g, liege) != 2 {
		t.Error("the Liege pumps itself")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, owner.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	worm := effectstest.FindBattlefieldByName(g, "Worm")
	if worm == uuid.Nil {
		t.Fatal("no Worm token")
	}
	if p, tg := effectstest.EffectivePower(t, g, worm), effectstest.EffectiveToughness(t, g, worm); p != 3 || tg != 3 {
		t.Errorf("Worm is %d/%d, want 3/3 (black and green)", p, tg)
	}
}

// --- Creative Technique ---

func TestSocB3CreativeTechniqueExilesTheFirstNonlandAndOffersItFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	lands := socB3Library(me, 4, "Basic Land — Mountain", "")
	spell := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: spell, Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{6}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.CastFromHandForTest(t, g, me, "Creative Technique", "Sorcery", "{4}{R}", socB3TechniqueOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// Demonstrate asks first; decline it.
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, spell)
	if !socB3InExile(g, spell) {
		t.Fatal("the nonland card is not in exile")
	}
	for _, id := range lands {
		if !me.Library.Contains(id) {
			t.Error("a revealed land left the library")
		}
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, spell)
}

func TestSocB3CreativeTechniqueDemonstratesForYouAndAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	socB3Library(me, 3, "Sorcery", "{2}")
	socB3Library(opp, 3, "Sorcery", "{2}")
	effectstest.CastFromHandForTest(t, g, me, "Creative Technique", "Sorcery", "{4}{R}", socB3TechniqueOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	copies := map[uuid.UUID]int{}
	for _, it := range g.StackMeta {
		if it != nil && it.IsCopy && it.Kind == game.StackItemSpell {
			copies[it.Controller]++
		}
	}
	if copies[me.ID] != 1 || copies[opp.ID] != 1 || len(copies) != 2 {
		t.Fatalf("copies by controller %v, want one for you and one for the chosen opponent", copies)
	}
	// The opponent's copy resolves first, over the opponent's library.
	for i := 0; i < 8 && effectstest.ResolutionCastOfferFor(g, opp.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	offer := effectstest.ResolutionCastOfferFor(g, opp.ID)
	if offer == nil {
		t.Fatal("the opponent's copy offered them nothing")
	}
	for _, id := range offer.DiscardOptions {
		if c, ok := g.LookupCardForEffect(id); !ok || c.Owner != opp.ID {
			t.Error("the opponent's copy walked another player's library")
		}
	}
}

func TestSocB3CreativeTechniqueWithNoNonlandCardCastsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	lands := socB3Library(me, 3, "Basic Land — Mountain", "")
	effectstest.CastFromHandForTest(t, g, me, "Creative Technique", "Sorcery", "{4}{R}", socB3TechniqueOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a library of lands offered a cast")
	}
	for _, id := range lands {
		if !me.Library.Contains(id) {
			t.Error("a land left the library")
		}
	}
}

// --- Curiosity Crafter ---

func TestSocB3CuriosityCrafterDrawsForEachTokenThatConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Curiosity Crafter", "Creature — Bird Wizard", socB3CrafterOracle, 3, 3)
	t1 := effectstest.B12Creature(g, me.ID, "Spirit", "Token Creature — Spirit", 1, 1)
	t2 := effectstest.B12Creature(g, me.ID, "Spirit Two", "Token Creature — Spirit", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, t1, t2, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (two tokens connected; the bear is not a token)", got)
	}
	spec, _ := effects.Lookup(socB3CrafterOracle)
	if !spec.NoMaxHandSize {
		t.Error("no maximum hand size is not declared")
	}
}

// --- Dance with Calamity ---

func TestSocB3DanceWithCalamityCastsFreeAtThirteenOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	socB3Library(me, 6, "Sorcery", "{3}")
	effectstest.CastFromHandForTest(t, g, me, "Dance with Calamity", "Sorcery", "{7}{R}", socB3DanceOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4; i++ {
		effectstest.AnswerConfirmFor(t, g, me, true)
	}
	if me.Library.Size() != 2 {
		t.Fatalf("library %d after four exiles, want 2", me.Library.Size())
	}
	effectstest.AnswerConfirmFor(t, g, me, false)
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast)
	if len(offer.DiscardOptions) != 4 {
		t.Fatalf("offered %d cards, want the four exiled (total 12)", len(offer.DiscardOptions))
	}
	first := offer.DiscardOptions[0]
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, first, game.CastSpellParams{}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	// The other three are still offered; declining leaves them in exile.
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.CastInsideTheResolution(t, g, first)
	for _, id := range offer.DiscardOptions[1:] {
		if !socB3InExile(g, id) {
			t.Error("an uncast card left exile")
		}
	}
}

func TestSocB3DanceWithCalamityOverThirteenCastsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ids := socB3Library(me, 6, "Sorcery", "{3}")
	effectstest.CastFromHandForTest(t, g, me, "Dance with Calamity", "Sorcery", "{7}{R}", socB3DanceOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 5; i++ {
		effectstest.AnswerConfirmFor(t, g, me, true)
	}
	effectstest.AnswerConfirmFor(t, g, me, false)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a total of 15 offered a cast")
	}
	exiled := 0
	for _, id := range ids {
		if socB3InExile(g, id) {
			exiled++
		}
	}
	if exiled != 5 {
		t.Errorf("%d cards in exile, want 5", exiled)
	}
}

// --- Deadly Brew ---

func TestSocB3DeadlyBrewEdictsAndReturnsAnotherPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.PushVanillaCreature(g, me.ID, "My Spare", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.PushVanillaCreature(g, opp.ID, "Their Spare", 2, 2)
	rock := effectstest.B17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	instant := effectstest.B17GraveyardCard(me, "Old Instant", "Instant", "{U}")

	effectstest.CastCatalogSpell(t, g, "Deadly Brew", "Sorcery", socB3BrewOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, mine)
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no return choice after you sacrificed")
	}
	if !effectstest.HasID(pick.ChooseCards, rock) {
		t.Error("a permanent card in your graveyard is not offered")
	}
	if effectstest.HasID(pick.ChooseCards, instant) {
		t.Error("a nonpermanent card is offered")
	}
	if effectstest.HasID(pick.ChooseCards, mine) {
		t.Error("the permanent you sacrificed is offered; the text says another")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	if !me.Hand.Contains(rock) {
		t.Error("the chosen card did not return to your hand")
	}
}

func TestSocB3DeadlyBrewReturnsNothingIfYouSacrificedNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.B17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	effectstest.CastCatalogSpell(t, g, "Deadly Brew", "Sorcery", socB3BrewOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's only creature survived the edict")
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a return was offered though you sacrificed nothing")
	}
}

// --- Decisive Denial ---

func TestSocB3DecisiveDenialFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.Pr7Creature(g, me.ID, "Mine", 4, "G")
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 1, "B")
	if err := effectstest.RfCastModal(t, g, "Decisive Denial", socB3DenialOracle, []int{0}, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, theirs); got != 4 && g.Battlefield.Contains(theirs) {
		t.Errorf("their creature took %d, want 4", got)
	}
	if got := effectstest.Pr6Marked(g, mine); got != 1 {
		t.Errorf("your creature took %d, want 1", got)
	}
}

func TestSocB3DecisiveDenialCountersANoncreatureSpellUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	sorcery := effectstest.HandCardFull(me, "Some Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast sorcery: %v", err)
	}
	if err := effectstest.RfCastModal(t, g, "Decisive Denial", socB3DenialOracle, []int{1}, effectstest.RfCardTarget(sorcery)); err != nil {
		t.Fatalf("cast Decisive Denial: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(sorcery) {
		t.Error("declining the {3} should counter the spell")
	}
}

func TestSocB3DecisiveDenialCantCounterACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.HobA7BearSpell(t, g, me)
	if err := effectstest.RfCastModal(t, g, "Decisive Denial", socB3DenialOracle, []int{1}, effectstest.RfCardTarget(bear)); err == nil {
		t.Error("a creature spell was accepted as the counter's target")
	}
}
