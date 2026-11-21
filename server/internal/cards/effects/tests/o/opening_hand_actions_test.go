package testso

import (
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsg "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/g"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The offer is made when the window closes, to the seat holding the
// card and to nobody else; saying yes puts the Leyline onto the
// battlefield with no mana paid and its static in force.
func TestLeylineIsOfferedAtTheCloseOfTheMulliganWindowAndBeginsOnTheBattlefield(t *testing.T) {
	g := effectstest.OpeningTable(t)
	ids := effectstest.HandWith(g, 2, effectstest.LeylineOfSanctity(), effectstest.FillerCard("Spare"))
	for _, s := range []int{0, 1, 3} {
		effectstest.HandWith(g, s, effectstest.FillerCard("Nothing special"))
	}

	// Mid-window: three seats have kept, the fourth has not. Nothing yet.
	for _, s := range []int{0, 1, 2} {
		if err := g.KeepHand(g.Seats[s].ID); err != nil {
			t.Fatalf("KeepHand seat %d: %v", s, err)
		}
	}
	if got := effectstest.ChoicesOf(g, game.PendingChoiceConfirm); len(got) != 0 {
		t.Fatalf("offered before the window closed: %d prompts", len(got))
	}
	if err := g.KeepHand(g.Seats[3].ID); err != nil {
		t.Fatalf("KeepHand seat 3: %v", err)
	}

	offers := effectstest.ChoicesOf(g, game.PendingChoiceConfirm)
	if len(offers) != 1 || offers[0].Chooser != g.Seats[2].ID {
		t.Fatalf("offers = %+v, want exactly one, to seat 2", offers)
	}
	if offers[0].Source != ids[0] {
		t.Errorf("the prompt's source is %s, want the Leyline %s", offers[0].Source, ids[0])
	}
	// The table is parked on it, as on any open choice.
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("PassPriority with the offer open: err = %v, want ErrChoicePending", err)
	}
	// Only the asked seat may answer.
	if err := g.ResolveConfirm(offers[0].ID, g.Seats[1].ID, true); !errors.Is(err, game.ErrNotTheChooser) {
		t.Errorf("another seat answered: err = %v", err)
	}

	if err := g.ResolveConfirm(offers[0].ID, g.Seats[2].ID, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	if !effectstest.BeganOnBattlefield(g, ids[0]) || g.Seats[2].Hand.Contains(ids[0]) {
		t.Fatal("the Leyline did not begin the game on the battlefield")
	}
	if !g.Seats[2].Hand.Contains(ids[1]) {
		t.Error("the Leyline took another card with it")
	}
	if len(g.Seats[2].ManaPool) != 0 {
		t.Error("beginning the game with it cost mana")
	}
	if len(effectstest.ChoicesOf(g, game.PendingChoiceConfirm)) != 0 || len(g.PendingChoices) != 0 {
		t.Errorf("prompts left over: %+v", g.PendingChoices)
	}
	if err := g.PassPriority(); errors.Is(err, game.ErrChoicePending) {
		t.Error("the table is still parked after the only offer was answered")
	}
}

// "You may": a no leaves the card in hand to be drawn and cast like any
// other, and nothing else changes.
func TestLeylineDeclinedStaysInHand(t *testing.T) {
	g := effectstest.OpeningTable(t)
	ids := effectstest.HandWith(g, 1, effectstest.LeylineOfSanctity())
	effectstest.KeepAll(t, g)
	offers := effectstest.ChoicesOf(g, game.PendingChoiceConfirm)
	if len(offers) != 1 {
		t.Fatalf("offers = %d, want 1", len(offers))
	}
	if err := g.ResolveConfirm(offers[0].ID, g.Seats[1].ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if effectstest.BeganOnBattlefield(g, ids[0]) || !g.Seats[1].Hand.Contains(ids[0]) {
		t.Error("declining moved the card")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("prompts left over: %+v", g.PendingChoices)
	}
}

// A game with no such card in any opening hand asks nothing — the
// baseline every other test in the tree relies on.
func TestNoOfferWithoutAnOpeningHandCard(t *testing.T) {
	g := effectstest.OpeningTable(t)
	for s := range g.Seats {
		effectstest.HandWith(g, s, effectstest.FillerCard("Plain"))
	}
	effectstest.KeepAll(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("prompts with no opening-hand card in play: %+v", g.PendingChoices)
	}
}

// CR 103.6: in turn order from the starting player, and one prompt per
// qualifying card in hand order.
func TestOpeningHandOffersFollowTurnOrderFromTheStartingPlayer(t *testing.T) {
	g := effectstest.OpeningTable(t)
	g.StartingSeat = 2 // the starting player is whoever the roll found
	effectstest.HandWith(g, 0, effectstest.LeylineOfSanctity())
	effectstest.HandWith(g, 1, effectstest.LeylineOfSanctity(), game.Card{Name: "Leyline of the Void", TypeLine: "Enchantment", ManaCost: "{2}{B}{B}", OracleID: effectstest.LeylineOfTheVoidOracle})
	effectstest.HandWith(g, 2, effectstest.FillerCard("Plain"))
	effectstest.HandWith(g, 3, game.Card{Name: "Leyline of Anticipation", TypeLine: "Enchantment", ManaCost: "{2}{U}{U}", OracleID: effectstest.LeylineOfAnticipationOracle})
	effectstest.KeepAll(t, g)
	var got []uuid.UUID
	for _, c := range effectstest.ChoicesOf(g, game.PendingChoiceConfirm) {
		got = append(got, c.Chooser)
	}
	want := []uuid.UUID{g.Seats[3].ID, g.Seats[0].ID, g.Seats[1].ID, g.Seats[1].ID}
	if len(got) != len(want) {
		t.Fatalf("choosers = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("offer %d is for seat %v, want %v (turn order from seat 2)", i, got[i], want[i])
		}
	}
}

// Gemstone Caverns: the starting player is never asked.
func TestGemstoneCavernsIsNotOfferedToTheStartingPlayer(t *testing.T) {
	g := effectstest.OpeningTable(t)
	effectstest.HandWith(g, 0, effectstest.GemstoneCaverns(), effectstest.FillerCard("Spare"))
	effectstest.KeepAll(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("the starting player was asked: %+v", g.PendingChoices)
	}
}

// Gemstone Caverns, accepted: it begins the game with a luck counter,
// then its owner exiles a card from what is left in hand. The luck
// counter turns its {C} into a five-colour pick.
func TestGemstoneCavernsBeginsWithALuckCounterAndExilesACard(t *testing.T) {
	g := effectstest.OpeningTable(t)
	ids := effectstest.HandWith(g, 3, effectstest.GemstoneCaverns(), effectstest.FillerCard("Keep"), effectstest.FillerCard("Pitch"))
	effectstest.KeepAll(t, g)
	offers := effectstest.ChoicesOf(g, game.PendingChoiceConfirm)
	if len(offers) != 1 || offers[0].Chooser != g.Seats[3].ID {
		t.Fatalf("offers = %+v, want one, to seat 3", offers)
	}
	if err := g.ResolveConfirm(offers[0].ID, g.Seats[3].ID, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	if !effectstest.BeganOnBattlefield(g, ids[0]) {
		t.Fatal("the Caverns did not begin on the battlefield")
	}
	if c, ok := cardByIDOK(g, ids[0]); !ok || c.Counters["luck"] != 1 {
		t.Fatalf("luck counters = %v, want 1", c.Counters)
	}

	// "If you do, exile a card from your hand": a pick among what is left.
	picks := effectstest.ChoicesOf(g, game.PendingChoiceChooseCards)
	if len(picks) != 1 || picks[0].Chooser != g.Seats[3].ID {
		t.Fatalf("exile prompts = %+v, want one, to seat 3", picks)
	}
	cands := map[uuid.UUID]bool{}
	for _, id := range picks[0].ChooseCards {
		cands[id] = true
	}
	if cands[ids[0]] || !cands[ids[1]] || !cands[ids[2]] || len(cands) != 2 {
		t.Errorf("candidates = %v, want exactly the two other cards in hand", cands)
	}
	// It is a real pick: nothing, or the Caverns itself, is refused.
	if err := g.ResolveChooseCards(picks[0].ID, g.Seats[3].ID, nil); err == nil {
		t.Error("exiling nothing was accepted")
	}
	if err := g.ResolveChooseCards(picks[0].ID, g.Seats[3].ID, []uuid.UUID{ids[0]}); err == nil {
		t.Error("exiling the Caverns itself was accepted")
	}
	if err := g.ResolveChooseCards(picks[0].ID, g.Seats[3].ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("exile pick: %v", err)
	}
	if !g.Exile.Contains(ids[2]) || g.Seats[3].Hand.Contains(ids[2]) {
		t.Error("the picked card is not in exile")
	}
	if !g.Seats[3].Hand.Contains(ids[1]) {
		t.Error("the card that was not picked left the hand")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("prompts left over: %+v", g.PendingChoices)
	}

	// With the counter it makes any colour, and it is the engine's own
	// mana ability that says so.
	if got := cardsg.GemstoneCavernsProduced(g, g.Seats[3].ID, ids[0]); got != "{W|U|B|R|G}" {
		t.Errorf("produces %q with a luck counter, want the five-colour pick", got)
	}
}

// A Caverns alone in hand has nothing to exile, so nothing is asked.
func TestGemstoneCavernsWithNothingElseInHandExilesNothing(t *testing.T) {
	g := effectstest.OpeningTable(t)
	ids := effectstest.HandWith(g, 2, effectstest.GemstoneCaverns())
	effectstest.KeepAll(t, g)
	offers := effectstest.ChoicesOf(g, game.PendingChoiceConfirm)
	if len(offers) != 1 {
		t.Fatalf("offers = %d, want 1", len(offers))
	}
	if err := g.ResolveConfirm(offers[0].ID, g.Seats[2].ID, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	if !effectstest.BeganOnBattlefield(g, ids[0]) {
		t.Fatal("the Caverns did not begin on the battlefield")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("an empty hand was asked to exile: %+v", g.PendingChoices)
	}
}

// Declined, a Caverns is the ordinary land it prints elsewhere: no
// counter, {C}, and no exile.
func TestGemstoneCavernsDeclinedIsAnOrdinaryLand(t *testing.T) {
	g := effectstest.OpeningTable(t)
	ids := effectstest.HandWith(g, 1, effectstest.GemstoneCaverns(), effectstest.FillerCard("Spare"))
	effectstest.KeepAll(t, g)
	offers := effectstest.ChoicesOf(g, game.PendingChoiceConfirm)
	if err := g.ResolveConfirm(offers[0].ID, g.Seats[1].ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if !g.Seats[1].Hand.Contains(ids[0]) || g.Exile.Size() != 0 || len(g.PendingChoices) != 0 {
		t.Error("declining changed the board")
	}
	if got := cardsg.GemstoneCavernsProduced(g, g.Seats[1].ID, ids[0]); got != "{C}" {
		t.Errorf("a Caverns with no luck counter produces %q, want {C}", got)
	}
}

// A seat that leaves while its offer is open must not wedge the table:
// the prompt is dropped and the game goes on.
func TestADepartedSeatsOfferDoesNotWedgeTheTable(t *testing.T) {
	g := effectstest.OpeningTable(t)
	effectstest.HandWith(g, 1, effectstest.LeylineOfSanctity())
	effectstest.KeepAll(t, g)
	if len(effectstest.ChoicesOf(g, game.PendingChoiceConfirm)) != 1 {
		t.Fatal("setup: expected the offer")
	}
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("the departed seat's offer is still open: %+v", g.PendingChoices)
	}
	if err := g.PassPriority(); errors.Is(err, game.ErrChoicePending) {
		t.Error("the table is parked on a prompt nobody can answer")
	}
}

// The roll-first start finds its own starting player; the offers follow
// whoever it found (CR 103.6), including "you're not the starting
// player" on Gemstone Caverns.
func TestOpeningHandOffersFollowTheRolledStartingPlayer(t *testing.T) {
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.StartWithFirstPlayerRoll(rand.New(rand.NewPCG(5, 6))); err != nil {
		t.Fatal(err)
	}
	for s := range g.Seats {
		effectstest.HandWith(g, s, effectstest.GemstoneCaverns(), effectstest.FillerCard("Spare"))
	}
	effectstest.KeepAll(t, g)
	offers := effectstest.ChoicesOf(g, game.PendingChoiceConfirm)
	if len(offers) != 3 {
		t.Fatalf("offers = %d, want one for each seat but the starting player", len(offers))
	}
	for _, o := range offers {
		if o.Chooser == g.Seats[g.StartingSeat].ID {
			t.Errorf("the starting player (seat %d) was asked", g.StartingSeat)
		}
	}
}

// The declarations the engine reads: the Leylines and Gemstone Caverns
// are complete now, and the riders are the ones printed.
func TestOpeningHandDeclarations(t *testing.T) {
	for _, oracle := range []string{
		effectstest.LeylineOfSanctityOracle, effectstest.LeylineOfAnticipationOracle, effectstest.LeylineOfTheVoidOracle,
		"997478aa-b790-4269-a626-abf0cb30fea0", // Leyline of Lifeforce
		"2608df54-dfbe-417d-aef5-49afbdfb03da", // Leyline of Punishment
		"caab67eb-65e7-4755-b116-6977e97f0844", // Leyline of Mutation
		"4f597675-0f6d-438c-990c-337171927a5e", // Leyline Axe
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.OpeningHand == nil || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: opening-hand action %v, completeness %v", oracle, spec.OpeningHand, spec.Completeness)
			continue
		}
		if a := spec.OpeningHand; a.NotStartingPlayer || len(a.EntersWithCounters) != 0 || a.ExileFromHand != 0 {
			t.Errorf("%s: a plain Leyline clause carries riders: %+v", oracle, a)
		}
	}
	spec, _ := effects.Lookup(effectstest.GemstoneCavernsOracle)
	a := spec.OpeningHand
	if a == nil || !a.NotStartingPlayer || a.EntersWithCounters["luck"] != 1 || a.ExileFromHand != 1 || spec.Completeness != effects.CompletenessFull {
		t.Errorf("Gemstone Caverns: %+v completeness %v", a, spec.Completeness)
	}
}

func TestRegisterRefusesAMalformedOpeningHandAction(t *testing.T) {
	for name, a := range map[string]*game.OpeningHandAction{
		"a counter with no count":  {EntersWithCounters: map[string]int{"luck": 0}},
		"a counter with no name":   {EntersWithCounters: map[string]int{"": 1}},
		"a negative exile count":   {ExileFromHand: -1},
		"a negative counter count": {EntersWithCounters: map[string]int{"luck": -2}},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("Register accepted it")
				}
			}()
			effects.CheckOpeningHand("Test Card", a)
		})
	}
}

func cardByIDOK(g *game.Game, id uuid.UUID) (game.Card, bool) {
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() { out, ok = g.LookupCardForEffect(id) })
	return out, ok
}
