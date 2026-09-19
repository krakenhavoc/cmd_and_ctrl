package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyword_action_replacements_test.go — #976 from the card side: the
// CR 614 window on a keyword action, watched through the ordinary
// Spec.Replacements slot.
//
// Tekuthal, Inquiry Dominus is the printed proliferate half and it is
// tested through its own catalog entry. The scry half has no printed
// card in the catalog yet — "if you would scry, scry that many plus
// one instead" is the Crystal Ball shape and nothing in the tree
// prints it — so the two probes below stand in for it. They declare
// the SAME helpers a card file would (ScryPlusOne, and the generic
// KeywordActionBecomes for the doubler), which is the point: the
// helper is what a card would reach for, so the helper is what is
// under test.

const (
	scryPlusOneProbeOracle    = "test-scry-plus-one-probe"
	scryDoubledProbeOracle    = "test-scry-doubled-probe"
	proliferateProbeOracle    = "test-proliferate-twice-probe"
	surveilPlusOneProbeOracle = "test-surveil-plus-one-probe"
)

func init() {
	Register(Spec{
		OracleID:     scryPlusOneProbeOracle,
		Name:         "Scrying Probe",
		Replacements: []game.ReplacementEffect{ScryPlusOne("Scrying Probe — scry one more")},
	})
	Register(Spec{
		OracleID: scryDoubledProbeOracle,
		Name:     "Doubling Probe",
		Replacements: []game.ReplacementEffect{
			KeywordActionBecomes(game.KeywordActionScry, func(n int) int { return n * 2 },
				"Doubling Probe — scry twice as many"),
		},
	})
	Register(Spec{
		OracleID:     proliferateProbeOracle,
		Name:         "Proliferating Probe",
		Replacements: []game.ReplacementEffect{ProliferateTwice("Proliferating Probe — proliferate twice")},
	})
	Register(Spec{
		OracleID:     surveilPlusOneProbeOracle,
		Name:         "Surveilling Probe",
		Replacements: []game.ReplacementEffect{SurveilPlusOne("Surveilling Probe — surveil one more")},
	})
}

// --- proliferate: Tekuthal's last caveat -----------------------------

// The headline. "If you would proliferate, proliferate twice instead"
// is a count of TIMES (CR 701.34), so Steady Progress's one
// proliferate becomes two and a creature that started with one +1/+1
// counter ends with three.
func TestTekuthalProliferatesTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTekuthal(g, me.ID)
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)
	theirs := pushCounterCreature(g, opp.ID, "Theirs", game.CounterMinusOne, 1)

	castCatalogSpell(t, g, "Steady Progress", "Instant", steadyProgressOracle, nil)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, mine, game.CounterPlusOne); got != 3 {
		t.Errorf("my +1/+1 counters = %d, want 3 (one printed, two proliferates)", got)
	}
	if got := counterCount(g, theirs, game.CounterMinusOne); got != 3 {
		t.Errorf("their -1/-1 counters = %d, want 3 — the doubled action takes the whole choice twice", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("one replacement queued %d prompts, want none", len(g.PendingChoices))
	}
}

// Without Tekuthal the same spell proliferates once, which is the
// control the test above needs: the difference is the card, not the
// plumbing.
func TestSteadyProgressWithoutTekuthalProliferatesOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)

	castCatalogSpell(t, g, "Steady Progress", "Instant", steadyProgressOracle, nil)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("my +1/+1 counters = %d, want 2", got)
	}
}

// Two Tekuthals are two objects contributing ONE declared effect, so
// #792's identical-window skip applies: four proliferates, and nobody
// is asked to order ×2 against ×2.
//
// The board is transient in paper — Tekuthal is legendary, so CR
// 704.5j would bin one at the next state-based check — which is why
// the proliferate is taken directly here rather than through a spell
// that would sweep first. What is under test is the IDENTITY (the
// same catalog entry, the same slot, the same controller), and that
// is exactly what two copies of one legend give it.
func TestTwoTekuthalsProliferateFourTimesWithoutAPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTekuthal(g, me.ID)
	pushTekuthal(g, me.ID)
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)

	g.WithWriteLock(func() {
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{mine}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	if got := counterCount(g, mine, game.CounterPlusOne); got != 5 {
		t.Errorf("+1/+1 counters = %d, want 5 (one printed, four proliferates)", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("two copies of one card queued %d prompts, want none", len(g.PendingChoices))
	}
}

// The same skip on a board the legend rule allows: two copies of one
// non-legendary card carrying the same helper.
func TestTwoCopiesOfAProliferateDoublerDoNotPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedReplacementPermanent(g, proliferateProbeOracle, "Proliferating Probe", me.ID)
	seedReplacementPermanent(g, proliferateProbeOracle, "Proliferating Probe", me.ID)
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)

	g.WithWriteLock(func() {
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{mine}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	if got := counterCount(g, mine, game.CounterPlusOne); got != 5 {
		t.Errorf("+1/+1 counters = %d, want 5", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("two copies of one card queued %d prompts, want none", len(g.PendingChoices))
	}
}

// An opponent's Tekuthal does not double YOUR proliferate: the clause
// is "if YOU would proliferate".
func TestAnOpponentsTekuthalDoesNotDoubleYourProliferate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTekuthal(g, opp.ID)
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)

	g.WithWriteLock(func() {
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{mine}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	if got := counterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

// --- scry ------------------------------------------------------------

// "If you would scry, scry that many plus one instead": scry 2 becomes
// scry 3, and the prompt the player answers is the bigger one.
func TestScryPlusOneMakesScryTwoIntoScryThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedReplacementPermanent(g, scryPlusOneProbeOracle, "Scrying Probe", me.ID)

	g.WithWriteLock(func() {
		if err := (Scry{N: 2}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("Scry: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the scry prompt", len(g.PendingChoices))
	}
	p := g.PendingChoices[0]
	if p.Kind != game.PendingChoiceScry {
		t.Fatalf("prompt kind = %q, want %q", p.Kind, game.PendingChoiceScry)
	}
	if len(p.ScryCards) != 3 {
		t.Errorf("the prompt offers %d cards, want 3", len(p.ScryCards))
	}
}

// Surveil is the same keyword action with the other away lane, and a
// "plus one" on it is the same helper — and a SCRY replacement must
// not reach it, which is what the action discriminator is for.
func TestSurveilPlusOneMakesSurveilOneIntoSurveilTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedReplacementPermanent(g, surveilPlusOneProbeOracle, "Surveilling Probe", me.ID)
	seedReplacementPermanent(g, scryPlusOneProbeOracle, "Scrying Probe", me.ID)

	g.WithWriteLock(func() {
		if err := (Surveil{N: 1}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("Surveil: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the surveil prompt", len(g.PendingChoices))
	}
	p := g.PendingChoices[0]
	if p.Kind != game.PendingChoiceSurveil {
		t.Fatalf("prompt kind = %q, want %q — the scry probe must not have ordered against it", p.Kind, game.PendingChoiceSurveil)
	}
	if len(p.ScryCards) != 2 {
		t.Errorf("the prompt offers %d cards, want 2", len(p.ScryCards))
	}
}

// A doubler and a "plus one" in the same window are two DIFFERENT
// declared effects, so CR 616.1 gives the scrying player the order —
// and the orders differ: ×2 then +1 over a printed 2 is 5, +1 then ×2
// is 6.
func TestScryDoublerAndPlusOnePromptForOrder(t *testing.T) {
	for _, tc := range []struct {
		name  string
		order []string
		cards int
	}{
		{"double then plus one", []string{"Doubling Probe — scry twice as many", "Scrying Probe — scry one more"}, 5},
		{"plus one then double", []string{"Scrying Probe — scry one more", "Doubling Probe — scry twice as many"}, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			seedReplacementPermanent(g, scryPlusOneProbeOracle, "Scrying Probe", me.ID)
			seedReplacementPermanent(g, scryDoubledProbeOracle, "Doubling Probe", me.ID)

			g.WithWriteLock(func() {
				if err := (Scry{N: 2}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
					t.Fatalf("Scry: %v", err)
				}
			})
			if len(g.PendingChoices) != 1 {
				t.Fatalf("pending choices = %d, want the CR 616 ordering prompt", len(g.PendingChoices))
			}
			prompt := g.PendingChoices[0]
			if prompt.Kind != game.PendingChoiceReplacementOrder {
				t.Fatalf("prompt kind = %q, want %q", prompt.Kind, game.PendingChoiceReplacementOrder)
			}
			if prompt.Chooser != me.ID {
				t.Errorf("chooser = %s, want the scrying player %s", prompt.Chooser, me.ID)
			}
			if err := g.ResolveReplacementOrder(prompt.ID, me.ID,
				replacementIDsByLabel(t, g, prompt.ReplacementEffectIDs, tc.order)); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}
			if len(g.PendingChoices) != 1 {
				t.Fatalf("pending choices after the order = %d, want the scry prompt", len(g.PendingChoices))
			}
			scry := g.PendingChoices[0]
			if scry.Kind != game.PendingChoiceScry {
				t.Fatalf("prompt kind = %q, want %q", scry.Kind, game.PendingChoiceScry)
			}
			if len(scry.ScryCards) != tc.cards {
				t.Errorf("the resumed scry looks at %d cards, want %d", len(scry.ScryCards), tc.cards)
			}
		})
	}
}

// replacementIDsByLabel maps a prompt's effect IDs onto the labels the
// test names, in the order it names them.
func replacementIDsByLabel(t *testing.T, g *game.Game, ids []game.ReplacementEffectID, labels []string) []game.ReplacementEffectID {
	t.Helper()
	byLabel := make(map[string]game.ReplacementEffectID, len(ids))
	g.ReadSnapshot(func() {
		for _, id := range ids {
			label, _ := g.ReplacementOptionMetaForEffect(id)
			byLabel[label] = id
		}
	})
	out := make([]game.ReplacementEffectID, 0, len(labels))
	for _, want := range labels {
		id, ok := byLabel[want]
		if !ok {
			t.Fatalf("the prompt has no option labelled %q (has %v)", want, byLabel)
		}
		out = append(out, id)
	}
	return out
}
