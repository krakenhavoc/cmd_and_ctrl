package decks_test

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// dragon_hoard_test.go — ADR 0172 decision 8 (#3401): the one built-in
// deck with a five-colour commander carries Jegantha, the Wellspring
// as its companion, so the companion path runs in ordinary games.
//
// CI has no Scryfall dump, so the mana costs the condition reads are
// recorded in testdata/dragon_hoard_mana_costs.json, taken from the
// dump when the deck was built. The real-dump test in
// realdump_manual_test.go re-checks the whole pipeline against
// Scryfall itself.

const dragonHoardID = "wubrg-dragons"

func dragonHoard(t *testing.T) decks.Deck {
	t.Helper()
	d, ok := decks.Lookup(dragonHoardID)
	if !ok {
		t.Fatalf("no built-in deck %q", dragonHoardID)
	}
	return d
}

func dragonHoardCosts(t *testing.T) map[string]string {
	t.Helper()
	raw, err := os.ReadFile("testdata/dragon_hoard_mana_costs.json")
	if err != nil {
		t.Fatal(err)
	}
	var costs map[string]string
	if err := json.Unmarshal(raw, &costs); err != nil {
		t.Fatal(err)
	}
	return costs
}

// dragonHoardIndex is a Scryfall index with exactly the deck's cards:
// the recorded mana cost, the declared colour identity, and the
// companion keyword on Jegantha.
func dragonHoardIndex(t *testing.T, d decks.Deck, costs map[string]string) *cards.Index {
	t.Helper()
	idx := cards.NewIndex()
	for _, c := range d.WithCompanion() {
		typeLine := "Artifact"
		var keywords []string
		switch {
		case c.Basic:
			typeLine = "Basic Land — " + c.Name
		case c.Name == d.Commander.Name:
			typeLine = "Legendary Creature — Dragon Avatar"
		case d.Companion != nil && c.Name == d.Companion.Name:
			typeLine = "Legendary Creature — Elemental Elk"
			keywords = []string{"Companion"}
		}
		identity := make([]string, 0, len(c.Identity))
		for _, r := range c.Identity {
			identity = append(identity, string(r))
		}
		card := cards.Card{
			ID:            uuid.New(),
			Name:          c.Name,
			TypeLine:      typeLine,
			ManaCost:      costs[c.Name],
			Keywords:      keywords,
			ColorIdentity: identity,
			Legalities:    map[string]string{"commander": "legal"},
		}
		if c.OracleID != "" {
			card.OracleID = uuid.MustParse(c.OracleID)
		}
		idx.Put(card)
	}
	return idx
}

// TestDragonHoardFitsItsCompanion is the proof that the list is legal
// with Jegantha beside it: Load runs ParseText, Resolve and Validate,
// so a companion_* violation or any other kind refuses the deck.
func TestDragonHoardFitsItsCompanion(t *testing.T) {
	d := dragonHoard(t)
	costs := dragonHoardCosts(t)
	if d.Companion == nil || d.Companion.Name != "Jegantha, the Wellspring" {
		t.Fatalf("companion %+v, want Jegantha, the Wellspring", d.Companion)
	}
	if d.Identity != "WUBRG" {
		t.Fatalf("identity %q: Jegantha only fits a five-colour commander", d.Identity)
	}

	list, err := d.Load(dragonHoardIndex(t, d, costs))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if list.Companion == nil || list.Companion.Name != "Jegantha, the Wellspring" {
		t.Fatalf("the loaded list has companion %+v", list.Companion)
	}
	if n := len(list.Commanders) + len(list.Mainboard); n != 100 {
		t.Errorf("%d cards in the 100, want 100 (the companion is not one of them)", n)
	}

	// A second, independent reading of the condition: no symbol twice
	// in any cost of the starting deck, as printed.
	symbol := regexp.MustCompile(`\{[^}]+\}`)
	for _, c := range d.Cards() {
		seen := map[string]bool{}
		for _, s := range symbol.FindAllString(costs[c.Name], -1) {
			if seen[s] {
				t.Errorf("%s (%s) repeats %s", c.Name, costs[c.Name], s)
			}
			seen[s] = true
		}
	}
}

// TestDragonHoardBreaksWithoutTheConditionHolding is the negative
// control: the same list with one card's cost doubled is refused for
// the condition, so the test above is not passing on a validator that
// never looks.
func TestDragonHoardBreaksWithoutTheConditionHolding(t *testing.T) {
	d := dragonHoard(t)
	costs := dragonHoardCosts(t)
	costs["Swords to Plowshares"] = "{W}{W}"
	_, err := d.Load(dragonHoardIndex(t, d, costs))
	var verr *deck.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("Load: %v, want a validation error", err)
	}
	for _, v := range verr.Violations {
		if v.Code == deck.CodeCompanionConditionNotMet && v.Card == "Swords to Plowshares" {
			return
		}
	}
	t.Fatalf("violations %+v do not name Swords to Plowshares for the companion condition", verr.Violations)
}

// TestDragonHoardSeatsItsCompanionAndOffersIt seats the deck the way
// the lobby does, with Jegantha waiting in Player.Companion, and checks
// the move list a bot is handed offers the {3} companion special action
// once the table can pay for it, and only then.
func TestDragonHoardSeatsItsCompanionAndOffersIt(t *testing.T) {
	d := dragonHoard(t)
	idx := dragonHoardIndex(t, d, dragonHoardCosts(t))

	g := game.NewGame()
	for _, name := range []string{"Hoard", "Rival"} {
		list, err := d.Load(idx)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if _, err := g.AddPlayer(name, list.ToGameCards()); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithSource(nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}

	me := g.Seats[0]
	jegantha, ok := me.CompanionCard()
	if !ok || jegantha.Name != "Jegantha, the Wellspring" || !jegantha.IsCompanion {
		t.Fatalf("seat 0 companion %+v, %v: want Jegantha waiting outside the game", jegantha, ok)
	}
	if n := me.Library.Size() + me.Hand.Size(); n != 99 {
		t.Errorf("library + hand = %d, want 99", n)
	}
	if me.Command.Size() != 1 {
		t.Errorf("%d cards in the command zone, want the commander", me.Command.Size())
	}

	companionMoves := func(seat uuid.UUID) []legal.Move {
		var out []legal.Move
		for _, m := range legal.EnumerateFor(g, seat) {
			if m.Kind == legal.KindSpecialAction && m.Source == jegantha.InstanceID &&
				strings.Contains(string(m.Params), `"companion"`) {
				out = append(out, m)
			}
		}
		return out
	}

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Seats[g.Turn.ActiveSeat].ID != me.ID {
		t.Fatalf("seat %d is active, want the companion's owner", g.Turn.ActiveSeat)
	}
	if got := companionMoves(me.ID); len(got) != 0 {
		t.Fatalf("offered with no mana: %v", got)
	}
	for range 3 {
		land := game.Card{
			InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island",
			Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		}
		g.Battlefield.PushTop(land)
	}
	if got := companionMoves(me.ID); len(got) != 1 {
		t.Fatalf("with three lands the bot is offered %d companion moves, want 1", len(got))
	}
	if got := companionMoves(g.Seats[1].ID); len(got) != 0 {
		t.Errorf("the rival is offered the owner's companion: %v", got)
	}
}

// TestDragonHoardCompanionIsFull holds the companion to the same bar
// as the 100: the catalog grades Jegantha full, so the mana it makes
// is the mana the deck plays with.
func TestDragonHoardCompanionIsFull(t *testing.T) {
	d := dragonHoard(t)
	cov := decks.CoverageOf(nil, decks.Deck{Commander: *d.Companion})
	if !cov.Complete() || cov.Full != 1 {
		t.Errorf("Jegantha is not graded full: %+v", cov)
	}
}
