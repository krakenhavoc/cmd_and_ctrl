package game

import (
	"testing"

	"github.com/google/uuid"
)

// spell_control_test.go pins ADR 0104 (#1745): control of a SPELL is a
// layer-2 effect (CR 611.1, CR 613.1b) pinned to the stack object, and
// every rule the ADR names has a test here. The card half — Divination
// drawing for the thief, the retarget offer, "if you cast it" — is in
// cards/effects/spell_control_test.go.

// castSpellForControlTest puts a {0} spell of `typeLine` in p's hand
// and casts it (permissive), returning its stack item.
func castSpellForControlTest(t *testing.T, g *Game, p *Player, typeLine string) *StackItem {
	t.Helper()
	c := NewCard("Stolen Goods", p.ID)
	c.TypeLine = typeLine
	c.ManaCost = "{0}"
	c.Controller = p.ID
	if c.IsCreature() {
		c.Power, c.Toughness = 2, 2
	}
	var item *StackItem
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	if err := g.CastSpell(p.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", typeLine, err)
	}
	g.ReadSnapshot(func() { item = g.StackMeta[c.InstanceID] })
	if item == nil {
		t.Fatal("no stack item for the cast spell")
	}
	return item
}

// stealSpell runs GainControlOfSpellForEffect under the write lock.
func stealSpell(t *testing.T, g *Game, spell, thief uuid.UUID) bool {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.GainControlOfSpellForEffect(uuid.New(), spell, thief, "test — Aethersnatch")
	})
	return ok
}

func stackCardControllerForTest(t *testing.T, g *Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	var out uuid.UUID
	g.ReadSnapshot(func() {
		if c, ok := g.stackCardLocked(id); ok {
			out = c.Controller
		}
	})
	return out
}

func countEventsForTest(g *Game, kind EventKind, card uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind && ev.CardID == card {
			n++
		}
	}
	return n
}

func spellControlPermanentForTest(t *testing.T, g *Game, id uuid.UUID) Card {
	t.Helper()
	var out Card
	var found bool
	g.ReadSnapshot(func() {
		if c, ok := g.battlefieldCardLocked(id); ok {
			out, found = *c, true
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return out
}

// TestCastStampsTheSpellCardWithItsCaster is the first latent bug the
// ADR found: a card carries its OWNER as its controller from the deck
// load, and only a face-down cast re-stamped the stack card, so a
// spell cast off another player's card answered "target spell you
// don't control" by its owner. CR 601.2a: the caster becomes its
// controller, and the card on the stack says so.
func TestCastStampsTheSpellCardWithItsCaster(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	c := NewCard("Borrowed Bolt", opp.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{0}"
	c.Controller = opp.ID // as the deck load stamps it
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if got := stackCardControllerForTest(t, g, c.InstanceID); got != me.ID {
		t.Errorf("stack card controller = %s, want the caster %s (CR 601.2a)", got, me.ID)
	}
}

// TestActOfTreasonRevertsWhenTheThiefLeaves is the second latent bug:
// CR 800.4a ends every effect that gives a departed player control,
// and a resolving spell's theft is a ScopedEffect record that nothing
// ended. The rule's own example: Runeclaw Bears reverts to Bianca.
func TestActOfTreasonRevertsWhenTheThiefLeaves(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	owner, thief := g.Seats[1], g.Seats[2]
	bears := pushScopedTestCreature(g, owner.ID, 2, 2)
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), bears, thief.ID, IndefiniteDuration(), "test — Act of Treason") {
			t.Fatal("GainControlForEffect refused")
		}
	})
	if got := controllerOfCard(t, g, bears); got != thief.ID {
		t.Fatalf("setup: controller %s, want the thief %s", got, thief.ID)
	}
	if err := g.Concede(thief.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if !g.Battlefield.Contains(bears) {
		t.Fatalf("the Bears left the battlefield (exile: %v) — CR 800.4a ends the thief's effect first, so nothing is still under them",
			g.Exile.Contains(bears))
	}
	if got := controllerOfCard(t, g, bears); got != owner.ID {
		t.Errorf("controller after the thief left = %s, want %s", got, owner.ID)
	}
}

// TestGainControlOfSpellChangesItsController is the primitive: the
// item, the stack card, the default controller and the event.
func TestGainControlOfSpellChangesItsController(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	item := castSpellForControlTest(t, g, me, "Instant")

	if !stealSpell(t, g, item.ID, opp.ID) {
		t.Fatal("GainControlOfSpellForEffect refused a spell on the stack")
	}
	if item.Controller != opp.ID {
		t.Errorf("item controller = %s, want the thief %s", item.Controller, opp.ID)
	}
	if item.BaseController != me.ID {
		t.Errorf("item base = %s, want the caster %s (CR 110.2b)", item.BaseController, me.ID)
	}
	if got := stackCardControllerForTest(t, g, item.ID); got != opp.ID {
		t.Errorf("stack card controller = %s, want %s", got, opp.ID)
	}
	var ev *Event
	for i := range g.Events {
		if g.Events[i].Kind == EventSpellControlChanged && g.Events[i].CardID == item.ID {
			ev = &g.Events[i]
		}
	}
	if ev == nil {
		t.Fatal("no EventSpellControlChanged")
	}
	if ev.Actor != opp.ID || ev.Target != me.ID {
		t.Errorf("event actor/target = %s/%s, want gained %s / lost %s", ev.Actor, ev.Target, opp.ID, me.ID)
	}
	if n := countEventsForTest(g, EventControlChanged, item.ID); n != 0 {
		t.Errorf("a spell emitted %d EventControlChanged — that kind is about permanents", n)
	}
}

// TestGainControlOfSpellRefusals: no spell, the player already has
// it, a player who has left (CR 800.4b), an ability.
func TestGainControlOfSpellRefusals(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	advanceTo(t, g, StepPrecombatMain)
	me, opp, gone := g.Seats[0], g.Seats[1], g.Seats[2]
	item := castSpellForControlTest(t, g, me, "Instant")
	if err := g.Concede(gone.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if stealSpell(t, g, uuid.New(), opp.ID) {
		t.Error("stole a spell that is not on the stack")
	}
	if stealSpell(t, g, item.ID, me.ID) {
		t.Error("stole a spell for the player who already controls it")
	}
	if stealSpell(t, g, item.ID, gone.ID) {
		t.Error("CR 800.4b: gave a spell to a player who has left the game")
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("refused steals registered %d records", n)
	}
}

// TestAStackPinTakesOnlyAControlMod — the stack step applies layer 2
// alone, so registration refuses anything else on a spell.
func TestAStackPinTakesOnlyAControlMod(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	item := castSpellForControlTest(t, g, g.Seats[0], "Instant")
	defer func() {
		if recover() == nil {
			t.Error("a keyword mod pinned to a spell was registered")
		}
	}()
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.New(), []AffectedObject{PinStackObject(item.ID, 1)},
			[]Mod{AddKeywordsMod("flying")}, IndefiniteDuration(), "test")
	})
}

// TestStolenPermanentSpellEntersUnderTheThief is CR 110.2b and
// CR 400.7a: the thief controls the permanent, its default controller
// is the caster, the record now follows the permanent, and the entry
// itself fires no control-change event.
func TestStolenPermanentSpellEntersUnderTheThief(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	item := castSpellForControlTest(t, g, me, "Creature — Bear")
	stealSpell(t, g, item.ID, opp.ID)

	resolveTop(t, g)
	g.ReadSnapshot(func() {})

	perm := spellControlPermanentForTest(t, g, item.ID)
	if perm.Controller != opp.ID {
		t.Errorf("permanent controller = %s, want the thief %s (CR 110.2b)", perm.Controller, opp.ID)
	}
	if perm.BaseController != me.ID {
		t.Errorf("permanent base = %s, want the caster %s (CR 110.2b)", perm.BaseController, me.ID)
	}
	if n := countEventsForTest(g, EventControlChanged, item.ID); n != 0 {
		t.Errorf("the entry emitted %d EventControlChanged — its control had already settled", n)
	}
	if len(g.ScopedEffects) != 1 {
		t.Fatalf("registry holds %d records, want the one re-pinned theft", len(g.ScopedEffects))
	}
	a := g.ScopedEffects[0].Affected
	if len(a) != 1 || a[0].OnStack || a[0].ID != item.ID || a[0].Epoch != perm.ObjectEpoch {
		t.Errorf("record pin = %+v, want the permanent by its epoch %d", a, perm.ObjectEpoch)
	}
	if prov := perm.Provenance; prov.Caster != me.ID {
		t.Errorf("provenance caster = %s, want %s", prov.Caster, me.ID)
	}
	if cast, known := perm.CastByItsController(); !known || cast {
		t.Errorf("CastByItsController = %v (known %v), want false: the thief did not cast it", cast, known)
	}
}

// TestUnstolenPermanentKnowsItsControllerCastIt — the common case of
// the same record.
func TestUnstolenPermanentKnowsItsControllerCastIt(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	item := castSpellForControlTest(t, g, me, "Creature — Bear")
	resolveTop(t, g)
	perm := spellControlPermanentForTest(t, g, item.ID)
	if cast, known := perm.CastByItsController(); !known || !cast {
		t.Errorf("CastByItsController = %v (known %v), want true", cast, known)
	}
	if perm.BaseController != uuid.Nil && perm.BaseController != me.ID {
		t.Errorf("base = %s, want the caster", perm.BaseController)
	}
}

// TestStolenPermanentGoesHomeWhenTheThiefLeaves — CR 800.4a after the
// hand-off: the permanent reverts to its default controller.
func TestStolenPermanentGoesHomeWhenTheThiefLeaves(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	advanceTo(t, g, StepPrecombatMain)
	me, thief := g.Seats[0], g.Seats[2]
	item := castSpellForControlTest(t, g, me, "Creature — Bear")
	stealSpell(t, g, item.ID, thief.ID)
	resolveTop(t, g)
	if err := g.Concede(thief.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if !g.Battlefield.Contains(item.ID) {
		t.Fatalf("the permanent left the battlefield when the thief did")
	}
	if got := controllerOfCard(t, g, item.ID); got != me.ID {
		t.Errorf("controller = %s, want the caster %s", got, me.ID)
	}
}

// TestStolenSpellGoesBackWhenTheThiefLeaves — CR 800.4a on the stack:
// the effect ends before anything the thief still controls is exiled.
func TestStolenSpellGoesBackWhenTheThiefLeaves(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	advanceTo(t, g, StepPrecombatMain)
	me, thief := g.Seats[0], g.Seats[2]
	item := castSpellForControlTest(t, g, me, "Instant")
	stealSpell(t, g, item.ID, thief.ID)
	if err := g.Concede(thief.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if !g.Stack.Contains(item.ID) {
		t.Fatalf("the spell left the stack (exile: %v)", g.Exile.Contains(item.ID))
	}
	if it := g.StackMeta[item.ID]; it == nil || it.Controller != me.ID {
		t.Errorf("spell controller after the thief left = %v, want the caster %s", it, me.ID)
	}
	if got := stackCardControllerForTest(t, g, item.ID); got != me.ID {
		t.Errorf("stack card controller = %s, want %s", got, me.ID)
	}
}

// TestSecondThiefLeavingHandsTheSpellToTheFirst — CR 613.7: two
// records, the later one wins, and when it ends the earlier one is
// what applies, not the caster.
func TestSecondThiefLeavingHandsTheSpellToTheFirst(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	advanceTo(t, g, StepPrecombatMain)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	item := castSpellForControlTest(t, g, me, "Instant")
	stealSpell(t, g, item.ID, first.ID)
	stealSpell(t, g, item.ID, second.ID)
	if item.Controller != second.ID {
		t.Fatalf("after two steals the controller is %s, want the later thief %s", item.Controller, second.ID)
	}
	if err := g.Concede(second.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if it := g.StackMeta[item.ID]; it == nil || it.Controller != first.ID {
		t.Errorf("controller after the second thief left = %v, want the first thief %s", it, first.ID)
	}
}

// TestSpellFallingBackToADepartedCasterIsExiled — CR 800.4c for the
// stack: a spell cast off a third player's card, stolen, whose caster
// has since left, has nobody to go back to when the thief leaves.
func TestSpellFallingBackToADepartedCasterIsExiled(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	advanceTo(t, g, StepPrecombatMain)
	caster, thief, cardOwner := g.Seats[0], g.Seats[1], g.Seats[2]
	c := NewCard("Borrowed Instant", cardOwner.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{0}"
	g.WithWriteLock(func() { caster.Hand.PushTop(c) })
	if err := g.CastSpell(caster.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	stealSpell(t, g, c.InstanceID, thief.ID)
	if err := g.Concede(caster.ID); err != nil {
		t.Fatalf("Concede caster: %v", err)
	}
	if !g.Stack.Contains(c.InstanceID) {
		t.Fatal("setup: the stolen spell left with its caster, who does not own it")
	}
	if err := g.Concede(thief.ID); err != nil {
		t.Fatalf("Concede thief: %v", err)
	}
	if g.Stack.Contains(c.InstanceID) {
		t.Error("the spell stayed on the stack under a departed player")
	}
}

// TestStolenSpellCounteredLeavesNoRecord — the stack pin's duration
// ends once the object is no longer on the stack.
func TestStolenSpellCounteredLeavesNoRecord(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	item := castSpellForControlTest(t, g, me, "Instant")
	stealSpell(t, g, item.ID, opp.ID)
	if err := g.CounterSpell(item.ID, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	g.WithWriteLock(func() {
		g.layerVersion.Add(1)
		g.RecomputeLayersIfStaleLocked()
	})
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("registry holds %d records after the stolen spell was countered", n)
	}
	if !me.Graveyard.Contains(item.ID) {
		t.Error("CR 701.6a: a countered stolen spell goes to its OWNER's graveyard")
	}
}

// TestExchangeControlOfSpellAndPermanent — CR 701.12 across the two
// kinds of object: both halves, one timestamp.
func TestExchangeControlOfSpellAndPermanent(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	item := castSpellForControlTest(t, g, me, "Instant")
	chimera := pushScopedTestCreature(g, opp.ID, 2, 3)
	var ok bool
	g.WithWriteLock(func() {
		ok = g.ExchangeControlOfSpellAndPermanentForEffect(uuid.New(), item.ID, chimera, "test — exchange")
	})
	if !ok {
		t.Fatal("the exchange was refused")
	}
	if item.Controller != opp.ID {
		t.Errorf("spell controller = %s, want %s", item.Controller, opp.ID)
	}
	if got := controllerOfCard(t, g, chimera); got != me.ID {
		t.Errorf("creature controller = %s, want %s", got, me.ID)
	}
	if len(g.ScopedEffects) != 2 || g.ScopedEffects[0].Timestamp != g.ScopedEffects[1].Timestamp {
		t.Errorf("want two records sharing one timestamp (CR 613.7), got %+v", g.ScopedEffects)
	}
}

// TestExchangeIsAllOrNothing — CR 701.12a (either object gone) and
// CR 701.12b (one player controls both).
func TestExchangeIsAllOrNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	item := castSpellForControlTest(t, g, me, "Instant")
	mine := pushScopedTestCreature(g, me.ID, 2, 2)
	theirs := pushScopedTestCreature(g, opp.ID, 2, 2)
	g.WithWriteLock(func() {
		if g.ExchangeControlOfSpellAndPermanentForEffect(uuid.New(), item.ID, mine, "same controller") {
			t.Error("CR 701.12b: exchanged a spell and a permanent one player controls")
		}
		if g.ExchangeControlOfSpellAndPermanentForEffect(uuid.New(), item.ID, uuid.New(), "gone") {
			t.Error("CR 701.12a: exchanged with a permanent that is not there")
		}
		if g.ExchangeControlOfSpellAndPermanentForEffect(uuid.New(), uuid.New(), theirs, "no spell") {
			t.Error("CR 701.12a: exchanged with a spell that is not there")
		}
	})
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("refused exchanges registered %d records", n)
	}
}

// TestStolenFaceDownSpellIsSeenByItsNewControllerOnly — CR 708.5,
// literally (owner decision 7): the caster loses the look.
func TestStolenFaceDownSpellIsSeenByItsNewControllerOnly(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	item := castSpellForControlTest(t, g, me, "Creature — Bear")
	g.WithWriteLock(func() {
		g.applyFaceDownLandingLocked(g.Stack, item.ID, FaceDownMorphed, nil)
	})
	stealSpell(t, g, item.ID, opp.ID)
	g.ReadSnapshot(func() {
		c, _ := g.stackCardLocked(item.ID)
		if !c.KnownBy[opp.ID] {
			t.Error("the new controller may not look at the face-down spell they control")
		}
		if c.KnownBy[me.ID] {
			t.Error("CR 708.5: the caster can still look at a face-down spell they no longer control")
		}
	})
}

// TestStolenAdventureIsTheThiefsToCastLater — CR 715.3d names the
// spell's controller (owner decision 4).
func TestStolenAdventureIsTheThiefsToCastLater(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	id := handWithAdventure(t, g, me)
	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure half: %v", err)
	}
	stealSpell(t, g, id, opp.ID)
	resolveTop(t, g)
	if !g.Exile.Contains(id) {
		t.Fatal("the resolved Adventure is not in exile")
	}
	perm := adventureGrantFor(g, id)
	if !perm.Granted() {
		t.Fatal("no grant on the exiled Adventure")
	}
	if perm.Player != opp.ID {
		t.Errorf("grant holder = %s, want the spell's controller %s (CR 715.3d)", perm.Player, opp.ID)
	}
}

// TestStolenSpellSurvivesARestorePoint — ADR 0104 §9: all data, so a
// table with a stolen spell on the stack is a restore point, and the
// restored spell resolves the same way.
func TestStolenSpellSurvivesARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	item := castSpellForControlTest(t, g, me, "Creature — Bear")
	stealSpell(t, g, item.ID, opp.ID)

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a stolen spell on the stack blocks the restore point: %+v", snap.Continuations)
	}
	restored, err := throughJSON(t, snap).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	it := restored.StackMeta[item.ID]
	if it == nil || it.Controller != opp.ID || it.BaseController != me.ID {
		t.Fatalf("restored item = %+v, want controller %s base %s", it, opp.ID, me.ID)
	}
	resolveTop(t, restored)
	perm := spellControlPermanentForTest(t, restored, item.ID)
	if perm.Controller != opp.ID || perm.BaseController != me.ID {
		t.Errorf("restored resolution: controller %s base %s, want %s / %s",
			perm.Controller, perm.BaseController, opp.ID, me.ID)
	}
}
