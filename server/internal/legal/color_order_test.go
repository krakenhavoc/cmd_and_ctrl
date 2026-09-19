package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// color_order_test.go — #986. The prompt has carried the card's
// `ColorPurpose` since #780, and until now only the heuristic policy
// read it: the enumerator ordered every colour prompt by "what this
// seat has most of", so Wash Out offered a mono-green seat green
// first, and so did the human's picker.
//
// Each test below is one purpose on one board, asserting the ORDER and
// never the membership — CR 105.4 makes all five legal and the
// ordering may not narrow that, which the last test pins separately.

// colorAnswerLabels is the enumerated colour answers, in order, as the
// colour words the move labels end with.
func colorAnswerLabels(t *testing.T, g *game.Game, seat uuid.UUID) []string {
	t.Helper()
	moves := legal.EnumerateFor(g, seat)
	if len(moves) == 0 {
		t.Fatal("the seat owing a colour prompt was offered nothing")
	}
	out := make([]string, 0, len(moves))
	for _, m := range moves {
		out = append(out, m.Label[strings.LastIndex(m.Label, ": ")+2:])
	}
	return out
}

func wantOrder(t *testing.T, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("answers = %v, want %v", got, want)
	}
}

// monoGreenFacingMonoRed is the board the issue names: the chooser has
// three green permanents and one opponent has four red ones, and
// nobody has anything else.
func monoGreenFacingMonoRed(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newTable(t)
	me, them := g.Seats[0], g.Seats[1]
	for range 3 {
		battlefieldCard(g, me, creature("Llanowar Elves", "{G}", 1, 1))
	}
	for range 4 {
		battlefieldCard(g, them, creature("Goblin Guide", "{R}", 2, 2))
	}
	return g, me
}

func queueColorPrompt(t *testing.T, g *game.Game, chooser uuid.UUID, purpose game.ColorPurpose, options []string) {
	t.Helper()
	g.WithWriteLock(func() {
		g.QueueColorChoiceForEffect(chooser, uuid.New(), "Prompt", options, purpose)
	})
}

// TestManaPromptStillRanksTheSeatsOwnColourFirst — the arm that must
// not move. Coldsteel Heart wants the mana the chooser will spend.
func TestManaPromptStillRanksTheSeatsOwnColourFirst(t *testing.T) {
	g, me := monoGreenFacingMonoRed(t)
	queueColorPrompt(t, g, me.ID, game.ColorForMana, nil)
	wantOrder(t, colorAnswerLabels(t, g, me.ID), "green", "white", "blue", "black", "red")
}

// TestUndeclaredPromptKeepsTheManaOrder — a prompt nobody has
// annotated is ordered exactly as it was before #986, for the reason
// the heuristic's default arm exists: an unannotated card must not get
// worse, only un-improved.
func TestUndeclaredPromptKeepsTheManaOrder(t *testing.T) {
	g, me := monoGreenFacingMonoRed(t)
	queueColorPrompt(t, g, me.ID, "", nil)
	wantOrder(t, colorAnswerLabels(t, g, me.ID), "green", "white", "blue", "black", "red")
}

// TestBenefitPromptRanksTheSeatsOwnColourFirst — Heraldic Banner's
// anthem and Selective Obliteration's survivor are the same question
// from the chooser's side: name the colour you want to keep.
func TestBenefitPromptRanksTheSeatsOwnColourFirst(t *testing.T) {
	g, me := monoGreenFacingMonoRed(t)
	queueColorPrompt(t, g, me.ID, game.ColorForBenefit, nil)
	wantOrder(t, colorAnswerLabels(t, g, me.ID), "green", "white", "blue", "black", "red")
}

// TestHarmPromptRanksTheOppositionsColourFirst is the bug, stated:
// Wash Out for a mono-green seat facing a mono-red board names RED,
// and green — which the pre-#986 order put first — now sorts LAST,
// because bouncing it costs the chooser three permanents and the
// opposition none.
func TestHarmPromptRanksTheOppositionsColourFirst(t *testing.T) {
	g, me := monoGreenFacingMonoRed(t)
	queueColorPrompt(t, g, me.ID, game.ColorForHarm, nil)
	wantOrder(t, colorAnswerLabels(t, g, me.ID), "red", "white", "blue", "black", "green")
}

// boardWhereHarmAndFilterDisagree: the chooser holds MORE red than the
// opposition does, and one opponent has a single, enormous white
// creature. Every purpose sorts this board differently, which is what
// makes it worth one fixture.
//
//	mine   G=3  R=5
//	theirs R=4  W=1
//	threat R=2  W=7
func boardWhereHarmAndFilterDisagree(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newTable(t)
	me, them := g.Seats[0], g.Seats[1]
	for range 3 {
		battlefieldCard(g, me, creature("Llanowar Elves", "{G}", 1, 1))
	}
	for range 5 {
		battlefieldCard(g, me, creature("Mogg Fanatic", "{R}", 1, 1))
	}
	for range 4 {
		battlefieldCard(g, them, creature("Goblin Guide", "{R}", 2, 2))
	}
	battlefieldCard(g, them, creature("Serra Avatar", "{W}{W}{W}", 7, 7))
	return g, me
}

// TestHarmIsNetOfWhatItCostsTheChooser — the chooser's own board is a
// term with a MINUS sign. Red is the commonest colour the opposition
// has, and naming it is still wrong here because the chooser has more
// of it; white, which costs the chooser nothing, sorts first.
func TestHarmIsNetOfWhatItCostsTheChooser(t *testing.T) {
	g, me := boardWhereHarmAndFilterDisagree(t)
	queueColorPrompt(t, g, me.ID, game.ColorForHarm, nil)
	wantOrder(t, colorAnswerLabels(t, g, me.ID), "white", "blue", "black", "red", "green")
}

// TestFilterIgnoresTheChoosersOwnBoard — Oona names a colour to pick
// out somebody else's cards, and nothing of the chooser's is at stake.
// Same board as the harm test, and a different answer.
func TestFilterIgnoresTheChoosersOwnBoard(t *testing.T) {
	g, me := boardWhereHarmAndFilterDisagree(t)
	queueColorPrompt(t, g, me.ID, game.ColorForFilter, nil)
	wantOrder(t, colorAnswerLabels(t, g, me.ID), "red", "white", "blue", "black", "green")
}

// TestProtectRanksTheBiggestThreatNotTheCommonest — Mother of Runes
// is held up against the creature that is about to kill you, not
// against the colour there is most of. One 7/7 outranks four 2/2s.
func TestProtectRanksTheBiggestThreatNotTheCommonest(t *testing.T) {
	g, me := boardWhereHarmAndFilterDisagree(t)
	queueColorPrompt(t, g, me.ID, game.ColorForProtection, nil)
	wantOrder(t, colorAnswerLabels(t, g, me.ID), "white", "red", "blue", "black", "green")
}

// TestEveryPurposeStillOffersEveryLegalAnswer is the floor. An
// ORDERING may not become a filter: CR 105.4 makes all five legal, and
// a narrowed printed list ("a color other than blue") stays exactly as
// narrow as the card made it, whatever the purpose.
func TestEveryPurposeStillOffersEveryLegalAnswer(t *testing.T) {
	for _, purpose := range append([]game.ColorPurpose{""}, game.AllColorPurposes...) {
		t.Run(string(purpose)+"/all-five", func(t *testing.T) {
			g, me := boardWhereHarmAndFilterDisagree(t)
			queueColorPrompt(t, g, me.ID, purpose, nil)
			got := colorAnswerLabels(t, g, me.ID)
			if len(got) != 5 {
				t.Fatalf("offered %d answers, want five: %v", len(got), got)
			}
			seen := map[string]bool{}
			for _, c := range got {
				seen[c] = true
			}
			for _, want := range []string{"white", "blue", "black", "red", "green"} {
				if !seen[want] {
					t.Errorf("%s is not on offer: %v", want, got)
				}
			}
			dispatchAll(t, g, me.ID, legal.EnumerateFor(g, me.ID))
		})

		t.Run(string(purpose)+"/other-than-blue", func(t *testing.T) {
			g, me := boardWhereHarmAndFilterDisagree(t)
			queueColorPrompt(t, g, me.ID, purpose, game.ColorsOtherThan("U"))
			got := colorAnswerLabels(t, g, me.ID)
			if len(got) != 4 {
				t.Fatalf("offered %d answers, want four: %v", len(got), got)
			}
			for _, c := range got {
				if c == "blue" {
					t.Errorf("blue was offered for a prompt that excludes it: %v", got)
				}
			}
		})
	}
}

// TestColorOrderIsStableAcrossCalls — the sort is stable and the
// inputs are counts, so two enumerations of one board are identical.
// Without it a map iteration order would leak into the move list and a
// bot's tie-break would wobble between frames.
func TestColorOrderIsStableAcrossCalls(t *testing.T) {
	g, me := boardWhereHarmAndFilterDisagree(t)
	queueColorPrompt(t, g, me.ID, game.ColorForHarm, nil)
	first := strings.Join(colorAnswerLabels(t, g, me.ID), ",")
	for range 20 {
		if again := strings.Join(colorAnswerLabels(t, g, me.ID), ","); again != first {
			t.Fatalf("two enumerations of one board disagreed: %q then %q", first, again)
		}
	}
}

// TestTheWireButtonsAreTheEnumeratorsOrder is the whole point of
// #986's second half: the human's colour buttons and the bot's move
// list come out of the SAME ordering function, so the button under the
// cursor and the first offered move are the same colour. A second
// ranking rule on either side is what this pins shut.
func TestTheWireButtonsAreTheEnumeratorsOrder(t *testing.T) {
	for _, purpose := range append([]game.ColorPurpose{""}, game.AllColorPurposes...) {
		t.Run(string(purpose), func(t *testing.T) {
			g, me := boardWhereHarmAndFilterDisagree(t)
			queueColorPrompt(t, g, me.ID, purpose, nil)

			v := protocol.ViewOfGameFor(g, me.ID.String())
			if len(v.PendingChoices) != 1 {
				t.Fatalf("view carries %d prompts, want 1", len(v.PendingChoices))
			}
			prompt := v.PendingChoices[0]
			if prompt.ColorPurpose != string(purpose) {
				t.Errorf("color_purpose on the wire = %q, want %q", prompt.ColorPurpose, purpose)
			}
			words := make([]string, 0, len(prompt.ColorOptions))
			for _, c := range prompt.ColorOptions {
				words = append(words, game.ColorName(c))
			}
			wantOrder(t, words, colorAnswerLabels(t, g, me.ID)...)
		})
	}
}

// TestOrderingAnEmptyOrSingleOptionListIsANoOp — the degenerate inputs,
// pinned rather than reasoned about.
//
// #1016 made a dropped prompt run its continuation with "nobody chose"
// (`NoChoiceIndex`), and that constant's own doc names an EMPTY OPTION
// LIST as one of the two ways a question is never put at all. A colour
// prompt cannot reach that state today — `normaliseColorOptions` turns
// nil or empty into all five (CR 105.4), and a dropped prompt is
// dequeued, so neither the enumerator nor the view projection is ever
// called for one — but the ordering function is what both of them reach
// THROUGH, and "it cannot happen" is not a reason for it to panic if it
// ever does. It short-circuits before it walks the battlefield.
func TestOrderingAnEmptyOrSingleOptionListIsANoOp(t *testing.T) {
	g, me := boardWhereHarmAndFilterDisagree(t)
	for _, purpose := range append([]game.ColorPurpose{""}, game.AllColorPurposes...) {
		for _, options := range [][]string{nil, {}, {"G"}} {
			var got []string
			g.ReadSnapshot(func() {
				got = legal.OrderColorOptionsLocked(g, me.ID, options, purpose)
			})
			if strings.Join(got, ",") != strings.Join(options, ",") {
				t.Errorf("%q ordering %v returned %v", purpose, options, got)
			}
		}
	}
	// And an unseated chooser. CR 800.4a takes the object out of the
	// game along with the player, so this is the same "cannot happen",
	// held to the same standard.
	var got []string
	g.ReadSnapshot(func() {
		got = legal.OrderColorOptionsLocked(g, uuid.Nil, game.AllColors, game.ColorForHarm)
	})
	if strings.Join(got, ",") != strings.Join(game.AllColors, ",") {
		t.Errorf("an unseated chooser reordered the list: %v", got)
	}
}

// TestTheEngineNeverQueuesAColourPromptWithNoOptions is the other half:
// the reason the degenerate case above stays hypothetical. CR 105.4
// makes all five colours legal, so "no options" is not a colour prompt
// — it is a bug, and `normaliseColorOptions` is where it is stopped.
func TestTheEngineNeverQueuesAColourPromptWithNoOptions(t *testing.T) {
	for _, options := range [][]string{nil, {}, {"C"}, {"not a colour"}} {
		g := newTable(t)
		me := g.Seats[0]
		queueColorPrompt(t, g, me.ID, game.ColorForHarm, options)
		var queued []string
		g.ReadSnapshot(func() {
			for _, c := range g.PendingChoices {
				if c.Kind == game.PendingChoiceColor {
					queued = c.ColorOptions
				}
			}
		})
		if len(queued) != len(game.AllColors) {
			t.Errorf("a prompt built from %v queued %v, want all five (CR 105.4)", options, queued)
		}
	}
}
