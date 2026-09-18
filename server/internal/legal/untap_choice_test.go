package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// untap_choice_test.go — #826 / ADR 0070, the enumerator half.
//
// A seat owing a choice is offered that choice's answers and NOTHING
// else, so a kind the enumerator cannot answer is a bot asleep holding
// the whole table (#544, #499). The untap prompt is the first one that
// can open with no priority anywhere in the game, which makes that
// failure mode worse: nobody could even pass out of it.
//
// The rule pinned here is the search / choose-cards one. Every offered
// answer is an answer the resolver accepts, because the enumerator
// asks the engine (ChooseCardsPickLegalLocked) rather than enumerating
// a superset past the cap solver it cannot see.

// untapCapsOverLands stubs a Winter Orb: "no more than n lands",
// keyed on an oracle ID the test puts on a battlefield card.
func untapCapsOverLands(t *testing.T, key string, n int) {
	t.Helper()
	prev := game.CatalogUntapCaps
	game.CatalogUntapCaps = func(k string) []game.UntapCap {
		if k != key {
			return nil
		}
		return []game.UntapCap{{
			Max:    n,
			Label:  "no more than n lands",
			Counts: func(_ *game.Game, _ *game.Card, target *game.Card) bool { return target.IsLand() },
		}}
	}
	t.Cleanup(func() { game.CatalogUntapCaps = prev })
}

func pushTappedLegalPermanent(g *game.Game, owner uuid.UUID, name, oracleID, typeLine string, tapped bool) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		OracleID:   oracleID,
		TypeLine:   typeLine,
		Owner:      owner,
		Controller: owner,
		Tapped:     tapped,
	})
	return id
}

// openUntapPrompt drives the cursor to the first untap prompt.
func openUntapPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 60; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceUntapChoice {
				return c
			}
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward the untap prompt: %v", err)
		}
	}
	t.Fatal("the untap step never asked")
	return nil
}

// Every answer the enumerator offers is one the dispatcher accepts,
// and the cap solver's refusals never reach the move list.
func TestUntapChoiceOffersOnlyAnswersTheEngineAccepts(t *testing.T) {
	g := newTable(t)
	const orb = "test-winter-orb"
	untapCapsOverLands(t, orb, 1)
	seat := g.Seats[1]
	pushTappedLegalPermanent(g, seat.ID, "Orb", orb, "Artifact", false)
	pushTappedLegalPermanent(g, seat.ID, "Forest", "", "Land", true)
	pushTappedLegalPermanent(g, seat.ID, "Island", "", "Land", true)
	pushTappedLegalPermanent(g, seat.ID, "Swamp", "", "Land", true)

	choice := openUntapPrompt(t, g)
	moves := legal.EnumerateFor(g, seat.ID)
	if len(moves) == 0 {
		t.Fatal("the seat owing the untap determination was offered nothing")
	}
	dispatchAll(t, g, seat.ID, moves)
	for _, m := range moves {
		if m.AlwaysLegal {
			t.Errorf("a prompt with a floor of one marked %q always-legal", m.Label)
		}
		ids := pickedCardIDs(t, m)
		if len(ids) != 1 {
			t.Errorf("offered %q (%d cards), want exactly one land", m.Label, len(ids))
		}
		var ok bool
		g.ReadSnapshot(func() { ok = g.ChooseCardsPickLegalLocked(choice, ids) })
		if !ok {
			t.Errorf("offered %q, which the engine refuses", m.Label)
		}
	}
	if len(moves) != 3 {
		t.Errorf("offered %d answers, want one per land: %v", len(moves), labels(moves))
	}
}

// Nothing else is offered while the determination is owed: the untap
// step grants no priority, so a seat that could be offered a pass here
// would be passing out of a step nobody holds.
func TestUntapChoiceIsTheOnlyThingOffered(t *testing.T) {
	g := newTable(t)
	const orb = "test-winter-orb"
	untapCapsOverLands(t, orb, 1)
	seat := g.Seats[1]
	pushTappedLegalPermanent(g, seat.ID, "Orb", orb, "Artifact", false)
	pushTappedLegalPermanent(g, seat.ID, "Forest", "", "Land", true)
	pushTappedLegalPermanent(g, seat.ID, "Island", "", "Land", true)
	openUntapPrompt(t, g)

	for _, m := range legal.EnumerateFor(g, seat.ID) {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("offered %q (%s) while the untap determination is owed", m.Label, m.Type)
		}
	}
	// And no other seat is offered anything at all, because the prompt
	// blocks the table.
	for _, other := range []*game.Player{g.Seats[0], g.Seats[2], g.Seats[3]} {
		if moves := legal.EnumerateFor(g, other.ID); len(moves) != 0 {
			t.Errorf("seat %s was offered %v while an untap prompt blocks the table",
				other.Name, labels(moves))
		}
	}
}

// The opt-out board: the floor is zero, so "untap nothing" is offered
// and is the answer that always terminates.
func TestUntapChoiceOffersTheEmptyAnswerWhenEverythingIsOptional(t *testing.T) {
	g := newTable(t)
	const tick = "test-rust-tick"
	prev := game.CatalogUntapOptOuts
	game.CatalogUntapOptOuts = func(k string) []game.UntapOptOut {
		if k != tick {
			return nil
		}
		return []game.UntapOptOut{{
			Optional: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
		}}
	}
	t.Cleanup(func() { game.CatalogUntapOptOuts = prev })
	seat := g.Seats[1]
	pushTappedLegalPermanent(g, seat.ID, "Tick", tick, "Artifact Creature", true)
	openUntapPrompt(t, g)

	moves := legal.EnumerateFor(g, seat.ID)
	dispatchAll(t, g, seat.ID, moves)
	always := 0
	for _, m := range moves {
		if m.AlwaysLegal {
			always++
			if ids := pickedCardIDs(t, m); len(ids) != 0 {
				t.Errorf("the always-legal answer names %d cards", len(ids))
			}
		}
	}
	if always != 1 {
		t.Errorf("%d always-legal answers, want exactly the empty one: %v", always, labels(moves))
	}
}
