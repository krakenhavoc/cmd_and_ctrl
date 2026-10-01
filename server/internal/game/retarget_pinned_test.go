package game

import (
	"testing"

	"github.com/google/uuid"
)

// retarget_pinned_test.go — #1743, the pinned retarget: "change a
// target of target spell or ability TO THIS CREATURE" (Spellskite).
// What these pin is that the destination is judged by exactly the gate
// a free retarget's answer passes, and that the only question the
// variant ever asks is WHICH slot.
//
// "Skite" is the destination throughout: an ordinary 2/2 standing in
// for the permanent the ability names.

const (
	pinnedTwoOfOneClauseOracle = "test-pinned-two-target-creatures"
	pinnedTwoClausesOracle     = "test-pinned-two-clauses-same-allowed"
)

// pinnedSpecs installs retarget_test.go's clauses plus two of this
// file's own: "two target creatures" (ONE clause, two picks, so the
// same creature cannot be named twice) and "target creature … target
// creature" (two clauses with no "another", so it can, CR 601.2c).
func pinnedSpecs(t *testing.T) {
	t.Helper()
	withCatalogTargetSpec(t, func(oracleID string) *TargetSpec {
		switch oracleID {
		case retargetBoltOracle:
			return anyTargetSpec()
		case retargetPairOracle:
			return twoCreatureSpec()
		case retargetOneOracle:
			return oneCreatureSpec()
		case pinnedTwoOfOneClauseOracle:
			s := oneCreatureSpec()
			s.Label = "two target creatures"
			s.Min, s.Max = 2, 2
			return s
		case pinnedTwoClausesOracle:
			first := oneCreatureSpec()
			second := oneCreatureSpec()
			return first.Then(second)
		}
		return nil
	})
}

func pinnedTo(id uuid.UUID) TargetRef { return TargetRef{Kind: TargetCard, ID: id} }

func setKeyword(g *Game, id uuid.UUID, kw string) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, kw)
			}
		}
	})
}

func offerPinned(t *testing.T, g *Game, item *StackItem, chooser, to uuid.UUID, optional bool) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: chooser, Policy: RetargetChangeOne,
			Optional: optional, To: pinnedTo(to), Reason: "Skite",
		}); err != nil {
			t.Fatalf("OfferRetargetForEffect (pinned): %v", err)
		}
	})
}

func targetIDs(g *Game, itemID uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, t := range g.StackMeta[itemID].Targets {
		out = append(out, t.ID)
	}
	return out
}

// The whole variant in one test: a Bolt at one creature is pulled onto
// Skite, with no prompt — there is one target and one place for it to
// go — and Skite becomes a target of the BOLT, attributed to the Bolt's
// controller, not to Skite's.
func TestPinnedRetargetMovesTheTargetToTheDestination(t *testing.T) {
	g := newActiveGame(t)
	pinnedSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushRetargetCreature(g, opp, "Victim")
	skite := pushRetargetCreature(g, opp, "Skite")

	item := castTargeted(t, g, me, retargetBoltOracle, []TargetRef{{Kind: TargetCard, ID: victim}})
	offerPinned(t, g, item, opp.ID, skite, false)

	if p := findRetargetPrompt(g, opp.ID); p != nil {
		t.Fatalf("a forced change asked anyway: %+v", p)
	}
	if got := targetIDs(g, item.ID); len(got) != 1 || got[0] != skite {
		t.Fatalf("targets = %v, want [Skite]", got)
	}
	var saw bool
	for _, ev := range g.Events {
		if ev.Kind == EventBecomesTarget && ev.Target == skite {
			saw = true
			if ev.Actor != me.ID {
				t.Errorf("EventBecomesTarget actor = %v, want the Bolt's controller", ev.Actor)
			}
		}
	}
	if !saw {
		t.Error("no EventBecomesTarget for the destination")
	}
}

// A player destination is the same code: the clause decides.
func TestPinnedRetargetToAPlayer(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	pinnedSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]

	item := castTargeted(t, g, me, retargetBoltOracle, []TargetRef{{Kind: TargetPlayer, ID: opp.ID}})
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: opp.ID, Policy: RetargetChangeOne,
			To: TargetRef{Kind: TargetPlayer, ID: third.ID},
		}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
	if got := g.StackMeta[item.ID].Targets; got[0].Kind != TargetPlayer || got[0].ID != third.ID {
		t.Fatalf("targets = %+v, want the third seat", got)
	}
}

// An illegal destination leaves the target exactly where it was and
// asks nobody — every way the gate can say no: the clause's predicate
// (an artifact under "target creature"), hexproof judged for the
// ITEM's controller, shroud, the destination already being the
// target, and the destination having left the battlefield.
func TestPinnedRetargetToAnIllegalDestinationChangesNothing(t *testing.T) {
	cases := []struct {
		name  string
		setup func(g *Game, me, opp *Player) (dest uuid.UUID, target uuid.UUID)
	}{
		{"not a creature", func(g *Game, me, opp *Player) (uuid.UUID, uuid.UUID) {
			return pushArtifact(g, opp, "Rock"), pushRetargetCreature(g, opp, "Victim")
		}},
		{"hexproof, controlled by an opponent of the spell's controller", func(g *Game, me, opp *Player) (uuid.UUID, uuid.UUID) {
			skite := pushRetargetCreature(g, opp, "Skite")
			setKeyword(g, skite, "hexproof")
			return skite, pushRetargetCreature(g, opp, "Victim")
		}},
		{"shroud, controlled by the spell's own controller", func(g *Game, me, opp *Player) (uuid.UUID, uuid.UUID) {
			skite := pushRetargetCreature(g, me, "Skite")
			setKeyword(g, skite, "shroud")
			return skite, pushRetargetCreature(g, opp, "Victim")
		}},
		{"already the target", func(g *Game, me, opp *Player) (uuid.UUID, uuid.UUID) {
			skite := pushRetargetCreature(g, opp, "Skite")
			return skite, skite
		}},
		{"left the battlefield", func(g *Game, me, opp *Player) (uuid.UUID, uuid.UUID) {
			skite := pushRetargetCreature(g, opp, "Skite")
			victim := pushRetargetCreature(g, opp, "Victim")
			g.WithWriteLock(func() {
				c, err := g.Battlefield.Remove(skite)
				if err != nil {
					panic(err)
				}
				opp.Graveyard.PushTop(c)
			})
			return skite, victim
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			pinnedSpecs(t)
			advanceTo(t, g, StepPrecombatMain)
			me, opp := g.Seats[0], g.Seats[1]
			dest, target := tc.setup(g, me, opp)

			item := castTargeted(t, g, me, retargetOneOracle, []TargetRef{{Kind: TargetCard, ID: target}})
			offerPinned(t, g, item, opp.ID, dest, false)
			if p := findRetargetPrompt(g, opp.ID); p != nil {
				t.Errorf("an illegal destination opened a prompt: %+v", p)
			}
			if got := targetIDs(g, item.ID); len(got) != 1 || got[0] != target {
				t.Errorf("targets = %v, want unchanged [%v]", got, target)
			}
		})
	}
}

// Hexproof is the item's controller's question, not the redirector's:
// the spell's own controller may have it pulled onto their own
// hexproof creature (CR 702.11b — hexproof stops OPPONENTS).
func TestPinnedRetargetHexproofIsJudgedForTheItemsController(t *testing.T) {
	g := newActiveGame(t)
	pinnedSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushRetargetCreature(g, opp, "Victim")
	skite := pushRetargetCreature(g, me, "My Skite")
	setKeyword(g, skite, "hexproof")

	item := castTargeted(t, g, me, retargetOneOracle, []TargetRef{{Kind: TargetCard, ID: victim}})
	offerPinned(t, g, item, me.ID, skite, false)
	if got := targetIDs(g, item.ID); got[0] != skite {
		t.Errorf("targets = %v, want [My Skite]", got)
	}
}

// "Change A target" over an item with several: the chooser picks
// WHICH, from the current targets that could become the destination,
// and that is the only question. The answer names the slot's current
// object; the destination is never the answer.
func TestPinnedRetargetWithSeveralSlotsAsksWhichOne(t *testing.T) {
	g := newActiveGame(t)
	pinnedSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	other := pushRetargetCreature(g, opp, "Bystander")
	skite := pushRetargetCreature(g, opp, "Skite")

	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0},
		{Kind: TargetCard, ID: b, Slot: 1},
	})
	offerPinned(t, g, item, opp.ID, skite, false)
	prompt := findRetargetPrompt(g, opp.ID)
	if prompt == nil {
		t.Fatal("two eligible targets opened no prompt")
	}
	if !hasUUID(prompt.PickTargetCards, a) || !hasUUID(prompt.PickTargetCards, b) || len(prompt.PickTargetCards) != 2 {
		t.Errorf("options = %v, want exactly the two current targets", prompt.PickTargetCards)
	}
	if prompt.PickTargetMin != 1 {
		t.Errorf("Spellskite's change is mandatory: min = %d, want 1", prompt.PickTargetMin)
	}
	if prompt.RetargetTo != pinnedTo(skite) {
		t.Errorf("RetargetTo = %+v, want Skite", prompt.RetargetTo)
	}

	if err := g.ResolveRetarget(prompt.ID, opp.ID, nil); err != ErrInvalidParam {
		t.Errorf("declining a mandatory pinned change = %v, want ErrInvalidParam", err)
	}
	for _, wrong := range []uuid.UUID{skite, other} {
		if err := g.ResolveRetarget(prompt.ID, opp.ID, []TargetRef{pinnedTo(wrong)}); err != ErrIllegalTarget {
			t.Errorf("answering with an object in no eligible slot = %v, want ErrIllegalTarget", err)
		}
	}
	if err := g.ResolveRetarget(prompt.ID, opp.ID, []TargetRef{pinnedTo(b)}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	got := g.StackMeta[item.ID].Targets
	if len(got) != 2 || got[0].ID != a || got[1].ID != skite {
		t.Fatalf("targets = %+v, want [A Skite]", got)
	}
	if got[1].Slot != 1 {
		t.Errorf("the changed ref moved clause: %+v", got[1])
	}
	if findRetargetPrompt(g, opp.ID) != nil {
		t.Error("the prompt is still open after a legal answer")
	}
}

// "If changing one target … to Spellskite would make other targets of
// that spell or ability illegal, that target can't be changed": a
// clause that already names Skite cannot name it twice, and "another
// target creature" cannot be Skite when the first slot is.
func TestPinnedRetargetRefusesAChangeThatBreaksAnotherSlot(t *testing.T) {
	t.Run("one clause, two picks", func(t *testing.T) {
		g := newActiveGame(t)
		pinnedSpecs(t)
		advanceTo(t, g, StepPrecombatMain)
		me, opp := g.Seats[0], g.Seats[1]
		a := pushRetargetCreature(g, opp, "A")
		skite := pushRetargetCreature(g, opp, "Skite")
		item := castTargeted(t, g, me, pinnedTwoOfOneClauseOracle, []TargetRef{
			{Kind: TargetCard, ID: a}, {Kind: TargetCard, ID: skite},
		})
		offerPinned(t, g, item, opp.ID, skite, false)
		if findRetargetPrompt(g, opp.ID) != nil {
			t.Error("a prompt opened with nothing changeable")
		}
		if got := targetIDs(g, item.ID); got[0] != a || got[1] != skite {
			t.Errorf("targets = %v, want unchanged [A Skite]", got)
		}
	})
	t.Run("another target creature", func(t *testing.T) {
		g := newActiveGame(t)
		pinnedSpecs(t)
		advanceTo(t, g, StepPrecombatMain)
		me, opp := g.Seats[0], g.Seats[1]
		b := pushRetargetCreature(g, opp, "B")
		skite := pushRetargetCreature(g, opp, "Skite")
		// Skite is the FIRST slot; the second says "another", so it
		// cannot become Skite, and the first already is.
		item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
			{Kind: TargetCard, ID: skite, Slot: 0}, {Kind: TargetCard, ID: b, Slot: 1},
		})
		offerPinned(t, g, item, opp.ID, skite, false)
		if got := targetIDs(g, item.ID); got[0] != skite || got[1] != b {
			t.Errorf("targets = %v, want unchanged [Skite B]", got)
		}
	})
	t.Run("only the slot that stays legal is changed", func(t *testing.T) {
		g := newActiveGame(t)
		pinnedSpecs(t)
		advanceTo(t, g, StepPrecombatMain)
		me, opp := g.Seats[0], g.Seats[1]
		a := pushRetargetCreature(g, opp, "A")
		b := pushRetargetCreature(g, opp, "B")
		skite := pushRetargetCreature(g, opp, "Skite")
		item := castTargeted(t, g, me, pinnedTwoOfOneClauseOracle, []TargetRef{
			{Kind: TargetCard, ID: a}, {Kind: TargetCard, ID: b},
		})
		// Both slots could become Skite on their own; the chooser picks.
		offerPinned(t, g, item, opp.ID, skite, false)
		prompt := findRetargetPrompt(g, opp.ID)
		if prompt == nil || len(prompt.PickTargetCards) != 2 {
			t.Fatalf("prompt = %+v, want one over both targets", prompt)
		}
		if err := g.ResolveRetarget(prompt.ID, opp.ID, []TargetRef{pinnedTo(a)}); err != nil {
			t.Fatalf("ResolveRetarget: %v", err)
		}
		if got := targetIDs(g, item.ID); got[0] != skite || got[1] != b {
			t.Errorf("targets = %v, want [Skite B]", got)
		}
	})
}

// Two instances of "target" that named the SAME creature are one
// object on the board: a mandatory change with nothing else eligible is
// forced, and lands on the earlier slot. This is Spellskite's declared
// caveat, pinned so that changing it is a decision.
func TestPinnedRetargetSameObjectInTwoClausesChangesTheEarlier(t *testing.T) {
	g := newActiveGame(t)
	pinnedSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	skite := pushRetargetCreature(g, opp, "Skite")
	item := castTargeted(t, g, me, pinnedTwoClausesOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0}, {Kind: TargetCard, ID: a, Slot: 1},
	})
	offerPinned(t, g, item, opp.ID, skite, false)
	if findRetargetPrompt(g, opp.ID) != nil {
		t.Error("one object to offer, and a mandatory change: nothing to ask")
	}
	if got := targetIDs(g, item.ID); got[0] != skite || got[1] != a {
		t.Errorf("targets = %v, want [Skite A]", got)
	}
}

// A "you may" (Mizzium Meddler) asks even over one target, and the
// decline leaves it alone.
func TestPinnedRetargetYouMayAsksAndMayBeDeclined(t *testing.T) {
	g := newActiveGame(t)
	pinnedSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushRetargetCreature(g, opp, "Victim")
	skite := pushRetargetCreature(g, opp, "Meddler")

	item := castTargeted(t, g, me, retargetOneOracle, []TargetRef{{Kind: TargetCard, ID: victim}})
	offerPinned(t, g, item, opp.ID, skite, true)
	prompt := findRetargetPrompt(g, opp.ID)
	if prompt == nil {
		t.Fatal(`a "you may" opened no prompt`)
	}
	if prompt.PickTargetMin != 0 || len(prompt.PickTargetCards) != 1 || prompt.PickTargetCards[0] != victim {
		t.Errorf("prompt = min %d over %v, want min 0 over [Victim]", prompt.PickTargetMin, prompt.PickTargetCards)
	}
	if err := g.ResolveRetarget(prompt.ID, opp.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if got := targetIDs(g, item.ID); got[0] != victim {
		t.Errorf("targets = %v, want unchanged", got)
	}
	if findRetargetPrompt(g, opp.ID) != nil {
		t.Error("the prompt is still open after a decline")
	}
}

// The destination leaves while the prompt is open: nothing can move any
// more, so the answer drops the prompt and changes nothing (CR 115.7a)
// rather than wedging the table on a question with no legal answer.
func TestPinnedRetargetDestinationLeavesUnderAnOpenPrompt(t *testing.T) {
	g := newActiveGame(t)
	pinnedSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	skite := pushRetargetCreature(g, opp, "Skite")
	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0}, {Kind: TargetCard, ID: b, Slot: 1},
	})
	offerPinned(t, g, item, opp.ID, skite, false)
	prompt := findRetargetPrompt(g, opp.ID)
	if prompt == nil {
		t.Fatal("no prompt")
	}
	g.WithWriteLock(func() {
		c, err := g.Battlefield.Remove(skite)
		if err != nil {
			t.Fatalf("remove: %v", err)
		}
		opp.Graveyard.PushTop(c)
	})
	if err := g.ResolveRetarget(prompt.ID, opp.ID, []TargetRef{pinnedTo(a)}); err != nil {
		t.Fatalf("ResolveRetarget after the destination left = %v, want nil", err)
	}
	if got := targetIDs(g, item.ID); got[0] != a || got[1] != b {
		t.Errorf("targets = %v, want unchanged [A B]", got)
	}
	if findRetargetPrompt(g, opp.ID) != nil {
		t.Error("the prompt should be dropped")
	}
}

// The item leaves between the activation and its resolution: the offer
// reports ErrCardNotFound, which ChangeTargets absorbs as "nothing
// happened". And a pinned offer only speaks "change the target".
func TestPinnedRetargetItemGoneAndPolicy(t *testing.T) {
	g := newActiveGame(t)
	pinnedSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushRetargetCreature(g, opp, "Victim")
	skite := pushRetargetCreature(g, opp, "Skite")
	item := castTargeted(t, g, me, retargetOneOracle, []TargetRef{{Kind: TargetCard, ID: victim}})

	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: opp.ID, Policy: RetargetChooseNew, To: pinnedTo(skite),
		}); err != ErrInvalidParam {
			t.Errorf("pinned under ChooseNew = %v, want ErrInvalidParam", err)
		}
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID: uuid.New(), Chooser: opp.ID, Policy: RetargetChangeOne, To: pinnedTo(skite),
		}); err != ErrCardNotFound {
			t.Errorf("pinned at an item no longer on the stack = %v, want ErrCardNotFound", err)
		}
	})
	if got := targetIDs(g, item.ID); got[0] != victim {
		t.Errorf("targets = %v, want unchanged", got)
	}
}

// A game paused on a pinned prompt is a restorable snapshot, and the
// restored prompt is still the pinned question: answering with a
// current target moves THAT slot to the destination, rather than
// (as a free retarget would read it) moving the asked slot onto the
// object clicked.
func TestPinnedRetargetPromptSurvivesARestore(t *testing.T) {
	g := newActiveGame(t)
	pinnedSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	skite := pushRetargetCreature(g, opp, "Skite")
	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0}, {Kind: TargetCard, ID: b, Slot: 1},
	})
	offerPinned(t, g, item, opp.ID, skite, false)

	_, restored := roundTrip(t, g)
	prompt := findRetargetPrompt(restored, opp.ID)
	if prompt == nil {
		t.Fatal("the pinned prompt did not survive the restore")
	}
	if prompt.RetargetTo != pinnedTo(skite) {
		t.Fatalf("restored RetargetTo = %+v, want Skite", prompt.RetargetTo)
	}
	if err := restored.ResolveRetarget(prompt.ID, opp.ID, []TargetRef{pinnedTo(b)}); err != nil {
		t.Fatalf("ResolveRetarget on the restored game: %v", err)
	}
	if got := targetIDs(restored, item.ID); got[0] != a || got[1] != skite {
		t.Errorf("targets = %v, want [A Skite]", got)
	}
}
