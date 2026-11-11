package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s22_cards_test.go — the "draw-matters, hand size and flash
// grants" Commander batch (#5368). Each test plays the real catalog entry.

const (
	s22TriskOracle     = "ef238ba8-acf0-4e62-9abd-5291886a32e8"
	s22TishanaOracle   = "479dd04b-0048-446a-9202-a4dc9472e5c3"
	s22KefnetOracle    = "caf626a8-7abc-496b-9d3a-23194ce70e01"
	s22HoofprintOracle = "17739f55-486b-46ea-8785-70cc2bd2d872"
	s22JolraelOracle   = "f8640899-2458-40f2-98a3-35c402ee9a7a"
	s22KnowledgeOracle = "8806ee05-10d5-4492-b672-2f6bcab94be0"
	s22AlandraOracle   = "97bb1589-2941-4a75-a5f6-916871e51e9e"
	s22KianneOracle    = "7fc9e800-da66-4dda-a56d-79daec9e943e"
	s22PhantomOracle   = "f07db0f7-7d25-4395-9af3-6b0ca9197b61"
	s22SphinxOracle    = "e6284fb3-7cf8-4730-b156-10085b70b0e8"
	s22ProgSphxOracle  = "92c470c9-1373-4f3a-9805-5e9613874793"
)

func TestS22EveryBuiltCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s22TriskOracle:     "Triskaidekaphile",
		s22SphinxOracle:    "Sphinx of Enlightenment",
		s22ProgSphxOracle:  "Prognostic Sphinx",
		s22TishanaOracle:   "Tishana, Voice of Thunder",
		s22KefnetOracle:    "Kefnet the Mindful",
		s22HoofprintOracle: "Hoofprints of the Stag",
		s22JolraelOracle:   "Jolrael, Mwonvuli Recluse",
		s22KnowledgeOracle: "Knowledge Is Power",
		s22AlandraOracle:   "Alandra, Sky Dreamer",
		s22KianneOracle:    "Kianne, Corrupted Memory",
		s22PhantomOracle:   "Final-Word Phantom",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// s22FillHand sets a seat's hand to exactly n filler cards.
func s22FillHand(p *game.Player, n int) {
	for p.Hand.Size() > 0 {
		p.Hand.Remove(p.Hand.Cards[0].InstanceID)
	}
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
}

// --- Triskaidekaphile ---

func TestS22TriskaidekaphileWinsWithExactlyThirteenAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushPermanentForTest(g, me.ID, "Triskaidekaphile", s22TriskOracle, "Creature — Human Wizard")
	s22FillHand(me, 13)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if g.State != game.StateEnded || g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Fatalf("state %s outcome %+v, want a win for the controller", g.State, g.Outcome)
	}
}

func TestS22TriskaidekaphileDoesNotWinAtTwelveOrFourteen(t *testing.T) {
	for _, n := range []int{12, 14} {
		g := newCatalogGame(t)
		me := g.Seats[1]
		pushPermanentForTest(g, me.ID, "Triskaidekaphile", s22TriskOracle, "Creature — Human Wizard")
		s22FillHand(me, n)
		advanceToUpkeepOf(t, g, 1)
		passPriorityAroundTable(t, g)
		if g.State == game.StateEnded {
			t.Errorf("hand of %d won the game", n)
		}
	}
}

func TestS22TriskaidekaphileInterveningIfRechecksOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushPermanentForTest(g, me.ID, "Triskaidekaphile", s22TriskOracle, "Creature — Human Wizard")
	s22FillHand(me, 13)
	advanceToUpkeepOf(t, g, 1)
	// Triggered as the upkeep began; a card leaves the hand before it resolves.
	s22FillHand(me, 12)
	passPriorityAroundTable(t, g)
	if g.State == game.StateEnded {
		t.Fatal("won with twelve cards in hand as the trigger resolved")
	}
}

func TestS22TriskaidekaphileHasNoMaximumHandSizeAndDrawsForFourMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	phile := pushCatalogPermanent(g, me.ID, "Triskaidekaphile", "Creature — Human Wizard", s22TriskOracle, false)
	if got := g.EffectiveMaxHandSizeLocked(me); got != game.NoMaxHandSize {
		t.Errorf("max hand size %d, want no limit", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, phile, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without mana")
	}
	stdB23Mana(t, g, me, "{U}{U}{U}{U}")
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, phile, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), before+1)
	}
}

// --- Sphinx of Enlightenment ---

func TestS22SphinxOfEnlightenmentDrawsForTargetOpponentAndYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp == me {
		opp = g.Seats[2]
	}
	meBefore, oppBefore := me.Hand.Size(), opp.Hand.Size()
	castCreatureCard(t, g, "Sphinx of Enlightenment", "Creature — Sphinx", s22SphinxOracle, 5, 5)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != oppBefore+1 {
		t.Errorf("opponent hand %d, want %d", opp.Hand.Size(), oppBefore+1)
	}
	if me.Hand.Size() != meBefore+3 {
		t.Errorf("my hand %d, want %d (the Sphinx left it, three drawn)", me.Hand.Size(), meBefore+3)
	}
}

// --- Prognostic Sphinx ---

func TestS22PrognosticSphinxDiscardGivesHexproofAndTapsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sphinx := pushCatalogPermanent(g, me.ID, "Prognostic Sphinx", "Creature — Sphinx", s22ProgSphxOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, sphinx, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without discarding a card")
	}
	card := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: card, Name: "Filler", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, sphinx, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{card}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Contains(card) {
		t.Fatal("the discard was not paid")
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, sphinx, "hexproof") {
		t.Error("no hexproof")
	}
	if !dskRT2Tapped(g, sphinx) {
		t.Error("not tapped")
	}
}

func TestS22PrognosticSphinxScriesThreeWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	seedLibrary(me, "One", "Two", "Three", "Four")
	sphinx := pushCatalogPermanent(g, me.ID, "Prognostic Sphinx", "Creature — Sphinx", s22ProgSphxOracle, false)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sphinx, them.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 3 {
		t.Fatalf("scry 3 offered %d cards", len(c.ScryCards))
	}
}

// --- Tishana, Voice of Thunder ---

func TestS22TishanaSizeFollowsTheHandAndDrawsPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	s22FillHand(me, 2) // a card in hand keeps her alive: she is X/X
	tishana := castCreatureCard(t, g, "Tishana, Voice of Thunder", "Legendary Creature — Merfolk Shaman", s22TishanaOracle, 0, 0)
	if got := g.EffectiveMaxHandSizeLocked(me); got != game.NoMaxHandSize {
		t.Errorf("max hand size %d, want no limit", got)
	}
	passPriorityAroundTable(t, g)
	// Two Bears and Tishana: three cards drawn on top of the two held.
	if me.Hand.Size() != 5 {
		t.Fatalf("hand %d, want 5 (a card per creature you control)", me.Hand.Size())
	}
	pt := game.Characteristic{}
	pt.Power, pt.Toughness = effectivePTOf(t, g, tishana)
	if pt.Power != 5 || pt.Toughness != 5 {
		t.Errorf("Tishana is %d/%d with five cards in hand, want 5/5", pt.Power, pt.Toughness)
	}
}

// --- Kefnet the Mindful ---

func TestS22KefnetAttacksAndBlocksOnlyWithSevenCardsInHand(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	kefnet := pushTestPermanent(g, me.ID, game.Card{Name: "Kefnet the Mindful", OracleID: s22KefnetOracle,
		TypeLine: "Legendary Creature — God", Power: 5, Toughness: 5})
	s22FillHand(me, 6)
	advanceToDeclareAttackersOf(t, g, 0)
	wantAttackRefused(t, g, kefnet, them.ID, "fewer than seven cards in hand")
	s22FillHand(me, 7)
	if err := declareResult(g, kefnet, them.ID); err != nil {
		t.Errorf("attack with seven cards: %v", err)
	}
}

func TestS22KefnetBlocksOnlyWithSevenCardsInHand(t *testing.T) {
	g := newCatalogGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, attacker.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	kefnet := b12Push(g, defender.ID, "Kefnet the Mindful", "Legendary Creature — God", s22KefnetOracle, 5, 5)
	s22FillHand(defender, 6)
	b12Creature(g, defender.ID, "Elf", "Creature — Elf", 1, 1)
	brAttack(t, g, bear)
	brRefusal(t, g.Clone().DeclareBlocker(kefnet, bear), game.BlockReasonCantBlockAttacker)
	defender.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", Owner: defender.ID, Controller: defender.ID})
	if err := g.DeclareBlocker(kefnet, bear); err != nil {
		t.Errorf("block with seven cards in hand: %v", err)
	}
}

func TestS22KefnetDrawsThenMayReturnALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	kefnet := pushCatalogPermanent(g, me.ID, "Kefnet the Mindful", "Legendary Creature — God", s22KefnetOracle, false)
	land := pushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	stdB23Mana(t, g, me, "{U}{U}{U}{U}")
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, kefnet, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Fatalf("hand %d, want %d", me.Hand.Size(), before+1)
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatalf("no land choice: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1", pick.ChooseMin, pick.ChooseMax)
	}
	if !hasID(pick.ChooseCards, land) || hasID(pick.ChooseCards, rock) || hasID(pick.ChooseCards, kefnet) {
		t.Errorf("candidates %v, want only the land", pick.ChooseCards)
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) || onBattlefield(g, land) {
		t.Error("the land did not return to hand")
	}
}

// --- draw-matters cards ---

// s22Draw has `p` draw n cards (the library gets filler), letting any
// trigger the draws set off resolve after each.
func s22Draw(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		pushLibraryCardForTest(p, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: p.ID, Controller: p.ID})
		if err := g.DrawCard(p.ID); err != nil {
			t.Fatalf("draw %d: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
}

func s22TurnDraws(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = g.TurnTallyFor(p.ID).CardsDrawn })
	return n
}

// --- Knowledge Is Power ---

func TestS22KnowledgeIsPowerPumpsByCardsDrawnThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	pushCatalogPermanent(g, me.ID, "Knowledge Is Power", "Enchantment", s22KnowledgeOracle, false)
	base := s22TurnDraws(g, me)
	s22Draw(t, g, me, 2)
	want := 2 + base + 2
	if p, tough := effectivePTOf(t, g, mine); p != want || tough != want {
		t.Errorf("my Bear is %d/%d after %d draws this turn, want %d/%d", p, tough, base+2, want, want)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature is %d power, want 2", p)
	}
	s22Draw(t, g, opp, 1)
	if p, _ := effectivePTOf(t, g, mine); p != want {
		t.Errorf("an opponent's draw moved my Bear to %d", p)
	}
}

// --- Jolrael, Mwonvuli Recluse ---

func TestS22JolraelMakesACatOnTheSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Jolrael, Mwonvuli Recluse", "Legendary Creature — Human Druid", s22JolraelOracle, false)
	base := s22TurnDraws(g, me)
	if base > 1 {
		t.Skipf("the turn is already at %d draws", base)
	}
	s22Draw(t, g, me, 2-base-1)
	if n, _ := countTokensNamed(g, me.ID, "Cat"); n != 0 {
		t.Fatalf("%d Cats after %d draws", n, 2-base-1+base)
	}
	s22Draw(t, g, me, 1)
	if n, _ := countTokensNamed(g, me.ID, "Cat"); n != 1 {
		t.Fatalf("%d Cats after the second draw, want 1", n)
	}
	s22Draw(t, g, me, 1)
	if n, _ := countTokensNamed(g, me.ID, "Cat"); n != 1 {
		t.Errorf("%d Cats after the third draw, want still 1", n)
	}
}

func TestS22JolraelSetsBasePTToCardsInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	jolrael := pushCatalogPermanent(g, me.ID, "Jolrael, Mwonvuli Recluse", "Legendary Creature — Human Druid", s22JolraelOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	s22FillHand(me, 4)
	stdB23Mana(t, g, me, "{G}{G}{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, jolrael, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Jolrael": jolrael, "Bear": bear} {
		if p, tough := effectivePTOf(t, g, id); p != 4 || tough != 4 {
			t.Errorf("%s is %d/%d, want 4/4", name, p, tough)
		}
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature is %d power, want 2", p)
	}
}

// --- Alandra, Sky Dreamer ---

func TestS22AlandraDrakeOnSecondDrawAndPumpOnFifth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	alandra := pushCatalogPermanent(g, me.ID, "Alandra, Sky Dreamer", "Legendary Creature — Merfolk Wizard", s22AlandraOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	s22FillHand(me, 0)
	drawTo := func(n int) {
		t.Helper()
		if d := n - s22TurnDraws(g, me); d > 0 {
			s22Draw(t, g, me, d)
		}
	}
	if base := s22TurnDraws(g, me); base > 1 {
		t.Fatalf("the turn is already at %d draws", base)
	}
	drawTo(1)
	if n, _ := countTokensNamed(g, me.ID, "Drake"); n != 0 {
		t.Fatalf("%d Drakes after one draw", n)
	}
	drawTo(2)
	if n, _ := countTokensNamed(g, me.ID, "Drake"); n != 1 {
		t.Fatalf("%d Drakes after the second draw, want 1", n)
	}
	drawTo(4)
	if p, _ := effectivePTOf(t, g, alandra); p != 1 {
		t.Fatalf("Alandra is %d power before the fifth draw, want 1 (the test stand-in is 1/1)", p)
	}
	drawTo(5)
	x := me.Hand.Size()
	if x < 4 {
		t.Fatalf("hand %d, want at least 4", x)
	}
	if p, tough := effectivePTOf(t, g, alandra); p != 1+x || tough != 1+x {
		t.Errorf("Alandra is %d/%d, want %d/%d", p, tough, 1+x, 1+x)
	}
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Errorf("a non-Drake is %d power, want 2", p)
	}
	drake := tokensNamed(g, "Drake")
	if len(drake) != 1 {
		t.Fatalf("%d Drakes", len(drake))
	}
	if p, tough := effectivePTOf(t, g, drake[0].InstanceID); p != 2+x || tough != 2+x {
		t.Errorf("the Drake is %d/%d, want %d/%d", p, tough, 2+x, 2+x)
	}
	drawTo(6)
	if p, _ := effectivePTOf(t, g, alandra); p != 1+x {
		t.Errorf("a sixth draw changed Alandra to %d power", p)
	}
}

// --- Hoofprints of the Stag ---

func TestS22HoofprintsCountersAndElemental(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	stag := pushCatalogPermanent(g, me.ID, "Hoofprints of the Stag", "Kindred Enchantment — Elemental", s22HoofprintOracle, false)
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			// Declining puts no counter.
			answerLatestTriggerPrompt(t, g, me.ID, false)
			passPriorityAroundTable(t, g)
			if got := counterCount(g, stag, "hoofprint"); got != 0 {
				t.Fatalf("declined, still %d counters", got)
			}
			continue
		}
		answerLatestTriggerPrompt(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
	}
	if got := counterCount(g, stag, "hoofprint"); got != 3 {
		t.Fatalf("%d hoofprint counters, want 3", got)
	}
	stdB23Mana(t, g, me, "{W}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, stag, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with three counters")
	}
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, stag, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with four counters: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, stag, "hoofprint"); got != 0 {
		t.Errorf("%d counters left, want 0", got)
	}
	toks := tokensNamed(g, "Elemental")
	if len(toks) != 1 || toks[0].Power != 4 || toks[0].Toughness != 4 || !hasEffectiveKeyword(t, g, toks[0].InstanceID, "flying") {
		t.Errorf("tokens %+v, want one 4/4 flying Elemental", toks)
	}
}

func TestS22HoofprintsActivatesOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMainOf(t, g, 1)
	stag := pushCatalogPermanent(g, opp.ID, "Hoofprints of the Stag", "Kindred Enchantment — Elemental", s22HoofprintOracle, false)
	_ = stag
	me := g.Seats[0]
	stag2 := pushCatalogPermanent(g, me.ID, "Hoofprints of the Stag", "Kindred Enchantment — Elemental", s22HoofprintOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(stag2, "hoofprint", 4) })
	stdB23Mana(t, g, me, "{W}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, stag2, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated on an opponent's turn")
	}
}

// --- Kianne, Corrupted Memory ---

func s22TimingOpen(g *game.Game, p *game.Player, typeLine string) bool {
	var open bool
	g.WithWriteLock(func() {
		open = g.CastTimingOpenLocked(p.ID, game.Card{
			InstanceID: uuid.New(), Name: "Timing Probe", TypeLine: typeLine,
			Owner: p.ID, Controller: p.ID,
		}, game.ZoneHand, nil)
	})
	return open
}

func TestS22KianneFlashFollowsThePowerParityAndGrowsOnDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepBeginCombat)
	kianne := pushCatalogPermanent(g, me.ID, "Kianne, Corrupted Memory", "Legendary Creature — Illusion", s22KianneOracle, false)
	g.WithWriteLock(func() { g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Power = 2 })
	if p, _ := effectivePTOf(t, g, kianne); p != 2 {
		t.Fatalf("Kianne is %d power, want 2", p)
	}
	if !s22TimingOpen(g, me, "Sorcery") || !s22TimingOpen(g, me, "Enchantment") {
		t.Error("even power: noncreature spells should have flash")
	}
	if s22TimingOpen(g, me, "Creature — Bear") {
		t.Error("even power: creature spells should not have flash")
	}
	s22Draw(t, g, me, 1)
	if got := counterCount(g, kianne, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after a draw, want 1", got)
	}
	if s22TimingOpen(g, me, "Sorcery") {
		t.Error("odd power: noncreature spells should not have flash")
	}
	if !s22TimingOpen(g, me, "Creature — Bear") {
		t.Error("odd power: creature spells should have flash")
	}
	if s22TimingOpen(g, g.Seats[1], "Creature — Bear") {
		t.Error("an opponent got Kianne's flash")
	}
	s22Draw(t, g, me, 1)
	if !s22TimingOpen(g, me, "Sorcery") || s22TimingOpen(g, me, "Creature — Bear") {
		t.Error("even power again: noncreature only")
	}
}

// --- Final-Word Phantom ---

func TestS22FinalWordPhantomFlashOnlyInOpponentsEndSteps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Final-Word Phantom", "Creature — Spirit Detective", s22PhantomOracle, false)
	advanceToEndStepOf(t, g, 1)
	if !s22TimingOpen(g, me, "Sorcery") || !s22TimingOpen(g, me, "Creature — Bear") {
		t.Error("in an opponent's end step spells should have flash")
	}
	if s22TimingOpen(g, g.Seats[2], "Sorcery") {
		t.Error("another player got the Phantom's flash")
	}
	advanceToMainOf(t, g, 1)
	if s22TimingOpen(g, me, "Sorcery") {
		t.Error("in an opponent's main phase a sorcery should still be sorcery speed")
	}
	advanceToEndStepOf(t, g, 0)
	if s22TimingOpen(g, me, "Sorcery") {
		t.Error("in your own end step the Phantom grants nothing")
	}
}
