package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// planechase_die_test.go — ADR 0216 slice 2: the planar die (CR 901.9),
// the planeswalking ability (CR 901.8), planeswalking (CR 901.11), plane
// abilities from the command zone (CR 901.7) and the planar controller
// leaving (CR 901.6).

// planechaseMain is a two-seat Planechase game in the active player's
// precombat main phase, the active player holding priority.
func planechaseMain(t *testing.T) (*Game, *Player) {
	t.Helper()
	installPlanes(t, testPlanes(20))
	g := newPlanechaseGame(t, 2, 21)
	advanceIntoStep(t, g, StepPrecombatMain)
	return g, g.Seats[g.Turn.ActiveSeat]
}

func rollPlanarDie(g *Game, p *Player) error {
	return g.PerformSpecialAction(p.ID, uuid.Nil, SpecialActionPlanarDie, SpecialActionParams{})
}

// forcePlanarFace finishes a planar roll for `player` showing `n`
// (1 chaos, 6 planeswalk, else blank), the drawn number the dice door
// would have handed on.
func forcePlanarFace(t *testing.T, g *Game, player uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		g.beginEventBatchLocked()
		if err := g.finishDiceRollLocked(DiceRoll{Player: player, Sides: 6, N: 1, Planar: true}, []int{n}); err != nil {
			t.Fatalf("finishDiceRollLocked: %v", err)
		}
	})
}

func lastRoll(g *Game) (Event, bool) {
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == EventRollDie {
			return g.Events[i], true
		}
	}
	return Event{}, false
}

// stackItemsWithBody counts the waiting and stacked items with `key`.
func stackItemsWithBody(g *Game, key string) int {
	n := 0
	for _, it := range g.PendingTriggers {
		if it != nil && it.Body == key {
			n++
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Body == key {
			n++
		}
	}
	return n
}

// CR 901.9: {0}, then {1}, then {2} in one turn, and {0} again on the
// next turn. The roll has no number: Amount 0 and a face (CR 901.9d).
func TestThePlanarDieCostsOneMoreEachTimeThisTurn(t *testing.T) {
	g, p := planechaseMain(t)
	for want := range 3 {
		g.WithWriteLock(func() {
			if got := g.PlanarDieCostLocked(); got != "{"+string(rune('0'+want))+"}" {
				t.Fatalf("roll %d costs %s", want+1, got)
			}
		})
		// Drop any planeswalk trigger so the stack is empty again.
		g.WithWriteLock(func() { g.PendingTriggers = nil })
		for len(g.StackMeta) > 0 {
			resolveTop(t, g)
		}
		if err := rollPlanarDie(g, p); err != nil {
			t.Fatalf("roll %d: %v", want+1, err)
		}
		ev, ok := lastRoll(g)
		if !ok || ev.PlanarFace == "" || ev.Amount != 0 || ev.Actor != p.ID {
			t.Fatalf("roll %d: event %+v", want+1, ev)
		}
	}
	if g.Planechase.DieRolls != 3 {
		t.Errorf("DieRolls = %d, want 3", g.Planechase.DieRolls)
	}
	advanceOneTurn(t, g)
	g.WithWriteLock(func() {
		if got := g.PlanarDieCostLocked(); got != "{0}" {
			t.Errorf("next turn costs %s, want {0}", got)
		}
	})
}

// CR 901.9's window: the active player, a main phase, an empty stack,
// on a Planechase table, with no card named.
func TestThePlanarDieWindow(t *testing.T) {
	g, p := planechaseMain(t)
	other := g.Seats[1-g.Turn.ActiveSeat]
	if err := rollPlanarDie(g, other); !errors.Is(err, ErrSpecialActionTiming) {
		t.Errorf("the non-active player: %v, want ErrSpecialActionTiming", err)
	}
	if err := g.PerformSpecialAction(p.ID, uuid.New(), SpecialActionPlanarDie, SpecialActionParams{}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("with a card: %v, want ErrInvalidParam", err)
	}
	advanceIntoStep(t, g, StepBeginCombat)
	if err := rollPlanarDie(g, p); !errors.Is(err, ErrSpecialActionTiming) {
		t.Errorf("outside a main phase: %v, want ErrSpecialActionTiming", err)
	}
	if g.Planechase.DieRolls != 0 {
		t.Error("a refused roll was counted")
	}

	plain := newRestorableGame(t)
	advanceIntoStep(t, plain, StepPrecombatMain)
	if err := rollPlanarDie(plain, plain.Seats[plain.Turn.ActiveSeat]); !errors.Is(err, ErrSpecialActionNotOffered) {
		t.Errorf("a table without Planechase: %v, want ErrSpecialActionNotOffered", err)
	}
}

// A roll an effect causes costs nothing and does not count (CR 901.9's
// note), and the planar die has its own stream: it does not move the
// roll stream a card's d20 draws from.
func TestARollForAnEffectIsFreeAndUncounted(t *testing.T) {
	g, p := planechaseMain(t)
	g.WithWriteLock(func() {
		if err := g.RollPlanarDieForEffect(p.ID, uuid.Nil); err != nil {
			t.Fatalf("RollPlanarDieForEffect: %v", err)
		}
		if g.planarDieTakenThisTurnLocked() != 0 {
			t.Error("an effect's roll counted toward the cost")
		}
		_, label := g.rngPlayerIdentityLocked(p.ID)
		if n := g.rngCounters["roll/"+label+"/"+uuid.Nil.String()]; n != 0 {
			t.Errorf("the planar roll drew %d from the roll stream", n)
		}
	})
	if ev, ok := lastRoll(g); !ok || ev.PlanarFace == "" {
		t.Errorf("no planar roll event: %+v", ev)
	}
	plain := newRestorableGame(t)
	plain.WithWriteLock(func() {
		if err := plain.RollPlanarDieForEffect(plain.Seats[0].ID, uuid.Nil); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("without Planechase: %v", err)
		}
	})
}

// Undo rewinds the face and the count (ADR 0054 Decision 4): rolling
// again after the undo gives the same face.
func TestUndoRewindsThePlanarDie(t *testing.T) {
	g, p := planechaseMain(t)
	before := g.Clone()
	if err := rollPlanarDie(g, p); err != nil {
		t.Fatal(err)
	}
	first, _ := lastRoll(g)
	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if g.Planechase.DieRolls != 0 {
		t.Fatalf("the undo kept the count: %d", g.Planechase.DieRolls)
	}
	if err := rollPlanarDie(g, p); err != nil {
		t.Fatal(err)
	}
	if again, _ := lastRoll(g); again.PlanarFace != first.PlanarFace {
		t.Errorf("after the undo the die showed %s, before it %s", again.PlanarFace, first.PlanarFace)
	}
}

// CR 901.8: the Planeswalker symbol triggers one sourceless ability,
// controlled by the roller; it uses the stack, a table waiting on it is
// a restore point, and it planeswalks when it resolves (CR 901.11).
func TestThePlaneswalkerSymbolPlaneswalks(t *testing.T) {
	g, p := planechaseMain(t)
	from := g.PlanarFaceUpCards()[0]
	top := g.PlanarDeckCards()[len(g.PlanarDeckCards())-1]
	forcePlanarFace(t, g, p.ID, 6)
	if n := stackItemsWithBody(g, planeswalkBodyKey); n != 1 {
		t.Fatalf("%d planeswalking abilities, want 1", n)
	}
	for _, it := range append(append([]*StackItem(nil), g.PendingTriggers...), stackValues(g)...) {
		if it.Body == planeswalkBodyKey && (it.Controller != p.ID || it.SourceCardID != uuid.Nil) {
			t.Errorf("the ability is controlled by %v with source %v", it.Controller, it.SourceCardID)
		}
	}

	// A restore point while it waits.
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	r, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("a table waiting on the planeswalking ability is not a restore point: %v", err)
	}
	if stackItemsWithBody(r, planeswalkBodyKey) != 1 {
		t.Fatal("the restore dropped the planeswalking ability")
	}

	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	resolveTop(t, g)
	up := g.PlanarFaceUpCards()
	if len(up) != 1 || up[0].InstanceID != top.InstanceID {
		t.Fatalf("face up %v, want the deck's top card %s", attrNames(up), top.Name)
	}
	deck := g.PlanarDeckCards()
	if deck[0].InstanceID != from.InstanceID || deck[0].IsKnownTo(p.ID) {
		t.Errorf("the old plane is not face down on the bottom: %s", deck[0].Name)
	}
	var to, away bool
	for _, ev := range g.Events {
		to = to || (ev.Kind == EventPlaneswalkTo && ev.CardID == top.InstanceID && ev.Actor == p.ID)
		away = away || (ev.Kind == EventPlaneswalkFrom && ev.CardID == from.InstanceID && ev.Actor == p.ID)
	}
	if !to || !away {
		t.Errorf("planeswalk events: to %v, away %v", to, away)
	}
}

func stackValues(g *Game) []*StackItem {
	out := make([]*StackItem, 0, len(g.StackMeta))
	for _, it := range g.StackMeta {
		out = append(out, it)
	}
	return out
}

// CR 901.11c: with two face-up planes the player planeswalks away from
// both; the planar controller orders them onto the bottom.
func TestPlaneswalkingAwayFromTwoPlanes(t *testing.T) {
	g, p := planechaseMain(t)
	g.WithWriteLock(func() {
		// A second face-up plane, as "planeswalk to two planes" would
		// leave the table.
		c, _ := g.PlanarDeck.PopTop()
		g.PlanarFaceUp.PushTop(c)
	})
	a, b := g.PlanarFaceUpCards()[0], g.PlanarFaceUpCards()[1]
	g.WithWriteLock(func() {
		if err := g.PlaneswalkForEffect(p.ID); err != nil {
			t.Fatal(err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d pending choices, want the bottom-order pick", len(g.PendingChoices))
	}
	pc := g.PendingChoices[0]
	if pc.Chooser != g.Planechase.Controller {
		t.Errorf("the order is asked of %v, want the planar controller", pc.Chooser)
	}
	// Second card first.
	if err := g.ResolveOptionPick(pc.ID, pc.Chooser, 1); err != nil {
		t.Fatal(err)
	}
	deck := g.PlanarDeckCards()
	if deck[1].InstanceID != b.InstanceID || deck[0].InstanceID != a.InstanceID {
		t.Errorf("bottom order %s, %s; want %s under %s", deck[0].Name, deck[1].Name, a.Name, b.Name)
	}
	if up := g.PlanarFaceUpCards(); len(up) != 1 || up[0].InstanceID == a.InstanceID || up[0].InstanceID == b.InstanceID {
		t.Errorf("face up after planeswalking: %v", attrNames(up))
	}
}

// CR 901.11: an "until you planeswalk away" effect ends as its plane
// leaves the face-up set, and not before.
func TestUntilYouPlaneswalkAwayEnds(t *testing.T) {
	g, p := planechaseMain(t)
	plane := g.PlanarFaceUpCards()[0].InstanceID
	g.WithWriteLock(func() {
		d, ok := g.UntilPlaneswalkAwayDuration(plane)
		if !ok {
			t.Fatal("no duration for the face-up plane")
		}
		if d.Problem() != "" {
			t.Fatalf("duration problem: %s", d.Problem())
		}
		if !g.RegisterScopedRuleEffectForEffect(plane, ScopeGame, p.ID,
			[]Mod{{Kind: ModRollExtraDieIgnoreOne, Player: p.ID}}, d, "test plane") {
			t.Fatal("no effect registered")
		}
		if _, ok := g.UntilPlaneswalkAwayDuration(uuid.New()); ok {
			t.Error("a card that is not face up gave a duration")
		}
	})
	live := func() int {
		n := 0
		for _, e := range g.ScopedEffects {
			if e.Duration.Kind == UntilPlaneswalkAway && !e.Duration.Ended {
				n++
			}
		}
		return n
	}
	advanceOneTurn(t, g)
	if live() != 1 {
		t.Fatal("the effect ended on a turn boundary")
	}
	g.WithWriteLock(func() {
		if err := g.PlaneswalkForEffect(p.ID); err != nil {
			t.Fatal(err)
		}
	})
	if live() != 0 {
		t.Error("the effect outlived planeswalking away from its plane")
	}
	if !UntilPlaneswalkAway.Known() || UntilPlaneswalkAway+1 != UntilNextEndStep {
		t.Error("UntilNextEndStep does not follow UntilPlaneswalkAway")
	}
}

// CR 901.7: a face-up plane's triggered ability triggers from the command
// zone, controlled by the planar controller; chaos is one such ability
// (CR 901.9b).
func TestAFaceUpPlanesChaosAbilityTriggers(t *testing.T) {
	g, p := planechaseMain(t)
	plane := g.PlanarFaceUpCards()[0]
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != plane.OracleID {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventRollDie}, Key: "chaos",
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.PlanarFace == PlanarFaceChaos && ev.Actor == source.Controller
			},
			Effect: func(*Game, *StackItem) error { return nil },
		}}
	})
	forcePlanarFace(t, g, p.ID, 2)
	if countSourced(g, plane.InstanceID) != 0 {
		t.Fatal("a blank triggered the chaos ability")
	}
	forcePlanarFace(t, g, p.ID, 1)
	if countSourced(g, plane.InstanceID) != 1 {
		t.Fatal("chaos did not trigger the face-up plane's ability")
	}
	for _, it := range g.PendingTriggers {
		if it.SourceCardID == plane.InstanceID && it.Controller != p.ID {
			t.Errorf("the chaos ability is controlled by %v, want the planar controller", it.Controller)
		}
	}
	// A plane in the deck triggers nothing.
	deckPlane := g.PlanarDeckCards()[0]
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != deckPlane.OracleID {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventRollDie}, Key: "chaos",
			AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true },
			Effect:    func(*Game, *StackItem) error { return nil },
		}}
	})
	forcePlanarFace(t, g, p.ID, 1)
	if countSourced(g, deckPlane.InstanceID) != 0 {
		t.Error("a face-down plane's ability triggered")
	}
}

func countSourced(g *Game, source uuid.UUID) int {
	n := 0
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == source {
			n++
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == source {
			n++
		}
	}
	return n
}

// Squid Fire Knight's sticker applies to the planar die: it is still a
// die. The ignore prompt shows faces, and the kept roll is one face.
func TestTheExtraDieShieldAppliesToThePlanarDie(t *testing.T) {
	g, p := planechaseMain(t)
	other := g.Seats[1-g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		if !g.RollExtraDieNextTimeForEffect(uuid.Nil, other.ID, p.ID, "Squid Fire Knight") {
			t.Fatal("no shield")
		}
	})
	if err := rollPlanarDie(g, p); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d pending choices, want the ignore prompt", len(g.PendingChoices))
	}
	pc := g.PendingChoices[0]
	for _, o := range pc.PickOptions {
		if o.Label != "Ignore the blank (die 1)" && o.Label != "Ignore the blank (die 2)" &&
			o.Label != "Ignore the chaos (die 1)" && o.Label != "Ignore the chaos (die 2)" &&
			o.Label != "Ignore the planeswalk (die 1)" && o.Label != "Ignore the planeswalk (die 2)" {
			t.Errorf("an ignore option shows a number: %q", o.Label)
		}
	}
	rollsBefore := 0
	for _, ev := range g.Events {
		if ev.Kind == EventRollDie {
			rollsBefore++
		}
	}
	if err := g.ResolveOptionPick(pc.ID, pc.Chooser, 0); err != nil {
		t.Fatal(err)
	}
	rolls := 0
	for _, ev := range g.Events {
		if ev.Kind == EventRollDie {
			rolls++
			if ev.PlanarFace == "" {
				t.Error("a kept planar roll has no face")
			}
		}
	}
	if rolls-rollsBefore != 1 {
		t.Errorf("%d kept rolls, want 1", rolls-rollsBefore)
	}
	if g.Planechase.DieRolls != 1 {
		t.Errorf("the action was counted %d times", g.Planechase.DieRolls)
	}
}

// CR 901.6: the planar controller leaving hands the designation to the
// next player in turn order first, so none of the communal deck leaves
// with them (CR 901.15b, 800.4a).
func TestThePlanarControllerLeavingSweepsNoPlanarCard(t *testing.T) {
	installPlanes(t, testPlanes(30))
	g := newPlanechaseGame(t, 3, 4)
	advanceIntoStep(t, g, StepPrecombatMain)
	active := g.Seats[g.Turn.ActiveSeat]
	next := g.Seats[(g.Turn.ActiveSeat+1)%3]
	before := len(g.PlanarDeckCards()) + len(g.PlanarFaceUpCards())
	if err := g.Concede(active.ID); err != nil {
		t.Fatal(err)
	}
	if after := len(g.PlanarDeckCards()) + len(g.PlanarFaceUpCards()); after != before {
		t.Fatalf("%d planar cards after the departure, want %d", after, before)
	}
	if len(g.PlanarFaceUpCards()) != 1 {
		t.Errorf("%d face-up planes after the departure", len(g.PlanarFaceUpCards()))
	}
	if g.Planechase.Controller == active.ID {
		t.Fatal("the departed player is still the planar controller")
	}
	// The next player became planar controller as the old one left;
	// the turn that followed may since have moved it on to its own
	// active player, which is still never the departed seat.
	if g.Planechase.Controller != next.ID && g.Planechase.Controller != g.Seats[g.Turn.ActiveSeat].ID {
		t.Errorf("planar controller %v, want the next player or the new active player", g.Planechase.Controller)
	}
	for _, c := range append(g.PlanarDeckCards(), g.PlanarFaceUpCards()...) {
		if c.Owner != g.Planechase.Controller {
			t.Errorf("%s is owned by %v, not the planar controller", c.Name, c.Owner)
		}
	}
}
