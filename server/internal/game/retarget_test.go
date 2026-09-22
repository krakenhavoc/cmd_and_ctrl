package game

import (
	"testing"

	"github.com/google/uuid"
)

// retarget_test.go — CR 115.7 (#1196). The rules this file pins are
// the ones a reader of retarget.go would want proved: the redirect
// lands, an illegal one does not, protection and hexproof are
// re-checked as they are at announce, "you may" is answerable with
// nothing, every slot can move under "choose new targets", and the
// division of damage travels with the slot rather than staying on the
// old target's key.

const (
	retargetBoltOracle = "test-retarget-bolt"
	retargetPairOracle = "test-retarget-two-creatures"
	retargetOneOracle  = "test-retarget-one-creature"
)

// anyTargetSpec is Lightning Bolt's clause: any target, one of them.
func anyTargetSpec() *TargetSpec {
	return &TargetSpec{
		Mode:    "any",
		Label:   "any target",
		Players: true,
		Zones:   []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
}

// oneCreatureSpec is "target creature" — no players, so a board with
// exactly one creature on it has no alternative to offer.
func oneCreatureSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "creature",
		Label: "target creature",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
}

// twoCreatureSpec is a two-slot statement: one creature per clause,
// the second distinct from the first, which is the shape "divide N
// damage among two target creatures" announces under.
func twoCreatureSpec() *TargetSpec {
	first := &TargetSpec{
		Mode:  "creature",
		Label: "target creature",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
	second := *first
	second.Label = "another target creature"
	second.Distinct = true
	return first.Then(&second)
}

// retargetSpecs installs the two clauses above under their oracle
// ids for the test's duration.
func retargetSpecs(t *testing.T) {
	t.Helper()
	withCatalogTargetSpec(t, func(oracleID string) *TargetSpec {
		switch oracleID {
		case retargetBoltOracle:
			return anyTargetSpec()
		case retargetPairOracle:
			return twoCreatureSpec()
		case retargetOneOracle:
			return oneCreatureSpec()
		}
		return nil
	})
}

// castTargeted puts a spell with `oracle`'s clause on the stack,
// aimed at `targets`, and returns its item.
func castTargeted(t *testing.T, g *Game, caster *Player, oracle string, targets []TargetRef) *StackItem {
	t.Helper()
	c := NewCard("Test Spell", caster.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{0}"
	c.OracleID = oracle
	c.Controller = caster.ID
	caster.Hand.PushTop(c)
	if err := g.CastSpell(caster.ID, c.InstanceID, CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	item := g.StackMeta[c.InstanceID]
	if item == nil {
		t.Fatalf("spell did not reach the stack")
	}
	return item
}

func pushRetargetCreature(g *Game, owner *Player, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Test"
	c.Power, c.Toughness = 2, 2
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func findRetargetPrompt(g *Game, chooser uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceRetarget && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// A Bolt aimed at one creature is redirected to another, and that is
// where it resolves. The whole seam in one test.
func TestRetargetMovesTheSpellToTheNewTarget(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	first := pushRetargetCreature(g, opp, "First")
	second := pushRetargetCreature(g, opp, "Second")

	item := castTargeted(t, g, me, retargetBoltOracle,
		[]TargetRef{{Kind: TargetCard, ID: first}})

	// The retargeting effect's controller is the OPPONENT here — the
	// seat that did not cast the spell — which is the case the seam
	// exists for.
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID:  item.ID,
			Chooser: opp.ID,
			Policy:  RetargetChangeOne,
			Reason:  "Bolt Bend",
		}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
	prompt := findRetargetPrompt(g, opp.ID)
	if prompt == nil {
		t.Fatal("no retarget prompt for the retargeting effect's controller")
	}
	// The target already there is not one of its own alternatives:
	// "changed only to ANOTHER legal target" (CR 115.7a).
	for _, id := range prompt.PickTargetCards {
		if id == first {
			t.Error("the current target is offered as an alternative")
		}
	}
	if !hasUUID(prompt.PickTargetCards, second) {
		t.Fatalf("alternatives = %v, want the second creature", prompt.PickTargetCards)
	}
	if prompt.PickTargetMin != 1 {
		t.Errorf("a mandatory change offers min %d, want 1", prompt.PickTargetMin)
	}

	if err := g.ResolveRetarget(prompt.ID, opp.ID,
		[]TargetRef{{Kind: TargetCard, ID: second}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	if got := g.StackMeta[item.ID].Targets; len(got) != 1 || got[0].ID != second {
		t.Fatalf("item targets = %+v, want the second creature", got)
	}
	if findRetargetPrompt(g, opp.ID) != nil {
		t.Error("the prompt should be gone once it is answered")
	}
	// CR 115.7: the new pick has become the target of the spell, so a
	// ward or Monk Gyatso sees it — attributed to the SPELL's
	// controller, not to the seat that moved it.
	var saw bool
	for _, ev := range g.Events {
		if ev.Kind == EventBecomesTarget && ev.Target == second {
			saw = true
			if ev.Actor != me.ID {
				t.Errorf("EventBecomesTarget actor = %v, want the spell's controller", ev.Actor)
			}
		}
	}
	if !saw {
		t.Error("no EventBecomesTarget for the new target")
	}
}

// "a target may only be changed to a legal one": an answer outside
// the clause is refused and the prompt stays open.
func TestRetargetRefusesAnIllegalNewTarget(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	first := pushRetargetCreature(g, opp, "First")
	rock := pushArtifact(g, opp, "Rock")

	item := castTargeted(t, g, me, retargetBoltOracle,
		[]TargetRef{{Kind: TargetCard, ID: first}})
	g.WithWriteLock(func() {
		_ = g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: me.ID, Policy: RetargetChangeOne,
		})
	})
	prompt := findRetargetPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no retarget prompt")
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID,
		[]TargetRef{{Kind: TargetCard, ID: rock}}); err != ErrIllegalTarget {
		t.Fatalf("artifact under a creature clause: got %v, want ErrIllegalTarget", err)
	}
	if findRetargetPrompt(g, me.ID) == nil {
		t.Error("a refused answer must leave the prompt open")
	}
	if err := g.ResolveRetarget(prompt.ID, opp.ID,
		[]TargetRef{{Kind: TargetCard, ID: first}}); err != ErrNotTheChooser {
		t.Fatalf("wrong chooser: got %v, want ErrNotTheChooser", err)
	}
	if got := g.StackMeta[item.ID].Targets; got[0].ID != first {
		t.Errorf("a refused retarget must not move the target: %+v", got)
	}
}

// The announce gate's protection / hexproof check is the retarget's
// too, and it is judged for the ITEM's controller rather than for the
// seat doing the redirecting — the whole point of §2.
func TestRetargetRefusesAHexproofOrProtectedTarget(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	first := pushRetargetCreature(g, opp, "First")
	hexproof := pushRetargetCreature(g, opp, "Hexproof One")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == hexproof {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "hexproof")
			}
		}
	})

	item := castTargeted(t, g, me, retargetBoltOracle,
		[]TargetRef{{Kind: TargetCard, ID: first}})
	g.WithWriteLock(func() {
		_ = g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: me.ID, Policy: RetargetChangeOne,
		})
	})
	prompt := findRetargetPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no retarget prompt")
	}
	if hasUUID(prompt.PickTargetCards, hexproof) {
		t.Error("a hexproof creature an opponent controls reached the retarget picker")
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID,
		[]TargetRef{{Kind: TargetCard, ID: hexproof}}); err != ErrIllegalTarget {
		t.Fatalf("hexproof: got %v, want ErrIllegalTarget", err)
	}
}

// CR 115.7c's "you may": the answer to an optional prompt can be
// nothing at all, and the target stays where it was.
func TestRetargetYouMayCanBeDeclined(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	first := pushRetargetCreature(g, opp, "First")
	pushRetargetCreature(g, opp, "Second")

	item := castTargeted(t, g, me, retargetBoltOracle,
		[]TargetRef{{Kind: TargetCard, ID: first}})
	g.WithWriteLock(func() {
		_ = g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: me.ID, Policy: RetargetChangeOne, Optional: true,
		})
	})
	prompt := findRetargetPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no retarget prompt")
	}
	if prompt.PickTargetMin != 0 {
		t.Errorf(`a "you may" offers min %d, want 0`, prompt.PickTargetMin)
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if got := g.StackMeta[item.ID].Targets; len(got) != 1 || got[0].ID != first {
		t.Errorf("a declined retarget must leave the target alone: %+v", got)
	}
	if findRetargetPrompt(g, me.ID) != nil {
		t.Error("the prompt should be gone after a decline")
	}
}

// A MANDATORY change cannot be declined while there is somewhere for
// the target to go.
func TestRetargetMandatoryChangeRefusesADecline(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	first := pushRetargetCreature(g, opp, "First")
	pushRetargetCreature(g, opp, "Second")

	item := castTargeted(t, g, me, retargetBoltOracle,
		[]TargetRef{{Kind: TargetCard, ID: first}})
	g.WithWriteLock(func() {
		_ = g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: me.ID, Policy: RetargetChangeOne,
		})
	})
	prompt := findRetargetPrompt(g, me.ID)
	if err := g.ResolveRetarget(prompt.ID, me.ID, nil); err != ErrInvalidParam {
		t.Fatalf("declining a mandatory change: got %v, want ErrInvalidParam", err)
	}
}

// CR 115.7a: a target with nowhere else legal to go is unchanged, and
// no prompt is opened for it — a prompt with no answers is a table
// that cannot move (#791).
func TestRetargetWithNoAlternativeAsksNothing(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	only := pushRetargetCreature(g, opp, "The Only One")

	item := castTargeted(t, g, me, retargetOneOracle,
		[]TargetRef{{Kind: TargetCard, ID: only}})
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: me.ID, Policy: RetargetChangeOne,
		}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
	if findRetargetPrompt(g, me.ID) != nil {
		t.Error("a target with no alternative must not open a prompt")
	}
	if got := g.StackMeta[item.ID].Targets; got[0].ID != only {
		t.Errorf("target moved with nowhere to move to: %+v", got)
	}
}

// A spell with no targets cannot be retargeted (CR 115.7 has nothing
// to change), and the engine says so rather than pretending.
func TestRetargetRefusesASpellWithNoTargets(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	c := NewCard("Wrath of Test", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{0}"
	c.Controller = me.ID
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	g.WithWriteLock(func() {
		if g.RetargetableForEffect(c.InstanceID) {
			t.Error("a spell with no targets reports as retargetable")
		}
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID: c.InstanceID, Chooser: me.ID, Policy: RetargetChooseNew,
		}); err != ErrInvalidParam {
			t.Errorf("offering a retarget for an untargeted spell = %v, want ErrInvalidParam", err)
		}
		if err := g.RetargetStackItemForEffect(c.InstanceID, me.ID, RetargetChooseNew, nil); err != ErrInvalidParam {
			t.Errorf("retargeting an untargeted spell = %v, want ErrInvalidParam", err)
		}
	})
}

// CR 115.7c: "choose new targets" walks EVERY slot, one prompt at a
// time, and each one may be changed or left alone.
func TestChooseNewTargetsWalksEverySlot(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	c := pushRetargetCreature(g, opp, "C")

	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0},
		{Kind: TargetCard, ID: b, Slot: 1},
	})
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: opp.ID, Policy: RetargetChooseNew, Optional: true,
		}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})

	// Slot 0 first, and its alternatives exclude both the target
	// already there and the OTHER slot's pick, because clause 2 is
	// Distinct.
	first := findRetargetPrompt(g, opp.ID)
	if first == nil || first.RetargetSlot != 0 {
		t.Fatalf("first prompt = %+v, want slot 0", first)
	}
	if hasUUID(first.PickTargetCards, a) {
		t.Error("slot 0 offers the target already in it")
	}
	if err := g.ResolveRetarget(first.ID, opp.ID,
		[]TargetRef{{Kind: TargetCard, ID: c}}); err != nil {
		t.Fatalf("slot 0: %v", err)
	}

	// Slot 1 next, and declining it leaves B where it was.
	second := findRetargetPrompt(g, opp.ID)
	if second == nil || second.RetargetSlot != 1 {
		t.Fatalf("second prompt = %+v, want slot 1", second)
	}
	if err := g.ResolveRetarget(second.ID, opp.ID, nil); err != nil {
		t.Fatalf("slot 1 decline: %v", err)
	}
	if findRetargetPrompt(g, opp.ID) != nil {
		t.Error("the walk should be over after the last slot")
	}
	got := g.StackMeta[item.ID].Targets
	if len(got) != 2 || got[0].ID != c || got[1].ID != b {
		t.Errorf("targets = %+v, want [C B]", got)
	}
	if got[0].Slot != 0 || got[1].Slot != 1 {
		t.Errorf("a retarget must not move a ref between clauses: %+v", got)
	}
}

// "Change the target" is ONE slot: an entry-point call that moves two
// is refused outright.
func TestChangeOneRefusesTwoChangedSlots(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	c := pushRetargetCreature(g, opp, "C")
	d := pushRetargetCreature(g, opp, "D")

	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0},
		{Kind: TargetCard, ID: b, Slot: 1},
	})
	both := []TargetRef{
		{Kind: TargetCard, ID: c, Slot: 0},
		{Kind: TargetCard, ID: d, Slot: 1},
	}
	g.WithWriteLock(func() {
		if err := g.RetargetStackItemForEffect(item.ID, opp.ID, RetargetChangeOne, both); err != ErrInvalidParam {
			t.Errorf("two slots under ChangeOne = %v, want ErrInvalidParam", err)
		}
		// The same two moves are fine under "choose new targets".
		if err := g.RetargetStackItemForEffect(item.ID, opp.ID, RetargetChooseNew, both); err != nil {
			t.Errorf("two slots under ChooseNew = %v, want nil", err)
		}
		// And a list that is not one-for-one is refused under either.
		short := []TargetRef{{Kind: TargetCard, ID: a, Slot: 0}}
		if err := g.RetargetStackItemForEffect(item.ID, opp.ID, RetargetChooseNew, short); err != ErrInvalidParam {
			t.Errorf("dropping a slot = %v, want ErrInvalidParam", err)
		}
	})
	// The offer refuses a "change the target" over a multi-target
	// item, which is what the printed "with a single target" means.
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: opp.ID, Policy: RetargetChangeOne,
		}); err != ErrInvalidParam {
			t.Errorf(`"change the target" over two targets = %v, want ErrInvalidParam`, err)
		}
	})
}

// CR 115.7c: the division can't be changed, so a slot's portion
// travels with the slot rather than staying on the old target's key.
func TestRetargetPreservesTheDivisionOfDamage(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	c := pushRetargetCreature(g, opp, "C")

	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0},
		{Kind: TargetCard, ID: b, Slot: 1},
	})
	g.WithWriteLock(func() {
		item.Distribution = map[uuid.UUID]int{a: 3, b: 1}
		item.XValue = 4
		item.Paid.OnPaper = true
		if err := g.RetargetStackItemForEffect(item.ID, opp.ID, RetargetChooseNew, []TargetRef{
			{Kind: TargetCard, ID: c, Slot: 0},
			{Kind: TargetCard, ID: b, Slot: 1},
		}); err != nil {
			t.Fatalf("RetargetStackItemForEffect: %v", err)
		}
	})
	dist := g.StackMeta[item.ID].Distribution
	if dist[c] != 3 || dist[b] != 1 {
		t.Errorf("distribution = %v, want C:3 B:1 — the portion travels with the slot", dist)
	}
	if _, stale := dist[a]; stale {
		t.Errorf("the old target kept its portion: %v", dist)
	}
	// Everything else about the announcement survives.
	if g.StackMeta[item.ID].XValue != 4 || !g.StackMeta[item.ID].Paid.OnPaper {
		t.Error("a retarget must not touch X or the payment record")
	}
}

// The prompt's two obligations: it stops the table, and it is dropped
// rather than reassigned when the seat that owes it leaves.
func TestRetargetPromptIsClassifiedAndNotReassigned(t *testing.T) {
	if !ChoiceBlocksTable(PendingChoiceRetarget) {
		t.Error("a retarget prompt must stop the table — the item it redirects is still on the stack")
	}
	for _, kind := range ReassignableChoiceKinds() {
		if kind != PendingChoiceRetarget {
			continue
		}
		if choiceDepartureDecisions[kind].reassign {
			t.Error("a retarget must not be reassigned: nobody else aims another player's Deflecting Swat")
		}
		if choiceDepartureDecisions[kind].onDrop != dropDiscard {
			t.Error("a dropped retarget has nothing left to do — the targets are simply unchanged")
		}
		return
	}
	t.Fatal("PendingChoiceRetarget has no row in the departure table")
}

func hasUUID(ids []uuid.UUID, want uuid.UUID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

// CR 115.7c's other half: a target left UNCHANGED may stay even
// though it would now be illegal. The announce gate would refuse it,
// which is why the retarget runs that gate with the unchanged slots'
// predicate waived.
func TestChooseNewTargetsMayLeaveAnIllegalTargetAlone(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	c := pushRetargetCreature(g, opp, "C")

	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0},
		{Kind: TargetCard, ID: b, Slot: 1},
	})
	// In response, B stops being a creature: still on the battlefield,
	// still the item's target, no longer a legal one.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == b {
				g.Battlefield.Cards[i].TypeLine = "Enchantment"
			}
		}
		// Moving slot 0 while slot 1 keeps its now-illegal target is
		// legal (CR 115.7c).
		if err := g.RetargetStackItemForEffect(item.ID, opp.ID, RetargetChooseNew, []TargetRef{
			{Kind: TargetCard, ID: c, Slot: 0},
			{Kind: TargetCard, ID: b, Slot: 1},
		}); err != nil {
			t.Fatalf("leaving an illegal target alone = %v, want nil", err)
		}
		// CHANGING a slot to that same illegal object is refused —
		// the waiver is for what stays, not for what moves.
		if err := g.RetargetStackItemForEffect(item.ID, opp.ID, RetargetChooseNew, []TargetRef{
			{Kind: TargetCard, ID: b, Slot: 0},
			{Kind: TargetCard, ID: b, Slot: 1},
		}); err != ErrIllegalTarget {
			t.Fatalf("moving a slot ONTO an illegal object = %v, want ErrIllegalTarget", err)
		}
	})
	if got := g.StackMeta[item.ID].Targets; got[0].ID != c || got[1].ID != b {
		t.Errorf("targets = %+v, want [C B]", got)
	}
}

// CR 707.10c is CR 115.7c by reference, so the copy's "you may choose
// new targets" answer goes through the same check — which means a
// target the player left alone may stay even though it has since
// become illegal, and the NUMBER of targets may not change.
func TestCopyRetargetUsesTheSharedCR115Check(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushRetargetCreature(g, opp, "Victim")
	other := pushRetargetCreature(g, opp, "Other")

	item := castTargeted(t, g, me, retargetOneOracle,
		[]TargetRef{{Kind: TargetCard, ID: victim}})
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(item.ID, me.ID, true, nil); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
	})
	prompt := findPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no CR 707.10c re-target prompt for the copy")
	}
	// Two refs where the original announced one: CR 115.7 never
	// changes how many targets a spell has.
	if err := g.ResolvePickTargets(prompt.ID, me.ID, []TargetRef{
		{Kind: TargetCard, ID: victim},
		{Kind: TargetCard, ID: other},
	}); err != ErrInvalidParam {
		t.Errorf("a copy re-targeted onto two targets = %v, want ErrInvalidParam", err)
	}
	// The original's target stops being a creature. Leaving it alone
	// is still a legal answer (CR 115.7c), and the copy is created.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].TypeLine = "Enchantment"
			}
		}
	})
	if err := g.ResolvePickTargets(prompt.ID, me.ID,
		[]TargetRef{{Kind: TargetCard, ID: victim}}); err != nil {
		t.Fatalf("keeping the copy's original target = %v, want nil", err)
	}
	var copied *StackItem
	for id, it := range g.StackMeta {
		if id != item.ID {
			copied = it
		}
	}
	if copied == nil || !copied.IsCopy {
		t.Fatal("the copy was not created")
	}
	if len(copied.Targets) != 1 || copied.Targets[0].ID != victim {
		t.Errorf("copy targets = %+v, want the original's", copied.Targets)
	}
}
