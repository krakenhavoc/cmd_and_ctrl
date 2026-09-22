package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// entry_reveal_test.go — #1198 / ADR 0013 §5z, the enumerator half.
//
// A seat owing a choice is offered that choice's answers and NOTHING
// else, so a kind the enumerator cannot answer is a bot asleep
// holding the whole table (#544, #499). This prompt makes that
// failure mode as bad as it gets: it opens with a permanent HALFWAY
// onto the battlefield and the CR 614 pipeline suspended on the
// answer, so a seat with no move leaves the table unable to reach the
// board state it is already in.
//
// Two things are pinned. Every offered answer is one the dispatcher
// accepts (dispatchAll, through the real resolve_choice route), and
// the DECLINE is always there and always marked AlwaysLegal — which
// is the property that makes this prompt un-wedgeable, and the one
// choose_cards with a floor of one does not have.

const entryRevealOracle = "test-reveal-land"

// stubRevealLandCatalog wires a one-card catalog whose land carries
// the reveal clause: "as this land enters, you may reveal a Swamp
// card from your hand. If you don't, it enters tapped."
func stubRevealLandCatalog(t *testing.T) {
	t.Helper()
	def := &game.CardDef{
		Replacements: []game.ReplacementEffect{{
			Watches:         []game.EventKind{game.EventZoneMove},
			SelfReplacement: true,
			Label:           "Test Reveal Land",
			PromptQuestion:  "Test Reveal Land — reveal a Swamp card from your hand?",
			EntryHandReveal: &game.EntryHandReveal{
				Matches: func(c game.Card) bool { return c.IsLand() && c.HasSubtype("swamp") },
				Max:     1,
			},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove &&
					ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Controller: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				if ev != nil && ev.Actor != uuid.Nil {
					return ev.Actor
				}
				return uuid.Nil
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
		}},
	}
	prev := game.CatalogLookup
	game.CatalogLookup = func(key string) *game.CardDef {
		if key == entryRevealOracle {
			return def
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogLookup = prev })
}

// openEntryRevealPrompt plays the stubbed reveal-land from seat 0's
// hand and returns the prompt it pauses on.
func openEntryRevealPrompt(t *testing.T, g *game.Game, hand ...game.Card) *game.PendingChoice {
	t.Helper()
	me := g.Seats[0]
	clearHand(me)
	for _, c := range hand {
		handCard(me, c)
	}
	land := handCard(me, game.Card{
		Name:     "Test Reveal Land",
		TypeLine: "Land",
		OracleID: entryRevealOracle,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{}); err != nil {
		t.Fatalf("play the reveal-land: %v", err)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceEntryRevealFromHand {
			return c
		}
	}
	t.Fatal("playing the reveal-land queued no entry_reveal_from_hand prompt")
	return nil
}

func TestEntryRevealOffersOnlyAnswersTheEngineAccepts(t *testing.T) {
	stubRevealLandCatalog(t)
	g := newTable(t)
	choice := openEntryRevealPrompt(t, g,
		basic("Swamp", "Swamp"),
		basic("Bog", "Swamp"),
		basic("Island", "Island"),
		creature("Bear", "{1}{G}", 2, 2),
	)

	me := g.Seats[0]
	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) == 0 {
		t.Fatal("the seat owing the entry reveal was offered nothing")
	}
	dispatchAll(t, g, me.ID, moves)

	declines := 0
	for _, m := range moves {
		ids := pickedCardIDs(t, m)
		if len(ids) == 0 {
			declines++
			if !m.AlwaysLegal {
				t.Errorf("the decline %q is not marked AlwaysLegal", m.Label)
			}
			continue
		}
		if m.AlwaysLegal {
			t.Errorf("a non-empty answer %q is marked AlwaysLegal", m.Label)
		}
		var ok bool
		g.ReadSnapshot(func() { ok = g.ChooseCardsPickLegalLocked(choice, ids) })
		if !ok {
			t.Errorf("offered %q, which the engine refuses", m.Label)
		}
	}
	if declines != 1 {
		t.Errorf("%d decline answers, want exactly 1", declines)
	}
	// Two Swamps and the decline. The Island and the Bear are not
	// candidates, so the enumerator cannot reach them.
	if len(moves) != 3 {
		t.Errorf("offered %d answers, want 3 (decline + two Swamps): %v", len(moves), labels(moves))
	}
}

// TestEntryRevealAlwaysHasAnAnswer is the un-wedgeable claim, at the
// board that would wedge a choose_cards with a floor of one: the
// prompt's candidates are gone from the zone it re-checks against.
// The decline survives, because the engine skips both the zone
// re-check and Validate for an empty answer.
func TestEntryRevealAlwaysHasAnAnswer(t *testing.T) {
	stubRevealLandCatalog(t)
	g := newTable(t)
	openEntryRevealPrompt(t, g, basic("Swamp", "Swamp"))

	me := g.Seats[0]
	// Take the CANDIDATE out from under the open prompt, leaving the
	// entering land itself where it is (it is still in hand — that is
	// what "the entry is paused" means). Nothing in a real game can
	// do this while a choice blocks the table; the point is that even
	// then there is an answer.
	g.WithWriteLock(func() {
		kept := me.Hand.Cards[:0]
		for _, c := range me.Hand.Cards {
			if c.OracleID == entryRevealOracle {
				kept = append(kept, c)
			}
		}
		me.Hand.Cards = kept
	})

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) == 0 {
		t.Fatal("no answer to an entry reveal whose candidates left the hand")
	}
	dispatchAll(t, g, me.ID, moves)
	for _, m := range moves {
		if len(pickedCardIDs(t, m)) != 0 {
			t.Errorf("offered %q over a card that is no longer in hand", m.Label)
		}
	}
}

// TestEntryRevealIsTheOnlyMoveWhileItIsOpen is #791's gate seen from
// the enumerator: the prompt blocks the table, so the seat that owes
// it is offered its answers and no ordinary move.
func TestEntryRevealIsTheOnlyMoveWhileItIsOpen(t *testing.T) {
	stubRevealLandCatalog(t)
	g := newTable(t)
	openEntryRevealPrompt(t, g, basic("Swamp", "Swamp"))

	for _, seat := range g.Seats {
		for _, m := range legal.EnumerateFor(g, seat.ID) {
			if m.Kind != legal.KindChoice {
				t.Errorf("seat %s offered %q (%s) while the entry prompt is open", seat.Name, m.Label, m.Type)
			}
		}
	}
}
