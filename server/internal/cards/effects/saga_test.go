package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// saga_test.go covers the S27 Saga lifecycle end-to-end through a
// real catalog card: the entry lore counter (CR 714.3), the
// precombat-main advance, chapter dispatch in printed order, and the
// CR 704.5s sacrifice — including the half of that rule that is easy
// to get wrong, which is that the final chapter has to RESOLVE first.

const (
	historyOfBenaliaOracle  = "c15bb7eb-aaaa-4468-9641-8f706d6137e8"
	firstIroanGamesOracle   = "58934f6d-1aa2-414c-85c6-955a1e26d675"
	doublingSeasonSagaOracl = "01546b7d-a233-4176-8843-d732074dc5b6"
)

// advanceToPrecombatMainOf walks the turn engine to the named seat's
// precombat main phase — the step whose entry hook performs the
// CR 714.3 lore-counter turn-based action.
// Always advances at least one step, so calling it while already
// parked at the target seat's precombat main walks a full turn cycle
// round to the NEXT one rather than returning immediately. Every
// caller here means "the next advance", and the tests cast their Saga
// during precombat main.
func advanceToPrecombatMainOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward precombat main of seat %d: %v", seat, err)
		}
		if g.Turn.Step == game.StepPrecombatMain && g.Turn.ActiveSeat == seat {
			return
		}
	}
	t.Fatalf("never reached precombat main of seat %d", seat)
}

// loreCountersOn returns the lore-counter count on a battlefield
// card, or -1 when the card is no longer on the battlefield.
func loreCountersOn(g *game.Game, cardID uuid.UUID) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == cardID {
			return c.Counters[game.CounterLore]
		}
	}
	return -1
}

func countBattlefieldByName(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

func inGraveyardOf(g *game.Game, playerID, cardID uuid.UUID) bool {
	for _, p := range g.Seats {
		if p.ID != playerID {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.InstanceID == cardID {
				return true
			}
		}
	}
	return false
}

// TestSagaEntersWithLoreCounterAndFiresChapterOne is CR 714.3's
// first half: the counter arrives as the Saga enters, and the
// chapter it lands on triggers immediately rather than waiting for
// the next turn.
func TestSagaEntersWithLoreCounterAndFiresChapterOne(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	sagaID := castCatalogSpell(t, g, "History of Benalia", "Enchantment — Saga",
		historyOfBenaliaOracle, nil)
	passPriorityAroundTable(t, g)

	if got := loreCountersOn(g, sagaID); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}
	if got := countBattlefieldByName(g, "Knight"); got != 1 {
		t.Errorf("Knights after chapter I = %d, want 1", got)
	}
	if g.Turn.ActiveSeat != seat {
		t.Errorf("active seat drifted during the cast")
	}
}

// TestSagaAdvancesAtPrecombatMain is the second half of CR 714.3 —
// the turn-based action after the controller's draw step — plus the
// dispatch of the chapter that counter reaches.
func TestSagaAdvancesAtPrecombatMain(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	sagaID := castCatalogSpell(t, g, "History of Benalia", "Enchantment — Saga",
		historyOfBenaliaOracle, nil)
	passPriorityAroundTable(t, g)

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

	if got := loreCountersOn(g, sagaID); got != 2 {
		t.Fatalf("lore counters after one advance = %d, want 2", got)
	}
	if got := countBattlefieldByName(g, "Knight"); got != 2 {
		t.Errorf("Knights after chapters I and II = %d, want 2", got)
	}
}

// TestSagaOnlyAdvancesOnItsControllersTurn pins the "your draw step"
// half of CR 714.3: another player's precombat main must not move
// the counter.
func TestSagaOnlyAdvancesOnItsControllersTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	sagaID := castCatalogSpell(t, g, "History of Benalia", "Enchantment — Saga",
		historyOfBenaliaOracle, nil)
	passPriorityAroundTable(t, g)

	other := (seat + 1) % len(g.Seats)
	advanceToPrecombatMainOf(t, g, other)
	passPriorityAroundTable(t, g)

	if got := loreCountersOn(g, sagaID); got != 1 {
		t.Fatalf("lore counters after an OPPONENT's precombat main = %d, want 1", got)
	}
}

// TestSagaFinalChapterResolvesBeforeSacrifice is the rule everything
// else hangs off: CR 704.5s does not sacrifice a Saga whose chapter
// ability has triggered and not yet left the stack. Sacrificing early
// would silently delete the chapter the whole card exists for.
func TestSagaFinalChapterResolvesBeforeSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat].ID
	sagaID := castCatalogSpell(t, g, "History of Benalia", "Enchantment — Saga",
		historyOfBenaliaOracle, nil)
	passPriorityAroundTable(t, g)

	advanceToPrecombatMainOf(t, g, seat) // lore 2
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, seat) // lore 3 — the final chapter

	if got := loreCountersOn(g, sagaID); got != 3 {
		t.Fatalf("lore counters at the final chapter = %d, want 3", got)
	}
	if triggerOnStack(g, sagaID) == nil {
		t.Fatal("chapter III is not on the stack")
	}
	if loreCountersOn(g, sagaID) < 0 {
		t.Fatal("the Saga was sacrificed with its final chapter still on the stack")
	}

	passPriorityAroundTable(t, g)

	if loreCountersOn(g, sagaID) >= 0 {
		t.Error("the Saga survived its final chapter's resolution")
	}
	if !inGraveyardOf(g, owner, sagaID) {
		t.Error("the sacrificed Saga did not reach its owner's graveyard")
	}
	// Chapter III really ran: the two Knights are 4/3 until cleanup.
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Knight" {
			continue
		}
		if got := c.CurrentPower(); got != 4 {
			t.Errorf("Knight power after chapter III = %d, want 4", got)
		}
	}
}

// TestFourChapterSagaSurvivesChapterThree proves the final chapter is
// read off the card's declarations rather than hard-coded to III.
func TestFourChapterSagaSurvivesChapterThree(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	sagaID := castCatalogSpell(t, g, "The First Iroan Games", "Enchantment — Saga",
		firstIroanGamesOracle, nil)
	passPriorityAroundTable(t, g)

	for i := 0; i < 2; i++ {
		advanceToPrecombatMainOf(t, g, seat)
		answerAnyPendingTargetPrompts(t, g)
		passPriorityAroundTable(t, g)
	}
	if got := loreCountersOn(g, sagaID); got != 3 {
		t.Fatalf("lore counters = %d, want 3", got)
	}

	advanceToPrecombatMainOf(t, g, seat)
	answerAnyPendingTargetPrompts(t, g)
	passPriorityAroundTable(t, g)

	if loreCountersOn(g, sagaID) >= 0 {
		t.Error("a four-chapter Saga was sacrificed after chapter IV had not yet run")
	}
	if got := countBattlefieldByName(g, "Gold"); got != 1 {
		t.Errorf("Gold tokens after chapter IV = %d, want 1", got)
	}
}

// TestSagaEntersWithTwoLoreCountersUnderDoublingSeason is the
// interaction the entry counter is routed through the CR 614
// replacement pipeline FOR: "as a Saga enters, put a lore counter on
// it" is a counter-placement event, so Doubling Season doubles it and
// the Saga starts on chapter II — firing chapter I on the way past.
//
// The turn-based advance at precombat main is deliberately NOT
// routed the same way: a turn-based action is not an effect, so
// Doubling Season does nothing to it. TestSagaAdvancesAtPrecombatMain
// covers the undoubled step; this covers the doubled one.
func TestSagaEntersWithTwoLoreCountersUnderDoublingSeason(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat].ID
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Doubling Season",
		OracleID:   doublingSeasonSagaOracl,
		TypeLine:   "Enchantment",
		Owner:      owner,
		Controller: owner,
	})

	sagaID := castCatalogSpell(t, g, "History of Benalia", "Enchantment — Saga",
		historyOfBenaliaOracle, nil)
	// Chapters I and II trigger together, so CR 603.3b asks their
	// controller for a resolution order before either reaches the
	// stack. Answering it in the offered order is what the client's
	// default does.
	passPriorityUntilTriggerOrderPrompt(t, g)
	answerTriggerOrderInOfferedOrder(t, g)
	passPriorityAroundTable(t, g)

	if got := loreCountersOn(g, sagaID); got != 2 {
		t.Fatalf("lore counters under Doubling Season = %d, want 2", got)
	}
	if got := countBattlefieldByName(g, "Knight"); got != 2 {
		t.Errorf("Knights after chapters I and II fired together = %d, want 2", got)
	}
}

// TestRegisteredSagasDeclareContiguousChapters is the invariant that
// protects every Saga in the catalog at once, including ones added
// later. A gap in the chapter numbers is a typo with two silent
// consequences: the skipped chapter never fires, and — if the gap is
// at the top — the engine reads a SMALLER final chapter than the card
// prints and sacrifices the Saga before its last chapter ever
// triggers. Both are invisible in a card file and obvious here.
func TestRegisteredSagasDeclareContiguousChapters(t *testing.T) {
	for _, spec := range All() {
		seen := map[int]bool{}
		final := 0
		for _, tr := range spec.Triggered {
			if tr.Chapter == 0 {
				continue
			}
			if tr.Chapter < 0 {
				t.Errorf("%s declares chapter %d", spec.Name, tr.Chapter)
				continue
			}
			seen[tr.Chapter] = true
			if tr.Chapter > final {
				final = tr.Chapter
			}
		}
		if final == 0 {
			continue
		}
		for n := 1; n <= final; n++ {
			if !seen[n] {
				t.Errorf("%s declares chapters up to %d but is missing chapter %d",
					spec.Name, final, n)
			}
		}
	}
}

// TestTheBirthOfMeletisRunsAllThreeChaptersThenIsSacrificed is a
// second, independent walk of the whole lifecycle on a Saga whose
// chapters are all untargeted — so the assertion is about the
// chapters running, not about answering prompts.
func TestTheBirthOfMeletisRunsAllThreeChaptersThenIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	before := owner.Life

	sagaID := castCatalogSpell(t, g, "The Birth of Meletis", "Enchantment — Saga",
		"1ae54fe7-b1d3-4c13-a8ef-f502cf3eb1a0", nil)
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, sagaID); got != 1 {
		t.Fatalf("lore after entry = %d, want 1", got)
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldByName(g, "Wall"); got != 1 {
		t.Errorf("Walls after chapter II = %d, want 1", got)
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if got := owner.Life - before; got != 2 {
		t.Errorf("life gained by chapter III = %d, want 2", got)
	}
	if loreCountersOn(g, sagaID) >= 0 {
		t.Error("the Saga was not sacrificed after its final chapter resolved")
	}
}

// passPriorityUntilTriggerOrderPrompt passes priority until a
// CR 603.3b ordering prompt appears (or the stack settles). Distinct
// from passPriorityAroundTable, which fatals when a prompt stops the
// passes.
func passPriorityUntilTriggerOrderPrompt(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if triggerOrderPrompt(g) != nil || stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("no trigger_order prompt and the stack never settled")
}

func triggerOrderPrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
			return c
		}
	}
	return nil
}

func answerTriggerOrderInOfferedOrder(t *testing.T, g *game.Game) {
	t.Helper()
	ch := triggerOrderPrompt(g)
	if ch == nil {
		t.Fatal("no trigger_order prompt to answer")
	}
	if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, ch.TriggerOrderIDs); err != nil {
		t.Fatalf("ResolveTriggerOrder: %v", err)
	}
}

// answerAnyPendingTargetPrompts clicks through the pick_target
// prompts a targeted chapter queues, choosing the first legal option.
// The First Iroan Games' chapter II targets a creature you control,
// and its own Human Soldier is the only one on an empty board.
func answerAnyPendingTargetPrompts(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		var ch *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePickTarget {
				ch = c
				break
			}
		}
		if ch == nil {
			return
		}
		if len(ch.PickTargetCards) == 0 {
			t.Fatalf("pick_target prompt %q offered no card options", ch.Reason)
		}
		pick := game.TargetRef{Kind: game.TargetCard, ID: ch.PickTargetCards[0]}
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, pick); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
}
