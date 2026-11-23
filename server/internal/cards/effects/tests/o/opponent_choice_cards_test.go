package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestOpponentChoiceCardsAreRegistered(t *testing.T) {
	for oracleID, name := range map[string]string{
		effectstest.FactOrFictionOracle:       "Fact or Fiction",
		effectstest.TormentOfHailfireOracle:   "Torment of Hailfire",
		effectstest.CombustibleGearhulkOracle: "Combustible Gearhulk",
		effectstest.CharismaticConquerorOracl: "Charismatic Conqueror",
		effectstest.PainfulQuandaryOracle:     "Painful Quandary",
	} {
		spec, ok := effects.Lookup(oracleID)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracleID)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracleID, spec.Name)
		}
	}
}

// TestFactOrFictionSplitsAndTakesAPile is the end-to-end card: the
// reveal, the opponent's split, the controller's pick, and the two
// piles landing in the right zones.
func TestFactOrFictionSplitsAndTakesAPile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealed := effectstest.SeedFactOrFictionLibrary(me, "FF One", "FF Two", "FF Three", "FF Four", "FF Five")
	hand, graves := me.Hand.Size(), me.Graveyard.Size()

	effectstest.CastCatalogSpell(t, g, "Fact or Fiction", "Instant", effectstest.FactOrFictionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// #929: "an opponent" is the controller's choice now.
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)

	split := effectstest.LatestRevealPickFor(g, opp.ID)
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
	if err := g.ResolveRevealPick(split.ID, opp.ID, revealed[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the controller picks a pile: %+v", g.PendingChoices)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealed := effectstest.SeedFactOrFictionLibrary(me, "FF One", "FF Two", "FF Three", "FF Four", "FF Five")
	hand, graves := me.Hand.Size(), me.Graveyard.Size()

	effectstest.CastCatalogSpell(t, g, "Fact or Fiction", "Instant", effectstest.FactOrFictionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	split := effectstest.LatestRevealPickFor(g, opp.ID)
	if split == nil {
		t.Fatalf("no split prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(split.ID, opp.ID, nil); err != nil {
		t.Fatalf("ResolveRevealPick(empty): %v", err)
	}
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("an empty split still asks which pile: %+v", g.PendingChoices)
	}
	if len(pick.PickOptions[0].Cards) != 0 || len(pick.PickOptions[1].Cards) != len(revealed) {
		t.Errorf("one empty pile and one of everything: %+v", pick.PickOptions)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealed := effectstest.SeedFactOrFictionLibrary(me, "FF One", "FF Two", "FF Three", "FF Four", "FF Five")
	hand := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Fact or Fiction", "Instant", effectstest.FactOrFictionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	split := effectstest.LatestRevealPickFor(g, opp.ID)
	if split == nil {
		t.Fatalf("no split prompt: %+v", g.PendingChoices)
	}
	snap := g.Clone()
	if err := g.ResolveRevealPick(split.ID, opp.ID, revealed[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if g.Seats[g.Turn.ActiveSeat].Hand.Size() != hand+2 {
		t.Fatal("the chain ran")
	}
	g.RestoreFrom(snap)

	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if me.Hand.Size() != hand {
		t.Errorf("undo rewinds both links: hand %d, want %d", me.Hand.Size(), hand)
	}
	restored := effectstest.LatestRevealPickFor(g, opp.ID)
	if restored == nil {
		t.Fatalf("undo puts the split back: %+v", g.PendingChoices)
	}
	// The restored prompt is a chain link whose continuation did NOT
	// survive the snapshot census — it survives the CLONE, which is
	// what undo uses — so answering it must still work end to end.
	if err := g.ResolveRevealPick(restored.ID, opp.ID, revealed[:3]); err != nil {
		t.Fatalf("ResolveRevealPick after undo: %v", err)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if me.Hand.Size() != hand+3 {
		t.Errorf("answering the restored chain: hand %d, want %d", me.Hand.Size(), hand+3)
	}
}

// --- Combustible Gearhulk ------------------------------------------

func TestCombustibleGearhulkOpponentSaysDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	gearhulk := effectstest.B12Push(g, me.ID, "Combustible Gearhulk", "Artifact Creature — Construct", effectstest.CombustibleGearhulkOracle, 6, 6)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: gearhulk}) })
	// The ETB trigger targets an opponent; the pick comes first.
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Hand.Size() != hand+3 {
		t.Errorf("the opponent let them draw three: hand %d → %d", hand, me.Hand.Size())
	}
	if opp.Life != 40 {
		t.Errorf("nobody takes damage on the draw branch: opponent at %d", opp.Life)
	}
}

func TestCombustibleGearhulkOpponentRefusesAndTakesTheBurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
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

	gearhulk := effectstest.B12Push(g, me.ID, "Combustible Gearhulk", "Artifact Creature — Construct", effectstest.CombustibleGearhulkOracle, 6, 6)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: gearhulk}) })
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Graveyard.Size() != graves+3 {
		t.Errorf("they mill three: graveyard %d → %d", graves, me.Graveyard.Size())
	}
	if opp.Life != life-6 {
		t.Errorf("damage is the total mana value of the milled cards (3+2+1): life %d → %d", life, opp.Life)
	}
}

// --- Charismatic Conqueror ------------------------------------------

func TestCharismaticConquerorOpponentTapsItsPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Charismatic Conqueror", "Creature — Vampire Soldier", effectstest.CharismaticConquerorOracl, 2, 2)
	tokens := countVampireTokens(g, me.ID)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: opp.ID, CardID: theirs}) })
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, theirs); !ok || !c.Tapped {
		t.Error("the opponent tapped their own permanent")
	}
	if countVampireTokens(g, me.ID) != tokens {
		t.Error("tapping it makes no token")
	}
}

func TestCharismaticConquerorOpponentRefusesAndGivesAVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Charismatic Conqueror", "Creature — Vampire Soldier", effectstest.CharismaticConquerorOracl, 2, 2)
	tokens := countVampireTokens(g, me.ID)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: opp.ID, CardID: theirs}) })
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AnswerMayChoice(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, theirs); !ok || c.Tapped {
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
		if c.Controller == controller && effects.IsToken(c) && c.HasSubtype("Vampire") {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	// Give exactly this opponent something to sacrifice; everyone
	// else answers with life.
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life

	effectstest.CastTormentOfHailfire(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestOptionPickFor(g, opp.ID)
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
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	sac := effectstest.LatestChooseCardsFor(g, opp.ID)
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
	effectstest.DrainRemainingTormentPrompts(t, g)
}

// TestTormentOfHailfireDropsOptionsThePlayerCannotTake — CR 608.2's
// "as much as possible", enforced at queue time: a player with no
// nonland permanent and no cards in hand is offered one answer.
func TestTormentOfHailfireDropsOptionsThePlayerCannotTake(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp.Hand.Cards = nil
	life := opp.Life

	effectstest.CastTormentOfHailfire(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("still asked: %+v", g.PendingChoices)
	}
	if len(pick.PickOptions) != 1 || pick.PickOptions[0].Label != "Lose 3 life" {
		t.Fatalf("only the life branch is possible: %+v", pick.PickOptions)
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp.Life)
	}
	effectstest.DrainRemainingTormentPrompts(t, g)
}

// --- Painful Quandary -----------------------------------------------

func TestPainfulQuandaryTaxesAnOpponentsCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	// The enchantment's controller is NOT the active player, so the
	// active player's cast is an opponent's cast.
	active := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, owner.ID, "Painful Quandary", "Enchantment", effectstest.PainfulQuandaryOracle, 0, 0)
	life, graves := active.Life, active.Graveyard.Size()

	effectstest.CastCatalogSpell(t, g, "Filler Spell", "Instant", "not-in-catalog", nil)
	effectstest.PassPriorityAroundTable(t, g)

	// The caster chooses: discard.
	effectstest.AnswerMayChoice(t, g, active.ID, true)
	discard := effectstest.LatestChooseCardsFor(g, active.ID)
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
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, owner.ID, "Painful Quandary", "Enchantment", effectstest.PainfulQuandaryOracle, 0, 0)
	life := active.Life

	effectstest.CastCatalogSpell(t, g, "Filler Spell", "Instant", "not-in-catalog", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, active.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if active.Life != life-5 {
		t.Errorf("declining loses 5 life: %d → %d", life, active.Life)
	}
}

// TestPainfulQuandaryAsksNothingOfAnEmptyHand — hellbent has no
// choice to make, so the life comes straight off rather than through a
// prompt whose only answer the engine already knows.
func TestPainfulQuandaryAsksNothingOfAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, owner.ID, "Painful Quandary", "Enchantment", effectstest.PainfulQuandaryOracle, 0, 0)

	effectstest.CastCatalogSpell(t, g, "Filler Spell", "Instant", "not-in-catalog", nil)
	life := active.Life
	active.Hand.Cards = nil
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.LatestConfirmFor(g, active.ID) != nil {
		t.Errorf("no question for an empty hand: %+v", g.PendingChoices)
	}
	if active.Life != life-5 {
		t.Errorf("the life comes off anyway: %d → %d", life, active.Life)
	}
}
