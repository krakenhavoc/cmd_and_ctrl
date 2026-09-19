package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponent_choice_cards_test.go — #568. The five cards the
// opponent-facing resolution choice unblocked.

const (
	factOrFictionOracle       = "437b2dab-15e0-4b9a-a204-58622d37a3b3"
	tormentOfHailfireOracle   = "30a0a6c2-1fbb-4784-ab96-22611d57e62c"
	combustibleGearhulkOracle = "3494a575-f68e-4e78-9f6b-142cd8a0edea"
	charismaticConquerorOracl = "3dcc35f3-74dc-46a8-8aa9-3411189f0547"
	painfulQuandaryOracle     = "c37051cc-6683-4dbb-b5ff-5c3a5bdab1df"
)

func TestOpponentChoiceCardsAreRegistered(t *testing.T) {
	for oracleID, name := range map[string]string{
		factOrFictionOracle:       "Fact or Fiction",
		tormentOfHailfireOracle:   "Torment of Hailfire",
		combustibleGearhulkOracle: "Combustible Gearhulk",
		charismaticConquerorOracl: "Charismatic Conqueror",
		painfulQuandaryOracle:     "Painful Quandary",
	} {
		spec, ok := Lookup(oracleID)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracleID)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracleID, spec.Name)
		}
	}
}

// --- Fact or Fiction ------------------------------------------------

// seedFactOrFictionLibrary puts five named cards on top of a player's
// library, top card LAST in the returned slice's reading order — the
// reveal reports them top-first, so the helper returns them in reveal
// order.
func seedFactOrFictionLibrary(p *game.Player, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(names))
	for _, name := range names {
		id := uuid.New()
		p.Library.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Artifact",
			Owner: p.ID, Controller: p.ID,
		})
		ids = append(ids, id)
	}
	// PushTop leaves the LAST pushed on top, and the reveal reads
	// top-first, so reverse.
	out := make([]uuid.UUID, 0, len(ids))
	for i := len(ids) - 1; i >= 0; i-- {
		out = append(out, ids[i])
	}
	return out
}

// TestFactOrFictionSplitsAndTakesAPile is the end-to-end card: the
// reveal, the opponent's split, the controller's pick, and the two
// piles landing in the right zones.
func TestFactOrFictionSplitsAndTakesAPile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealed := seedFactOrFictionLibrary(me, "FF One", "FF Two", "FF Three", "FF Four", "FF Five")
	hand, graves := me.Hand.Size(), me.Graveyard.Size()

	castCatalogSpell(t, g, "Fact or Fiction", "Instant", factOrFictionOracle, nil)
	passPriorityAroundTable(t, g)
	// #929: "an opponent" is the controller's choice now.
	answerChoosePlayer(t, g, me.ID, opp)

	split := latestChooseCardsFor(g, opp.ID)
	if split == nil {
		t.Fatalf("the opponent separates the piles: %+v", g.PendingChoices)
	}
	if len(split.ChooseCards) != 5 {
		t.Fatalf("five cards to separate, got %d", len(split.ChooseCards))
	}
	// The cards are revealed, so the splitter is entitled to see them
	// — that is what the redaction pass checks, and what makes the
	// prompt answerable at all.
	for _, id := range revealed {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !c.KnownBy[opp.ID] {
			t.Fatalf("the reveal makes %s public to the splitter", id)
		}
	}
	if err := g.ResolveChooseCards(split.ID, opp.ID, revealed[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	pick := latestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the controller picks a pile: %+v", g.PendingChoices)
	}
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != hand+2 {
		t.Errorf("the taken pile goes to hand: %d → %d, want +2", hand, me.Hand.Size())
	}
	// The other three, plus Fact or Fiction itself.
	if me.Graveyard.Size() != graves+4 {
		t.Errorf("the other pile goes to the graveyard: %d → %d, want +4 (3 cards + the spell)", graves, me.Graveyard.Size())
	}
	for _, id := range revealed[:2] {
		if !me.Hand.Contains(id) {
			t.Errorf("%s should be in hand", id)
		}
	}
	for _, id := range revealed[2:] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("%s should be in the graveyard", id)
		}
	}
}

// TestFactOrFictionAcceptsAnEmptyPile — "two piles", one of which may
// be empty, is a legal and often correct split, and the controller can
// take the empty one.
func TestFactOrFictionAcceptsAnEmptyPile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealed := seedFactOrFictionLibrary(me, "FF One", "FF Two", "FF Three", "FF Four", "FF Five")
	hand, graves := me.Hand.Size(), me.Graveyard.Size()

	castCatalogSpell(t, g, "Fact or Fiction", "Instant", factOrFictionOracle, nil)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)
	split := latestChooseCardsFor(g, opp.ID)
	if split == nil {
		t.Fatalf("no split prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(split.ID, opp.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards(empty): %v", err)
	}
	pick := latestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("an empty split still asks which pile: %+v", g.PendingChoices)
	}
	if len(pick.PickOptions[0].Cards) != 0 || len(pick.PickOptions[1].Cards) != len(revealed) {
		t.Errorf("one empty pile and one of everything: %+v", pick.PickOptions)
	}
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != hand+5 {
		t.Errorf("taking the non-empty pile: hand %d → %d, want +5", hand, me.Hand.Size())
	}
	if me.Graveyard.Size() != graves+1 {
		t.Errorf("only the spell itself is binned: graveyard %d → %d", graves, me.Graveyard.Size())
	}
}

// TestFactOrFictionUndoesAcrossBothPrompts — the whole two-link chain
// rewinds, which is the claim the continuations make.
func TestFactOrFictionUndoesAcrossBothPrompts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealed := seedFactOrFictionLibrary(me, "FF One", "FF Two", "FF Three", "FF Four", "FF Five")
	hand := me.Hand.Size()

	castCatalogSpell(t, g, "Fact or Fiction", "Instant", factOrFictionOracle, nil)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)
	split := latestChooseCardsFor(g, opp.ID)
	if split == nil {
		t.Fatalf("no split prompt: %+v", g.PendingChoices)
	}
	snap := g.Clone()
	if err := g.ResolveChooseCards(split.ID, opp.ID, revealed[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	answerOptionPick(t, g, me.ID, 0)
	if g.Seats[g.Turn.ActiveSeat].Hand.Size() != hand+2 {
		t.Fatal("the chain ran")
	}
	g.RestoreFrom(snap)

	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if me.Hand.Size() != hand {
		t.Errorf("undo rewinds both links: hand %d, want %d", me.Hand.Size(), hand)
	}
	restored := latestChooseCardsFor(g, opp.ID)
	if restored == nil {
		t.Fatalf("undo puts the split back: %+v", g.PendingChoices)
	}
	// The restored prompt is a chain link whose continuation did NOT
	// survive the snapshot census — it survives the CLONE, which is
	// what undo uses — so answering it must still work end to end.
	if err := g.ResolveChooseCards(restored.ID, opp.ID, revealed[:3]); err != nil {
		t.Fatalf("ResolveChooseCards after undo: %v", err)
	}
	answerOptionPick(t, g, me.ID, 0)
	if me.Hand.Size() != hand+3 {
		t.Errorf("answering the restored chain: hand %d, want %d", me.Hand.Size(), hand+3)
	}
}

// --- Combustible Gearhulk ------------------------------------------

func TestCombustibleGearhulkOpponentSaysDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	gearhulk := b12Push(g, me.ID, "Combustible Gearhulk", "Artifact Creature — Construct", combustibleGearhulkOracle, 6, 6)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: gearhulk}) })
	// The ETB trigger targets an opponent; the pick comes first.
	b04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, opp.ID, true)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != hand+3 {
		t.Errorf("the opponent let them draw three: hand %d → %d", hand, me.Hand.Size())
	}
	if opp.Life != 40 {
		t.Errorf("nobody takes damage on the draw branch: opponent at %d", opp.Life)
	}
}

func TestCombustibleGearhulkOpponentRefusesAndTakesTheBurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Three known mana values on top of the library: 3 + 2 + 1 = 6 is
	// the damage the "doesn't" branch deals.
	for _, cost := range []string{"{3}", "{2}", "{1}"} {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Cost " + cost, TypeLine: "Artifact",
			ManaCost: cost, Owner: me.ID, Controller: me.ID,
		})
	}
	life, graves := opp.Life, me.Graveyard.Size()

	gearhulk := b12Push(g, me.ID, "Combustible Gearhulk", "Artifact Creature — Construct", combustibleGearhulkOracle, 6, 6)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: gearhulk}) })
	b04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)

	if me.Graveyard.Size() != graves+3 {
		t.Errorf("they mill three: graveyard %d → %d", graves, me.Graveyard.Size())
	}
	if opp.Life != life-6 {
		t.Errorf("damage is the total mana value of the milled cards (3+2+1): life %d → %d", life, opp.Life)
	}
}

// --- Charismatic Conqueror ------------------------------------------

func TestCharismaticConquerorOpponentTapsItsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Charismatic Conqueror", "Creature — Vampire Soldier", charismaticConquerorOracl, 2, 2)
	tokens := countVampireTokens(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: opp.ID, CardID: theirs}) })
	passPriorityAroundTable(t, g)

	answerMayChoice(t, g, opp.ID, true)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, theirs); !ok || !c.Tapped {
		t.Error("the opponent tapped their own permanent")
	}
	if countVampireTokens(g, me.ID) != tokens {
		t.Error("tapping it makes no token")
	}
}

func TestCharismaticConquerorOpponentRefusesAndGivesAVampire(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Charismatic Conqueror", "Creature — Vampire Soldier", charismaticConquerorOracl, 2, 2)
	tokens := countVampireTokens(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: opp.ID, CardID: theirs}) })
	passPriorityAroundTable(t, g)

	answerMayChoice(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, theirs); !ok || c.Tapped {
		t.Error("declining taps nothing")
	}
	if got := countVampireTokens(g, me.ID); got != tokens+1 {
		t.Errorf("the Conqueror's controller gets a Vampire: %d → %d", tokens, got)
	}
}

// countVampireTokens counts a player's Vampire tokens.
func countVampireTokens(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && IsToken(c) && c.HasSubtype("Vampire") {
			n++
		}
	}
	return n
}

// --- Torment of Hailfire --------------------------------------------

// TestTormentOfHailfireOffersThreeOptionsAndRepeats covers the option
// list being built from what the player can do, the three branches,
// and the repetitions being asked one at a time.
func TestTormentOfHailfireOffersThreeOptionsAndRepeats(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	// Give exactly this opponent something to sacrifice; everyone
	// else answers with life.
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life

	castTormentOfHailfire(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)

	pick := latestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("each opponent is asked: %+v", g.PendingChoices)
	}
	if len(pick.PickOptions) != 3 {
		t.Fatalf("lose life / sacrifice / discard, got %+v", pick.PickOptions)
	}
	if pick.PickOptions[0].LifeCost != 3 {
		t.Errorf("the life branch declares its price: %+v", pick.PickOptions[0])
	}
	// Sacrifice: a chained card pick over their nonland permanents.
	answerOptionPick(t, g, opp.ID, 1)
	sac := latestChooseCardsFor(g, opp.ID)
	if sac == nil {
		t.Fatalf("the sacrifice branch asks which permanent: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(sac.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the chosen permanent is sacrificed")
	}
	if opp.Life != life {
		t.Error("sacrificing instead of losing life costs no life")
	}
	// Everyone else answers with the life loss; the run then ends.
	drainRemainingTormentPrompts(t, g)
}

// TestTormentOfHailfireDropsOptionsThePlayerCannotTake — CR 608.2's
// "as much as possible", enforced at queue time: a player with no
// nonland permanent and no cards in hand is offered one answer.
func TestTormentOfHailfireDropsOptionsThePlayerCannotTake(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp.Hand.Cards = nil
	life := opp.Life

	castTormentOfHailfire(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)

	pick := latestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("still asked: %+v", g.PendingChoices)
	}
	if len(pick.PickOptions) != 1 || pick.PickOptions[0].Label != "Lose 3 life" {
		t.Fatalf("only the life branch is possible: %+v", pick.PickOptions)
	}
	answerOptionPick(t, g, opp.ID, 0)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp.Life)
	}
	drainRemainingTormentPrompts(t, g)
}

// castTormentOfHailfire casts the sorcery for the given X.
func castTormentOfHailfire(t *testing.T, g *game.Game, caster uuid.UUID, x int) {
	t.Helper()
	p := g.PlayerByIDForEffect(caster)
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Torment of Hailfire", TypeLine: "Sorcery",
		OracleID: tormentOfHailfireOracle, Owner: caster, Controller: caster,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(caster, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("CastSpell Torment of Hailfire: %v", err)
	}
}

// drainRemainingTormentPrompts answers every option pick still open
// with its always-legal first branch, so the chain runs to the end and
// the table is left clean.
func drainRemainingTormentPrompts(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionPick {
				open = c
				break
			}
		}
		if open == nil {
			return
		}
		if err := g.ResolveOptionPick(open.ID, open.Chooser, 0); err != nil {
			t.Fatalf("ResolveOptionPick: %v", err)
		}
	}
	t.Fatal("the Torment chain did not terminate")
}

// --- Painful Quandary -----------------------------------------------

func TestPainfulQuandaryTaxesAnOpponentsCast(t *testing.T) {
	g := newCatalogGame(t)
	// The enchantment's controller is NOT the active player, so the
	// active player's cast is an opponent's cast.
	active := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, owner.ID, "Painful Quandary", "Enchantment", painfulQuandaryOracle, 0, 0)
	life, graves := active.Life, active.Graveyard.Size()

	castCatalogSpell(t, g, "Filler Spell", "Instant", "not-in-catalog", nil)
	passPriorityAroundTable(t, g)

	// The caster chooses: discard.
	answerMayChoice(t, g, active.ID, true)
	discard := latestChooseCardsFor(g, active.ID)
	if discard == nil {
		t.Fatalf("the discard branch asks which card: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(discard.ID, active.ID, discard.ChooseCards[:1]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if active.Life != life {
		t.Errorf("discarding costs no life: %d → %d", life, active.Life)
	}
	if active.Graveyard.Size() != graves+1 {
		t.Error("a card was discarded")
	}
}

func TestPainfulQuandaryTakesFiveWhenTheCasterDeclines(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, owner.ID, "Painful Quandary", "Enchantment", painfulQuandaryOracle, 0, 0)
	life := active.Life

	castCatalogSpell(t, g, "Filler Spell", "Instant", "not-in-catalog", nil)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, active.ID, false)
	passPriorityAroundTable(t, g)

	if active.Life != life-5 {
		t.Errorf("declining loses 5 life: %d → %d", life, active.Life)
	}
}

// TestPainfulQuandaryAsksNothingOfAnEmptyHand — hellbent has no
// choice to make, so the life comes straight off rather than through a
// prompt whose only answer the engine already knows.
func TestPainfulQuandaryAsksNothingOfAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, owner.ID, "Painful Quandary", "Enchantment", painfulQuandaryOracle, 0, 0)

	castCatalogSpell(t, g, "Filler Spell", "Instant", "not-in-catalog", nil)
	life := active.Life
	active.Hand.Cards = nil
	passPriorityAroundTable(t, g)

	if latestConfirmFor(g, active.ID) != nil {
		t.Errorf("no question for an empty hand: %+v", g.PendingChoices)
	}
	if active.Life != life-5 {
		t.Errorf("the life comes off anyway: %d → %d", life, active.Life)
	}
}
