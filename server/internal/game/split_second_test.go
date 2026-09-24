package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// split_second_test.go — #1519. Split second (CR 702.61) read off the
// card rather than off a sandbox flag no client sends. The consumers
// (CastSpell, the activation paths, the enumerator, the wire) predate
// this; these tests are about the WRITER, and about the rule's edges
// the writer now reaches for real: what it stops, what it does not,
// and that it ends with its spell.

// pushSplitSecondSpell puts an instant that PRINTS split second into
// a hand, the way the deck importer stamps one: Card.Keywords carries
// the canonical token, and there is no catalog entry behind it.
func pushSplitSecondSpell(p *Player, name string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = "Instant"
	c.Keywords = []string{KeywordSplitSecond}
	p.Hand.PushTop(c)
	return c.InstanceID
}

// The token is canonical, so the importer keeps Scryfall's "Split
// second" instead of filtering it, and every reader spells it alike.
func TestSplitSecondIsACanonicalKeyword(t *testing.T) {
	tok, ok := CanonicalKeyword("Split second")
	if !ok || tok != KeywordSplitSecond {
		t.Fatalf(`CanonicalKeyword("Split second") = %q, %v; want %q, true`, tok, ok, KeywordSplitSecond)
	}
	if !slices.Contains(CanonicalKeywordTable(), KeywordSplitSecond) {
		t.Error("split second is missing from CanonicalKeywordTable")
	}
}

// CR 702.61a: a spell that prints split second has it, with no flag.
// While it is on the stack nobody casts and nobody activates a
// non-mana ability — the caster included — and the refusal comes from
// the stamp on the stack item, not from anything the caller said.
func TestPrintedSplitSecondStopsCastsAndActivations(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	grip := pushSplitSecondSpell(me, "Krosan Grip")
	if err := g.CastSpell(me.ID, grip, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if item := g.StackMeta[grip]; item == nil || !item.SplitSecond {
		t.Fatal("a spell printing split second was put on the stack without it")
	}
	if !g.SplitSecondActive {
		t.Fatal("SplitSecondActive is false with a split-second spell on the stack")
	}

	bolt := pushTypedCardToHand(opp, "Lightning Bolt", "Instant")
	if err := g.CastSpell(opp.ID, bolt, CastSpellParams{}); !errors.Is(err, ErrSplitSecondActive) {
		t.Errorf("an opponent cast in response: got %v, want ErrSplitSecondActive", err)
	}
	mine := pushTypedCardToHand(me, "Opt", "Instant")
	if err := g.CastSpell(me.ID, mine, CastSpellParams{}); !errors.Is(err, ErrSplitSecondActive) {
		t.Errorf("the caster cast on top of their own split second: got %v, want ErrSplitSecondActive", err)
	}
	src := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: src, Name: "Prodigal Sorcerer", TypeLine: "Creature — Human Wizard", Owner: opp.ID, Controller: opp.ID})
	})
	if err := g.ActivateAbility(opp.ID, src, AbilityParams{Label: "{T}: 1 damage"}); !errors.Is(err, ErrSplitSecondActive) {
		t.Errorf("an ability was activated in response: got %v, want ErrSplitSecondActive", err)
	}
	if err := g.ActivateLoyalty(opp.ID, src, "+1", 1); !errors.Is(err, ErrSplitSecondActive) {
		t.Errorf("a loyalty ability was activated in response: got %v, want ErrSplitSecondActive", err)
	}
}

// CR 702.61b: mana abilities and special actions are not stopped. A
// land taps for mana in response, and a face-down morph is turned face
// up — both with a PRINTED split second on the stack, so the
// asymmetry is tested against the real stamp and not a hand-set cache.
func TestPrintedSplitSecondLeavesManaAndSpecialActions(t *testing.T) {
	g, me, morph := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	opp := g.Seats[1]
	if err := castFaceDown(t, g, me, morph, "morph"); err != nil {
		t.Fatalf("cast face down: %v", err)
	}
	resolveTop(t, g)

	forest := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	})
	grip := pushSplitSecondSpell(opp, "Krosan Grip")
	if err := g.CastSpell(opp.ID, grip, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !g.SplitSecondActive {
		t.Fatal("setup: split second is not active")
	}

	if err := g.ActivateManaAbility(opp.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Errorf("CR 702.61b: a mana ability was refused under split second: %v", err)
	}
	fillPool(g, me, 4)
	if err := g.PerformSpecialAction(me.ID, morph, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("CR 702.61b: turning a morph face up was refused under split second: %v", err)
	}
	if c, _, _ := cardAnywhere(g, morph); c.FaceDown {
		t.Error("the morph is still face down")
	}
}

// It ends with its spell (CR 702.61a, "as long as this spell is on
// the stack"): resolving and being countered both clear it, and the
// next cast goes through.
func TestSplitSecondEndsWhenTheSpellLeavesTheStack(t *testing.T) {
	for _, tc := range []struct {
		name  string
		leave func(t *testing.T, g *Game, id uuid.UUID)
	}{
		{"resolves", func(t *testing.T, g *Game, _ uuid.UUID) { resolveTop(t, g) }},
		{"is countered", func(t *testing.T, g *Game, id uuid.UUID) {
			if err := g.CounterSpell(id, nil); err != nil {
				t.Fatalf("CounterSpell: %v", err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			id := pushSplitSecondSpell(me, "Sudden Shock")
			if err := g.CastSpell(me.ID, id, CastSpellParams{}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			tc.leave(t, g, id)
			if g.SplitSecondActive {
				t.Fatal("split second outlived its spell")
			}
			next := pushTypedCardToHand(me, "Lightning Bolt", "Instant")
			if err := g.CastSpell(me.ID, next, CastSpellParams{}); err != nil {
				t.Errorf("a cast after the spell left: %v", err)
			}
		})
	}
}

// CR 708.2 / 708.4: a spell cast FACE DOWN is a 2/2 with no text, so
// it has no split second whatever the card underneath prints — and
// the sandbox flag is refused too, because a face-down spell that
// shut the table down would say which card it was.
func TestAFaceDownSpellHasNoSplitSecond(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].Keywords = []string{KeywordSplitSecond}
			}
		}
		g.Turn.Step = StepPrecombatMain
	})
	err := g.CastSpell(me.ID, id, CastSpellParams{
		FromZone:        "hand",
		AlternativeCost: "morph",
		Strict:          true,
		SplitSecond:     true,
	})
	if err != nil {
		t.Fatalf("cast face down: %v", err)
	}
	if item := g.StackMeta[id]; item == nil || item.SplitSecond {
		t.Error("a face-down spell was stamped with split second")
	}
	if g.SplitSecondActive {
		t.Error("a face-down spell turned split second on")
	}
}

// The sandbox flag still works for a card nobody has data for — it is
// routed through the same writer, not retired.
func TestSandboxSplitSecondFlagStillApplies(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	id := pushTypedCardToHand(me, "Homebrew Instant", "Instant")
	if err := g.CastSpell(me.ID, id, CastSpellParams{SplitSecond: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if item := g.StackMeta[id]; item == nil || !item.SplitSecond || !g.SplitSecondActive {
		t.Error("the sandbox split_second flag no longer turns split second on")
	}
}
