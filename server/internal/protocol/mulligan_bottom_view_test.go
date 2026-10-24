package protocol

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mulligan_bottom_view_test.go — #4301, ADR 0168 (the London mulligan) and
// ADR 0133 (what a hand's owner alone may see). The pick of the cards a
// kept hand puts on the bottom of its library is a keyed choose_cards prompt
// and its order a keyed put_in_library prompt, so a table waiting on either
// is a restore point. The pick is private to its chooser (its words and its
// candidates name cards in a hand), and a restore must not undo that.

// mulliganOwingTable is a two-player table in the mulligan window where both
// seats took one mulligan (no free one in a two-player game, CR 103.5c), so
// each owes one card. It returns the seat that keeps first, now bottoming,
// and the other, still to decide.
func mulliganOwingTable(t *testing.T) (g *game.Game, keeper, other *game.Player) {
	t.Helper()
	g = buildActiveGame(t)
	if !g.MulligansOpen {
		t.Fatal("no mulligan window")
	}
	first := g.Seats[g.MulliganDecider()]
	if err := g.Mulligan(first.ID, game.OpeningHandSize); err != nil {
		t.Fatal(err)
	}
	second := g.Seats[g.MulliganDecider()]
	if err := g.Mulligan(second.ID, game.OpeningHandSize); err != nil {
		t.Fatal(err)
	}
	if err := g.KeepHand(first.ID); err != nil {
		t.Fatal(err)
	}
	return g, first, second
}

func openPromptOf(g *game.Game, kind game.PendingChoiceKind, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// restoreStrict sends g through a snapshot and back, refusing a table that
// is not a restore point.
func restoreStrict(t *testing.T, g *game.Game) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("the game is not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded game.GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	return restored
}

func seatByID(g *game.Game, id uuid.UUID) *game.Player {
	for _, p := range g.Seats {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func pendingViews(g *game.Game, viewer uuid.UUID) []PendingChoiceView {
	return ViewOfGameFor(g, viewer.String()).PendingChoices
}

// The bottom pick: a restore keeps it private to its chooser, the table's
// view of it is the same before and after, and the pick is answered on the
// restored table.
func TestLondonMulliganBottomPickStaysPrivateAcrossARestore(t *testing.T) {
	g, keeper, other := mulliganOwingTable(t)
	pick := openPromptOf(g, game.PendingChoiceChooseCards, keeper.ID)
	if pick == nil {
		t.Fatal("the keep raised no bottom pick")
	}
	if !pick.PrivateText() {
		t.Fatal("the bottom pick is not private to its chooser before the restore")
	}
	if len(pick.ChooseCards) != game.OpeningHandSize || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("the pick offers %d cards, %d to %d; want the whole hand and exactly one", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	hand := append([]uuid.UUID(nil), pick.ChooseCards...)
	keeperBefore, otherBefore := pendingViews(g, keeper.ID), pendingViews(g, other.ID)

	g = restoreStrict(t, g)
	keeper, other = seatByID(g, keeper.ID), seatByID(g, other.ID)
	pick = openPromptOf(g, game.PendingChoiceChooseCards, keeper.ID)
	if pick == nil {
		t.Fatal("the restored table lost the bottom pick")
	}
	if !pick.PrivateText() {
		t.Fatal("the restored bottom pick is no longer private to its chooser")
	}
	if !reflect.DeepEqual(pick.ChooseCards, hand) {
		t.Errorf("the restored pick offers %v, want the hand %v", pick.ChooseCards, hand)
	}

	// What each seat is shown is what it was shown before.
	keeperAfter, otherAfter := pendingViews(g, keeper.ID), pendingViews(g, other.ID)
	if !reflect.DeepEqual(keeperBefore, keeperAfter) {
		t.Errorf("the chooser's view changed across the restore:\nbefore %+v\nafter  %+v", keeperBefore, keeperAfter)
	}
	if !reflect.DeepEqual(otherBefore, otherAfter) {
		t.Errorf("the other seat's view changed across the restore:\nbefore %+v\nafter  %+v", otherBefore, otherAfter)
	}
	// And the chooser sees its hand while nobody else does.
	if len(keeperAfter) != 1 || len(keeperAfter[0].Options) != game.OpeningHandSize || keeperAfter[0].Chooser != keeper.ID.String() {
		t.Fatalf("the chooser's restored view is %+v, want its hand of %d to pick from", keeperAfter, game.OpeningHandSize)
	}
	if len(otherAfter) != 1 {
		t.Fatalf("the other seat sees %d prompts, want one: that the game is waiting on the keeper", len(otherAfter))
	}
	v := otherAfter[0]
	if !v.PrivateText || v.Chooser != keeper.ID.String() || len(v.Options) != 0 || v.ChooseMin != 0 || v.ChooseMax != 0 ||
		v.Source != "" || v.AcceptLabel != "" || v.DeclineLabel != "" {
		t.Errorf("the other seat was told what is in the keeper's hand after the restore: %+v", v)
	}
	raw, _ := json.Marshal(otherAfter)
	for _, id := range hand {
		if strings.Contains(string(raw), id.String()) {
			t.Errorf("the other seat's view names the card %s", id)
		}
	}

	// The restored pick is answered, and the order that follows is a restore
	// point too.
	bottom := hand[3]
	if err := g.ResolveChooseCards(pick.ID, keeper.ID, []uuid.UUID{bottom}); err != nil {
		t.Fatal(err)
	}
	if got := len(keeper.Hand.Cards); got != game.OpeningHandSize-1 {
		t.Fatalf("hand = %d after the pick, want %d", got, game.OpeningHandSize-1)
	}
	if !g.MulligansOpen {
		t.Fatal("the window closed with the other seat still to decide")
	}
}

// The order of two cards for the bottom is a keyed put_in_library prompt,
// and the window does not close while a restored seat is still bottoming:
// the other seat finishes first and the window stays open for the keeper.
func TestLondonMulliganWindowStaysOpenForARestoredBottomingSeat(t *testing.T) {
	g, keeper, other := twoOwedTable(t)
	pick := openPromptOf(g, game.PendingChoiceChooseCards, keeper.ID)
	if pick == nil || pick.ChooseMin != 2 {
		t.Fatalf("the keeper owes %+v, want a pick of two", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, keeper.ID, pick.ChooseCards[:2]); err != nil {
		t.Fatal(err)
	}
	order := openPromptOf(g, game.PendingChoicePutInLibrary, keeper.ID)
	if order == nil {
		t.Fatal("two cards to bottom asked for no order")
	}

	g = restoreStrict(t, g)
	keeper, other = seatByID(g, keeper.ID), seatByID(g, other.ID)
	order = openPromptOf(g, game.PendingChoicePutInLibrary, keeper.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 2 {
		t.Fatalf("the restored order prompt is %+v, want the two cards for the bottom", order)
	}

	// The other seat keeps and finishes its own bottoming while the keeper's
	// order is still open: the window must stay open for the keeper.
	if err := g.KeepHand(other.ID); err != nil {
		t.Fatal(err)
	}
	otherPick := openPromptOf(g, game.PendingChoiceChooseCards, other.ID)
	if otherPick == nil {
		t.Fatal("the second keep raised no bottom pick")
	}
	if err := g.ResolveChooseCards(otherPick.ID, other.ID, otherPick.ChooseCards[:otherPick.ChooseMin]); err != nil {
		t.Fatal(err)
	}
	if o := openPromptOf(g, game.PendingChoicePutInLibrary, other.ID); o != nil {
		if err := g.ResolvePutInLibrary(o.ID, other.ID, o.ScryCards, nil); err != nil {
			t.Fatal(err)
		}
	}
	if !g.MulligansOpen {
		t.Fatal("the window closed while the restored seat was still putting its cards on the bottom")
	}

	// The keeper's order, on the restored table, releases it.
	if err := g.ResolvePutInLibrary(order.ID, keeper.ID, order.ScryCards, nil); err != nil {
		t.Fatal(err)
	}
	if g.MulligansOpen {
		t.Error("the window is still open with every card down")
	}
	lib := keeper.Library.Cards
	if len(lib) < 2 || lib[0].InstanceID != order.ScryCards[1] || lib[1].InstanceID != order.ScryCards[0] {
		t.Error("the cards are not on the bottom in the order the restored prompt was answered")
	}
}

// twoOwedTable is mulliganOwingTable with a second mulligan for each seat,
// so the keeper owes two cards.
func twoOwedTable(t *testing.T) (g *game.Game, keeper, other *game.Player) {
	t.Helper()
	g = buildActiveGame(t)
	first := g.Seats[g.MulliganDecider()]
	for range 2 {
		for range 2 {
			p := g.Seats[g.MulliganDecider()]
			if err := g.Mulligan(p.ID, game.OpeningHandSize); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := g.KeepHand(first.ID); err != nil {
		t.Fatal(err)
	}
	for _, p := range g.Seats {
		if p.ID != first.ID {
			other = p
		}
	}
	return g, first, other
}

// A restore with the mulligan's key unknown to the binary is refused rather
// than restored as a question whose answer would run nothing.
func TestLondonMulliganBottomPickWithAnUnknownKeyIsRefused(t *testing.T) {
	g, keeper, _ := mulliganOwingTable(t)
	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PendingChoices {
		if c := &snap.PendingChoices[i]; c.Chooser == keeper.ID && c.Kind == game.PendingChoiceChooseCards {
			if c.PickThen != "choose-cards/london-mulligan-bottom" {
				t.Fatalf("the bottom pick is written under %q", c.PickThen)
			}
			c.PickThen = "choose-cards/never-registered"
			found = true
		}
	}
	if !found {
		t.Fatal("no bottom pick in the snapshot")
	}
	if _, err := snap.Restore(); err == nil {
		t.Fatal("restored a bottom pick naming a key this binary does not register")
	}
}
