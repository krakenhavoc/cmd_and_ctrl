package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mill_replacements_test.go — #569 from the card side: the CR 614
// window on the mill AMOUNT, watched through the ordinary
// Spec.Replacements slot.
//
// Two printed cards, deliberately, so the kind is not shaped around
// one of them: Bruvac the Grandiloquent's ×2 and The Water Crystal's
// +4 are the same sentence with different arithmetic, and CR 616.1's
// ordering question between them has two different answers.
//
// It is also the first printed board on which #982's rule is
// observable. Both replacements belong to whoever controls the
// artifacts; the affected player is the OPPONENT being milled, so the
// ordering prompt goes to them.

const (
	bruvacOracle          = "274b999f-f193-48fd-9a4a-0fdaf535e6c3"
	theWaterCrystalOracle = "f8d2a94f-7be1-4b17-8556-398dde531360"
)

// pushMillReplacement puts one of the two cards onto the battlefield
// under `owner`.
func pushMillReplacement(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return pushCatalogPermanent(g, owner, name, typeLine, oracle, false)
}

// millSeat runs a fire-and-forget mill of n cards against `seat` and
// reports how many cards reached their graveyard.
func millSeat(t *testing.T, g *game.Game, seat *game.Player, n int) int {
	t.Helper()
	before := seat.Graveyard.Size()
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(seat.ID, n); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})
	return seat.Graveyard.Size() - before
}

// --- the cards are wired ---------------------------------------------

func TestMillAmountReplacementsAreWired(t *testing.T) {
	for _, oracle := range []string{bruvacOracle, theWaterCrystalOracle} {
		if n := len(game.CatalogReplacements(oracle)); n != 1 {
			t.Errorf("%s declared %d replacements, want 1", oracle, n)
		}
	}
}

// --- Bruvac ----------------------------------------------------------

// The headline. "If an opponent would mill one or more cards, they mill
// twice that many cards instead" — one event per INSTRUCTION, so a mill
// of three becomes a mill of six rather than three mills of two.
func TestBruvacDoublesAnOpponentsMill(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushMillReplacement(g, me.ID, "Bruvac the Grandiloquent", "Legendary Creature — Human Advisor", bruvacOracle)

	if got := millSeat(t, g, opp, 3); got != 6 {
		t.Errorf("the opponent milled %d cards, want 6", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("one replacement queued %d prompts, want none (#792)", len(g.PendingChoices))
	}
}

// "An opponent" is the whole scope clause: Bruvac's own controller
// mills what they were told to.
func TestBruvacDoesNotDoubleItsOwnControllersMill(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushMillReplacement(g, me.ID, "Bruvac the Grandiloquent", "Legendary Creature — Human Advisor", bruvacOracle)

	if got := millSeat(t, g, me, 3); got != 3 {
		t.Errorf("Bruvac's controller milled %d cards, want 3", got)
	}
}

// Two Bruvacs are ×4 and nobody is asked to order them: two objects
// contributing ONE declared effect is #792's identical-window skip, and
// ×2 then ×2 is ×4 either way. (Two is not a legal board in Commander —
// Bruvac is legendary — but the identity rule is the engine's, not the
// legend rule's, and a token copy would get there.)
func TestTwoBruvacsAreFourTimesAndNoPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushMillReplacement(g, me.ID, "Bruvac the Grandiloquent", "Legendary Creature — Human Advisor", bruvacOracle)
	pushMillReplacement(g, me.ID, "Bruvac the Grandiloquent", "Legendary Creature — Human Advisor", bruvacOracle)

	if got := millSeat(t, g, opp, 2); got != 8 {
		t.Errorf("the opponent milled %d cards, want 8", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("two copies of one declared effect queued %d prompts, want none", len(g.PendingChoices))
	}
}

// --- The Water Crystal -----------------------------------------------

// "That many cards plus four" adds to the count, and the same scope
// clause applies.
func TestTheWaterCrystalAddsFourToAnOpponentsMill(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushMillReplacement(g, me.ID, "The Water Crystal", "Legendary Artifact", theWaterCrystalOracle)

	if got := millSeat(t, g, opp, 1); got != 5 {
		t.Errorf("the opponent milled %d cards, want 5", got)
	}
	if got := millSeat(t, g, me, 1); got != 1 {
		t.Errorf("the Crystal's own controller milled %d cards, want 1", got)
	}
}

// "Blue spells you cast cost {1} less to cast" — the colour predicate
// is new (ColoredSpell) and it reads the spell's effective colours, so
// a multicolour spell with blue in it is a blue spell (CR 105.2b).
func TestTheWaterCrystalDiscountsBlueSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushMillReplacement(g, me.ID, "The Water Crystal", "Legendary Artifact", theWaterCrystalOracle)

	for _, tc := range []struct {
		name, typeLine, cost string
		seat                 *game.Player
		want                 int
	}{
		{"Blue Bear", "Creature — Bear", "{2}{U}", me, 2},
		{"Azorius Bear", "Creature — Bear", "{1}{W}{U}", me, 2},
		{"Red Bear", "Creature — Bear", "{2}{R}", me, 3},
		{"Colourless Bear", "Artifact Creature — Bear", "{3}", me, 3},
		{"Their Blue Bear", "Creature — Bear", "{2}{U}", opp, 3},
	} {
		if got := priceInHand(t, g, tc.seat, tc.name, tc.typeLine, tc.cost); got != tc.want {
			t.Errorf("%s costs %d, want %d", tc.name, got, tc.want)
		}
	}
}

// --- the two of them together ----------------------------------------

// CR 616.1: two DIFFERENT declared effects in one window, and the
// orderings differ — ×2 then +4 is 10 from a base of 3, +4 then ×2 is
// 14 — so the affected player is really asked.
//
// And the affected player is the OPPONENT, not the player who controls
// both artifacts. That is #982's rule on a printed board: before it,
// affectedPlayerForEvent fell through to the first gathered effect's
// controller for any kind it had no case for.
func TestBruvacAndTheWaterCrystalAskTheMilledPlayerForTheOrder(t *testing.T) {
	for _, tc := range []struct {
		name  string
		order []string
		want  int
	}{
		{"Bruvac first", []string{
			"Bruvac the Grandiloquent — mill twice that many",
			"The Water Crystal — mill that many plus four",
		}, 10},
		{"the Crystal first", []string{
			"The Water Crystal — mill that many plus four",
			"Bruvac the Grandiloquent — mill twice that many",
		}, 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pushMillReplacement(g, me.ID, "Bruvac the Grandiloquent", "Legendary Creature — Human Advisor", bruvacOracle)
			pushMillReplacement(g, me.ID, "The Water Crystal", "Legendary Artifact", theWaterCrystalOracle)
			stockLibraryFor(g, opp, 30)

			before := opp.Graveyard.Size()
			g.WithWriteLock(func() {
				if err := g.MillNForEffect(opp.ID, 3); err != nil {
					t.Fatalf("MillNForEffect: %v", err)
				}
			})
			if n := opp.Graveyard.Size() - before; n != 0 {
				t.Fatalf("%d cards were milled before the prompt was answered", n)
			}

			if len(g.PendingChoices) != 1 {
				t.Fatalf("%d prompts open, want exactly 1", len(g.PendingChoices))
			}
			c := g.PendingChoices[0]
			if c.Kind != game.PendingChoiceReplacementOrder {
				t.Fatalf("prompt kind = %q, want %q", c.Kind, game.PendingChoiceReplacementOrder)
			}
			if c.Chooser != opp.ID {
				t.Fatalf("chooser = %s, want the milled opponent %s — CR 616.1 asks the AFFECTED player, "+
					"not the player who controls both replacements (#982)", c.Chooser, opp.ID)
			}

			if err := g.ResolveReplacementOrder(c.ID, c.Chooser, replacementOrderByLabel(t, g, c, tc.order...)); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}
			if got := opp.Graveyard.Size() - before; got != tc.want {
				t.Errorf("the opponent milled %d cards, want %d", got, tc.want)
			}
		})
	}
}

// --- the Crystal's activated ability ---------------------------------

// "{4}{U}{U}, {T}: Each opponent mills cards equal to the number of
// cards in your hand." Each opponent's mill is its own instruction, so
// the Crystal's own clause applies to each of them.
func TestTheWaterCrystalMillsEachOpponentForYourHandSize(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	crystal := pushMillReplacement(g, me.ID, "The Water Crystal", "Legendary Artifact", theWaterCrystalOracle)
	handSize := me.Hand.Size()
	if handSize == 0 {
		t.Fatal("setup: the opening hand is empty, so the ability would mill nothing")
	}
	before := make([]int, len(g.Seats))
	for i, s := range g.Seats {
		before[i] = s.Graveyard.Size()
	}

	ability := game.ActivatedAbilitiesForCard(game.Card{OracleID: theWaterCrystalOracle})
	if len(ability) != 1 {
		t.Fatalf("The Water Crystal declared %d activated abilities, want 1", len(ability))
	}
	g.WithWriteLock(func() {
		item := &game.StackItem{
			Kind:         game.StackItemActivated,
			Controller:   me.ID,
			SourceCardID: crystal,
		}
		if err := ability[0].Effect(g, item); err != nil {
			t.Fatalf("The Water Crystal ability: %v", err)
		}
	})

	for i, s := range g.Seats {
		got := s.Graveyard.Size() - before[i]
		want := handSize + 4
		if s.ID == me.ID {
			want = 0
		}
		if got != want {
			t.Errorf("seat %d milled %d cards, want %d", i, got, want)
		}
	}
}

// stockLibraryFor makes sure a seat has at least n cards to mill, so a
// count test measures the replacement and not the deck size.
func stockLibraryFor(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for p.Library.Size() < n {
			p.Library.PushTop(game.NewCard("basic-filler", p.ID))
		}
	})
}

// replacementOrderByLabel turns prompt labels into the ID permutation
// ResolveReplacementOrder takes.
func replacementOrderByLabel(t *testing.T, g *game.Game, c *game.PendingChoice, labels ...string) []game.ReplacementEffectID {
	t.Helper()
	byLabel := make(map[string]game.ReplacementEffectID, len(c.ReplacementEffectIDs))
	for _, id := range c.ReplacementEffectIDs {
		label, _ := g.ReplacementOptionMetaForEffect(id)
		byLabel[label] = id
	}
	out := make([]game.ReplacementEffectID, 0, len(labels))
	for _, want := range labels {
		id, ok := byLabel[want]
		if !ok {
			t.Fatalf("prompt has no option labelled %q (has %v)", want, byLabel)
		}
		out = append(out, id)
	}
	return out
}
