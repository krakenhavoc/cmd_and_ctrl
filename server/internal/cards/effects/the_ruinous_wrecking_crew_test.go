package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ruinous_wrecking_crew_test.go — #4617: "choose up to X —" where
// X is the X paid for the spell (CR 107.3m), through the engine.

// Bullet indexes, in printed order.
const (
	crewRummage = iota
	crewDrain
	crewToken
	crewEdict
)

// castCrew casts The Ruinous Wrecking Crew for `x` from the active
// seat's hand and passes until its enters trigger asks (or doesn't).
func castCrew(t *testing.T, g *game.Game, x int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "The Ruinous Wrecking Crew", TypeLine: "Legendary Creature — Human Villain",
		OracleID: theRuinousWreckingCrewOracle, ManaCost: "{X}{B}{R}", Power: 2, Toughness: 2,
		Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("cast for X=%d: %v", x, err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("X=%d: the Crew did not enter", x)
	}
	return id
}

// crewTriggerWaiting reports whether the Crew's trigger is anywhere
// between its harvest and its resolution.
func crewTriggerWaiting(g *game.Game, crew uuid.UUID) bool {
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == crew {
			return true
		}
	}
	return triggerOnStack(g, crew) != nil
}

func TestTheRuinousWreckingCrewIsFull(t *testing.T) {
	spec, ok := Lookup(theRuinousWreckingCrewOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Triggered) != 1 {
		t.Fatalf("one enters trigger, Full: %+v", spec)
	}
	ms := spec.Triggered[0].Modes
	if ms == nil || !ms.UpToCastX || ms.Min != 0 || ms.Max != 4 || len(ms.Options) != 4 {
		t.Fatalf("choose up to X of four bullets: %+v", ms)
	}
}

// X = 0: the counters are none, and "choose up to 0" chooses no mode,
// so the trigger is removed without asking (CR 700.2b).
func TestTheRuinousWreckingCrewAtXZeroChoosesNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crew := castCrew(t, g, 0)
	noModePick(t, g, me.ID, "X = 0 asks nothing")
	if crewTriggerWaiting(g, crew) {
		t.Error("X = 0: the trigger is removed, not put on the stack")
	}
	if n := countersOn(g, crew, game.CounterPlusOne); n != 0 {
		t.Errorf("X = 0: %d counters, want none", n)
	}
}

// X = 2: two counters, and a prompt for up to two bullets. Three are
// refused; two resolve in printed order.
func TestTheRuinousWreckingCrewAtXTwoChoosesUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := b12Creature(g, me.ID, "My Fodder", "Creature — Human", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life

	crew := castCrew(t, g, 2)
	if n := countersOn(g, crew, game.CounterPlusOne); n != 2 {
		t.Errorf("X = 2: %d counters, want 2", n)
	}
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("X = 2: no mode prompt")
	}
	if c.ModeMin != 0 || c.ModeMax != 2 {
		t.Errorf("prompt bounds %d..%d, want 0..2", c.ModeMin, c.ModeMax)
	}
	if !strings.Contains(c.Reason, "X is 2") {
		t.Errorf("the prompt names its X: %q", c.Reason)
	}
	// No token on the battlefield, so "destroy target token" can't be
	// chosen (CR 700.2a) and is not offered.
	if got := c.ModeOptionIndex; len(got) != 3 || got[0] != crewRummage || got[1] != crewDrain || got[2] != crewEdict {
		t.Errorf("offered %v, want the three bullets with something to do", got)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{crewRummage, crewDrain, crewEdict}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("three bullets at X = 2: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{crewEdict, crewDrain}); err != nil {
		t.Fatalf("two bullets at X = 2: %v", err)
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil || !hasID(pick.PickTargetPlayers, opp.ID) || hasID(pick.PickTargetPlayers, me.ID) {
		t.Fatalf("the drain targets an opponent: %+v", pick)
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, mine)
	answerSacrifice(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("target opponent loses 2 life: %d → %d", life, opp.Life)
	}
	if zoneOf(g, mine) == game.ZoneBattlefield || zoneOf(g, theirs) == game.ZoneBattlefield {
		t.Error("each player sacrifices a creature of their choice")
	}
	if !g.Battlefield.Contains(crew) {
		t.Error("the Crew's controller chose the fodder, so the Crew stays")
	}
}

// X = 6: the count caps at the four printed bullets, and all four run
// in printed order whatever order they were chosen in (CR 608.2c).
func TestTheRuinousWreckingCrewCountCapsAtFourBullets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	mine := b12Creature(g, me.ID, "My Fodder", "Creature — Human", 1, 1)
	me.Hand.PushTop(game.NewCard("Pitch", me.ID))
	seedLibrary(me, "A", "B")
	life := opp.Life

	crew := castCrew(t, g, 6)
	if n := countersOn(g, crew, game.CounterPlusOne); n != 6 {
		t.Errorf("X = 6: %d counters, want 6", n)
	}
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("X = 6: no mode prompt")
	}
	if c.ModeMax != 4 || len(c.ModeOptionIndex) != 4 {
		t.Fatalf("X = 6 caps at the four bullets: max %d, offered %v", c.ModeMax, c.ModeOptionIndex)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{crewEdict, crewEdict}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a bullet twice: %v, want ErrInvalidParam (CR 700.2d)", err)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{crewEdict, crewToken, crewDrain, crewRummage}); err != nil {
		t.Fatalf("all four: %v", err)
	}
	// One pick per targeted bullet, in the order they were chosen.
	pickCard(t, g, me.ID, token)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	hand := handCount(g, me)
	passPriorityAroundTable(t, g)
	// The rummage is printed first, so it asks first, and nothing after
	// it has happened yet.
	if opp.Life != life || zoneOf(g, token) != game.ZoneBattlefield {
		t.Fatal("the later bullets wait for the rummage's discard")
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if got := handCount(g, me); got != hand {
		t.Errorf("discard a card, then draw a card: hand %d → %d", hand, got)
	}
	if opp.Life != life-2 {
		t.Errorf("target opponent loses 2 life: %d → %d", life, opp.Life)
	}
	if zoneOf(g, token) == game.ZoneBattlefield {
		t.Error("destroy target token")
	}
	if zoneOf(g, mine) == game.ZoneBattlefield {
		t.Error("each player sacrifices a creature")
	}
}

// "Up to": no bullet at all is a legal answer, and the trigger is then
// removed (CR 700.2b) rather than asked again.
func TestTheRuinousWreckingCrewChoosingNoneDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	life := opp.Life
	crew := castCrew(t, g, 3)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{}); err != nil {
		t.Fatalf("choosing none: %v", err)
	}
	noModePick(t, g, me.ID, "an empty answer is final")
	if crewTriggerWaiting(g, crew) {
		t.Error("no mode chosen: the trigger never reaches the stack")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Error("nothing chosen, nothing happens")
	}
}

// A Crew that was not cast has no spell X (CR 107.3m), so it can
// choose no bullet and asks nothing.
func TestTheRuinousWreckingCrewNotCastChoosesNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "The Ruinous Wrecking Crew", TypeLine: "Legendary Creature — Human Villain",
		OracleID: theRuinousWreckingCrewOracle, ManaCost: "{X}{B}{R}", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	var entered uuid.UUID
	g.WithWriteLock(func() {
		var err error
		if entered, err = g.ReturnToBattlefieldForEffect(id, me.ID, false); err != nil {
			t.Errorf("return: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if entered == uuid.Nil || !g.Battlefield.Contains(entered) {
		t.Fatal("the Crew did not return")
	}
	noModePick(t, g, me.ID, "a permanent that was not cast has X = 0")
	if crewTriggerWaiting(g, entered) {
		t.Error("the trigger is removed")
	}
}

// The prompt's bound is recorded with it, so a restore point and an
// undo still refuse more bullets than X.
func TestTheRuinousWreckingCrewBoundSurvivesRestoreAndUndo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCrew(t, g, 1)

	restored := restoreRoundTrip(t, g, false)
	p := pendingOfKind(restored, game.PendingChoiceModePick)
	if p == nil || p.ModeMin != 0 || p.ModeMax != 1 {
		t.Fatalf("restored prompt bounds: %+v, want 0..1", p)
	}
	if err := restored.ResolveModePick(p.ID, me.ID, []int{crewRummage, crewEdict}); err == nil {
		t.Error("restored: two bullets at X = 1 must still be refused")
	}

	undo := g.Clone()
	p = pendingOfKind(g, game.PendingChoiceModePick)
	if err := g.ResolveModePick(p.ID, me.ID, []int{crewEdict}); err != nil {
		t.Fatalf("one bullet: %v", err)
	}
	g.RestoreFrom(undo)
	p = pendingOfKind(g, game.PendingChoiceModePick)
	if p == nil || p.ModeMax != 1 {
		t.Fatalf("after undo the prompt is back at X = 1: %+v", p)
	}
}

// Register keeps the count where CR 107.3m puts the spell's X: on an
// enters trigger, as a plain up-to count.
func TestChooseUpToXIsRefusedOffAnEntersTrigger(t *testing.T) {
	bullets := func() []game.ModeOption {
		return []game.ModeOption{Mode("Draw a card."), Mode("Gain 2 life.")}
	}
	noop := func(*game.Game, *game.StackItem) error { return nil }
	if msg := registerPanics(Spec{OracleID: "test-4617-spell", Name: "Up-to-X Spell", Modes: ChooseUpToX(bullets()...)}); !strings.Contains(msg, "choose up to X") {
		t.Errorf("a spell's modes: %q", msg)
	}
	expectRegisterPanic(t, "choose up to X", func() {
		checkUpToCastX("Fixture", "an activated ability's Modes", ChooseUpToX(bullets()...), nil)
	})
	attacks := WheneverThisAttacks("Fixture — attacks", noop)
	expectRegisterPanic(t, "not an enters trigger", func() {
		checkUpToCastX("Fixture", "a trigger's Modes", ChooseUpToX(bullets()...), &attacks)
	})
	both := WhenThisEntersOrAttacks("Fixture — enters or attacks", noop)
	expectRegisterPanic(t, "not an enters trigger", func() {
		checkUpToCastX("Fixture", "a trigger's Modes", ChooseUpToX(bullets()...), &both)
	})
	enters := WhenThisEnters("Fixture — enters", noop)
	repeat := ChooseUpToX(bullets()...)
	repeat.Repeatable = true
	expectRegisterPanic(t, "not a plain up-to count", func() {
		checkUpToCastX("Fixture", "a trigger's Modes", repeat, &enters)
	})
	floor := ChooseUpToX(bullets()...)
	floor.Min = 1
	expectRegisterPanic(t, "not a plain up-to count", func() {
		checkUpToCastX("Fixture", "a trigger's Modes", floor, &enters)
	})
	checkUpToCastX("Fixture", "a trigger's Modes", ChooseUpToX(bullets()...), &enters)
	granted := AbilityGrant{Key: "fixture/4617-up-to-x", Text: "x", Triggered: []game.TriggeredAbility{enters}}
	granted.Triggered[0].Modes = ChooseUpToX(bullets()...)
	expectRegisterPanic(t, "107.3j", func() { checkGrantAbilities("Fixture", granted) })
}
