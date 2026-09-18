package game

import (
	"testing"

	"github.com/google/uuid"
)

// rotation_test.go pins the single turn-rotation seam (ADR 0059
// Decision 6, #766): when the active player leaves the game, the rest
// of their turn ends at once — combat is cleared, the cleanup sweep
// runs, and the next turn begins through the same code the cursor
// takes past cleanup, so the next player starts clean.

// rotationProbe is #766's four-seat probe, set up to the moment the
// departure happens: seat 0 (A) attacks seat 1 (N) with a 2/2 and a
// 3/3, N blocks the 2/2 with a 4/4, and combat damage has been dealt.
// A has also cast a spell and triggered a once-per-turn ability, and
// a Fog-style shield and an "until end of turn" static are live.
type rotationProbe struct {
	g       *Game
	a, n    *Player
	blocker uuid.UUID
	tallied string
}

func newRotationProbe(t *testing.T) rotationProbe {
	t.Helper()
	g := newFourPlayerActiveGame(t)
	a, n := g.Seats[0], g.Seats[1]
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: active seat %d, want 0", g.Turn.ActiveSeat)
	}
	small := pushKeywordCreature(t, g, a, 2, 2)
	big := pushKeywordCreature(t, g, a, 3, 3)
	blocker := pushKeywordCreature(t, g, n, 4, 4)

	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, id := range []uuid.UUID{small, big} {
		if err := g.DeclareAttacker(id, n.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, small); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceIntoStep(t, g, StepCombatDamage)

	tallied := TallyKey(uuid.New(), "once each turn")
	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]CastTally{n.ID: {Total: 1, Noncreature: 1}}
		g.LoyaltyActivatedThisTurn = map[uuid.UUID]bool{uuid.New(): true}
		g.LandsPlayedThisTurn = map[uuid.UUID]int{a.ID: 1}
		g.DrawnThisTurn = map[uuid.UUID][]uuid.UUID{a.ID: {uuid.New()}}
		if g.TurnTally.Triggered == nil {
			g.TurnTally.Triggered = map[string]int{}
		}
		g.TurnTally.Triggered[tallied] = 1
		g.RegisterTurnScopedReplacement(ReplacementEffect{
			Watches: []EventKind{EventDealDamage},
			Label:   "Fog: prevent all combat damage this turn",
		})
		g.RegisterTurnScopedStaticForEffect(StaticAbility{
			Layer:     Layer7PT,
			SubLayer:  SubLayer7C_Modify,
			AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return target.InstanceID == blocker },
			Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.Power += 3; c.Toughness += 3 },
		}, uuid.New(), "test — +3/+3 until end of turn")
	})

	// The probe's "before" column: the departure is what must clear
	// all of this.
	if n.Life != 37 {
		t.Fatalf("setup: N at %d life after the unblocked 3/3, want 37", n.Life)
	}
	if c := findCard(g, blocker); c == nil || c.DamageMarked != 2 {
		t.Fatalf("setup: blocker should have 2 damage marked: %+v", c)
	}
	if got := g.TurnTallyFor(n.ID).LifeLost; got != 3 {
		t.Fatalf("setup: N's LifeLost tally %d, want 3", got)
	}
	if c := findCard(g, big); c == nil || c.AttackingTarget != n.ID {
		t.Fatalf("setup: the 3/3 should be attacking N")
	}
	return rotationProbe{g: g, a: a, n: n, blocker: blocker, tallied: tallied}
}

// assertCleanNextTurn is the probe's "rules" column, read at N's
// upkeep: every per-turn cache is empty, nothing is marked or in
// combat, and the turn-scoped registries are swept.
func (p rotationProbe) assertCleanNextTurn(t *testing.T) {
	t.Helper()
	g := p.g
	if g.State != StateActive {
		t.Fatalf("three players remain; the game should go on: %s", g.State)
	}
	if g.Turn.ActiveSeat != 1 || g.Turn.Step != StepUpkeep || g.Turn.PriorityHolder != 1 {
		t.Fatalf("cursor: seat %d step %s priority %d, want N's upkeep with priority", g.Turn.ActiveSeat, g.Turn.Step, g.Turn.PriorityHolder)
	}
	if got := g.CastTallyFor(p.n.ID); got != (CastTally{}) {
		t.Errorf("CastTallyFor(N) = %+v, want zero", got)
	}
	if len(g.LoyaltyActivatedThisTurn) != 0 || len(g.LandsPlayedThisTurn) != 0 || len(g.DrawnThisTurn) != 0 {
		t.Errorf("per-turn maps survived: loyalty %v lands %v drawn %v", g.LoyaltyActivatedThisTurn, g.LandsPlayedThisTurn, g.DrawnThisTurn)
	}
	if got := g.TurnTallyFor(p.n.ID).LifeLost; got != 0 {
		t.Errorf("TurnTallyFor(N).LifeLost = %d, want 0", got)
	}
	if g.TurnTally.Triggered[p.tallied] != 0 {
		t.Errorf("once-per-turn gate from A's turn still shut: %v", g.TurnTally.Triggered)
	}
	for _, c := range g.Battlefield.Cards {
		if c.DamageMarked != 0 || c.MarkedLethalByDeathtouch {
			t.Errorf("%s still has damage marked (%d)", c.Name, c.DamageMarked)
		}
		if c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil {
			t.Errorf("%s is still in combat: attacking %v blocking %v", c.Name, c.AttackingTarget, c.BlockingTarget)
		}
	}
	if len(g.TurnScopedReplacements) != 0 {
		t.Errorf("A's turn-scoped prevention shield survived into N's turn: %d", len(g.TurnScopedReplacements))
	}
	if len(g.TurnScopedStatics) != 0 {
		t.Errorf("A's until-end-of-turn static survived into N's turn: %d", len(g.TurnScopedStatics))
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := findCard(g, p.blocker); c == nil || c.CurrentPower() != 4 {
		t.Errorf("blocker should be a plain 4/4 again: %+v", c)
	}

	// N's own combat, with nothing declared: the departed player's
	// 3/3 is not attacking any more, so N takes nothing.
	advanceIntoStep(t, g, StepCombatDamage)
	if p.n.Life != 37 {
		t.Errorf("N took damage in its own combat: life %d, want 37", p.n.Life)
	}
}

// TestActiveSeatConcedeEndsTheTurnThroughTheSeam is #766's probe as a
// permanent test: A concedes in its combat damage step.
func TestActiveSeatConcedeEndsTheTurnThroughTheSeam(t *testing.T) {
	p := newRotationProbe(t)
	if err := p.g.Concede(p.a.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	p.assertCleanNextTurn(t)
}

// TestActiveSeatSBALossEndsTheTurnThroughTheSeam: an SBA loss (0
// life) in the same combat damage step takes the same path as a
// concede.
func TestActiveSeatSBALossEndsTheTurnThroughTheSeam(t *testing.T) {
	p := newRotationProbe(t)
	if _, err := p.g.ChangePlayerLife(p.a.ID, -p.a.Life); err != nil {
		t.Fatalf("ChangePlayerLife: %v", err)
	}
	if p.a.Eliminated {
		t.Fatal("setup: the life change itself must not run the SBA pass")
	}
	p.g.WithWriteLock(func() { p.g.runStateChecksLocked() })
	if !p.a.Eliminated {
		t.Fatal("A at 0 life should be eliminated by the SBA pass")
	}
	p.assertCleanNextTurn(t)
}

// TestPerTurnCachesClearWhenActiveSeatConcedes is the four-seat
// version of TestPerTurnCachesClearOnPriorityWrap: a real land drop on
// the departed player's turn, and every cache empty at the next seat's
// upkeep.
func TestPerTurnCachesClearWhenActiveSeatConcedes(t *testing.T) {
	g := newWrapGame(t, 4)
	active := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != StepPrecombatMain {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	land := active.Hand.Cards[0]
	if err := g.CastSpell(active.ID, land.InstanceID, CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	if got := g.LandsPlayedThisTurnFor(active.ID); got != 1 {
		t.Fatalf("LandsPlayedThisTurn after a land drop: %d", got)
	}
	next := g.Seats[(g.Turn.ActiveSeat+1)%4]
	g.SpellsCastThisTurn = map[uuid.UUID]CastTally{next.ID: {Total: 1, Noncreature: 1}}
	g.LoyaltyActivatedThisTurn = map[uuid.UUID]bool{uuid.New(): true}
	versionBefore := g.layerVersion.Load()

	if err := g.Concede(active.ID); err != nil {
		t.Fatal(err)
	}

	if g.Turn.ActiveSeat == active.Seat || g.Turn.Step != StepUpkeep {
		t.Fatalf("cursor should be at the next seat's upkeep: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if got := g.LandsPlayedThisTurnFor(active.ID); got != 0 {
		t.Errorf("LandsPlayedThisTurn survived the departure: %d", got)
	}
	if len(g.SpellsCastThisTurn) != 0 {
		t.Errorf("SpellsCastThisTurn survived the departure: %v", g.SpellsCastThisTurn)
	}
	if len(g.LoyaltyActivatedThisTurn) != 0 {
		t.Errorf("LoyaltyActivatedThisTurn survived the departure: %v", g.LoyaltyActivatedThisTurn)
	}
	if g.layerVersion.Load() == versionBefore {
		t.Error("the new turn did not invalidate the layer cache")
	}
	// The departed seat drew on its turn (CR 103.8c: nobody skips at a
	// four-player table); the next seat has not reached its draw step.
	if len(g.DrawnThisTurn) != 0 {
		t.Errorf("DrawnThisTurn survived the departure: %v", g.DrawnThisTurn)
	}
}

// TestPassTurnSweepsTheTurn: the sandbox pass_turn verb ends the turn
// through the same seam, so marked damage, combat and "until end of
// turn" effects end with it (#766, ADR 0059 Decision 6).
func TestPassTurnSweepsTheTurn(t *testing.T) {
	p := newRotationProbe(t)
	if err := p.g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	p.assertCleanNextTurn(t)
}

// TestSimultaneousLossesMoveTheTurnOnOnce: two players lose in one SBA
// pass, one of them the active seat. The turn moves on once, past both,
// so the second loser's turn never begins — no step of theirs ever
// announces.
//
// The probe used to be one of the second loser's own permanents, left
// tapped and expected to stay tapped through a turn that never
// untapped it. CR 800.4a (#769) takes that permanent out of the game
// with its controller, so "still tapped" and "not there at all" became
// the same observation and the probe stopped distinguishing anything.
// EventStepBegan carries the active player, which is the thing the
// test was always really asking about.
func TestSimultaneousLossesMoveTheTurnOnOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, n := g.Seats[0], g.Seats[1]
	nCreature := pushKeywordCreature(t, g, n, 1, 1)
	g.WithWriteLock(func() {
		a.Life = 0
		n.Life = 0
		g.runStateChecksLocked()
	})
	if !a.Eliminated || !n.Eliminated {
		t.Fatalf("both should have lost: a %v n %v", a.Eliminated, n.Eliminated)
	}
	if g.State != StateActive {
		t.Fatalf("two players remain: %s", g.State)
	}
	if g.Turn.ActiveSeat != 2 || g.Turn.Step != StepUpkeep {
		t.Errorf("cursor: seat %d step %s, want seat 2's upkeep", g.Turn.ActiveSeat, g.Turn.Step)
	}
	for _, ev := range g.Events {
		if ev.Kind == EventStepBegan && ev.Actor == n.ID {
			t.Errorf("the second loser's %s step began: their turn ran inside the batch", ev.Step)
		}
	}
	// CR 800.4a while we are here: the losers' permanents went with
	// them, so there is nothing of the second loser's left to untap.
	if c := findCard(g, nCreature); c != nil {
		t.Errorf("the second loser's creature is still on the battlefield: %+v", c)
	}
}

// TestGameOverKeepsTheFinalBoard: when a departure ends the game, the
// turn is not ended and no next turn begins, so the board the game
// ended on keeps its marked damage and its combat.
func TestGameOverKeepsTheFinalBoard(t *testing.T) {
	g := newActiveGame(t)
	a, n := g.Seats[0], g.Seats[1]
	atk := pushKeywordCreature(t, g, a, 3, 3)
	blocker := pushKeywordCreature(t, g, n, 4, 4)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, n.ID); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, atk); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	if err := g.Concede(n.ID); err != nil {
		t.Fatal(err)
	}
	if g.State != StateEnded {
		t.Fatalf("state %s, want ended", g.State)
	}
	if g.Turn.ActiveSeat != 0 || g.Turn.Step != StepCombatDamage {
		t.Errorf("cursor moved after the game ended: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if c := findCard(g, blocker); c == nil || c.DamageMarked != 3 || c.BlockingTarget != atk {
		t.Errorf("final board was swept: %+v", c)
	}
}

// TestNoTurnBeginsInsideAResolution pins the constraint on every
// caller of the seam: the active player losing all their life while
// their own spell resolves does not end the turn inside the resolving
// callback. The loss is read by the resolution bookend's SBA pass, and
// the next turn begins there.
func TestNoTurnBeginsInsideAResolution(t *testing.T) {
	const oracle = "catalog-self-inflicted"
	var (
		sawSeat       = -1
		sawStep       Step
		sawEliminated bool
		sawUpkeep     bool
	)
	var g *Game
	withEffectHooks(t,
		func(gg *Game, item *StackItem, oracleID string) error {
			if oracleID != oracle {
				return nil
			}
			a := gg.Seats[0]
			if err := gg.ChangePlayerLifeForEffect(item.SourceCardID, a.ID, -a.Life); err != nil {
				return err
			}
			sawSeat, sawStep, sawEliminated = gg.Turn.ActiveSeat, gg.Turn.Step, a.Eliminated
			return nil
		},
		nil,
		func(oracleID string) bool { return oracleID == oracle },
	)
	g = newActiveGameWithSeats(t, 3)
	advanceTo(t, g, StepPrecombatMain)
	a := g.Seats[0]
	spell := pushTypedCardToHand(a, "Self-Inflicted", "Sorcery")
	for i := range a.Hand.Cards {
		if a.Hand.Cards[i].InstanceID == spell {
			a.Hand.Cards[i].OracleID = oracle
		}
	}
	if err := g.CastSpell(a.ID, spell, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	upkeepsBefore := 0
	for _, ev := range g.Events {
		if ev.Kind == EventBeginUpkeep {
			upkeepsBefore++
		}
	}
	for i := 0; i < 3 && sawSeat < 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if sawSeat < 0 {
		t.Fatal("the spell never resolved")
	}
	if sawSeat != 0 || sawStep != StepPrecombatMain || sawEliminated {
		t.Errorf("inside the resolution: seat %d step %s eliminated %v; the turn must not move until the bookend", sawSeat, sawStep, sawEliminated)
	}
	upkeeps := 0
	for _, ev := range g.Events {
		if ev.Kind == EventBeginUpkeep {
			upkeeps++
		}
	}
	sawUpkeep = upkeeps == upkeepsBefore+1
	if !a.Eliminated {
		t.Fatal("the bookend's SBA pass should have eliminated A")
	}
	if g.Turn.ActiveSeat != 1 || g.Turn.Step != StepUpkeep || g.Turn.PriorityHolder != 1 || !sawUpkeep {
		t.Errorf("after the bookend: seat %d step %s priority %d (one new upkeep: %v), want seat 1's upkeep", g.Turn.ActiveSeat, g.Turn.Step, g.Turn.PriorityHolder, sawUpkeep)
	}
}

// sbaLossWithMarkedBoard is the #834 review's probe: at a four-seat
// table the active player (seat 0) is at 0 life in the same state
// check where seat 1's 2/2 has lethal damage marked, seat 2's 5/5 has
// been dealt damage by a deathtouch source, and seat 3's 3/3 has one
// damage marked. That is the board an Earthquake that also kills its
// caster leaves behind.
type sbaLossWithMarkedBoard struct {
	g                             *Game
	lethal, deathtouched, bruised uuid.UUID
}

func newSBALossWithMarkedBoard(t *testing.T) sbaLossWithMarkedBoard {
	t.Helper()
	g := newFourPlayerActiveGame(t)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: active seat %d, want 0", g.Turn.ActiveSeat)
	}
	p := sbaLossWithMarkedBoard{
		g:            g,
		lethal:       pushKeywordCreature(t, g, g.Seats[1], 2, 2),
		deathtouched: pushKeywordCreature(t, g, g.Seats[2], 5, 5),
		bruised:      pushKeywordCreature(t, g, g.Seats[3], 3, 3),
	}
	g.WithWriteLock(func() {
		findCard(g, p.lethal).DamageMarked = 2
		dt := findCard(g, p.deathtouched)
		dt.DamageMarked = 1
		dt.MarkedLethalByDeathtouch = true
		findCard(g, p.bruised).DamageMarked = 1
		g.Seats[0].Life = 0
	})
	return p
}

func countEvents(g *Game, kind EventKind) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}

// TestActiveSeatSBALossStillDestroysMarkedCreatures: the loss and the
// destruction SBAs are one event (CR 704.3), so the turn the loss ends
// is not swept until the lethal-damage (CR 704.5g) and deathtouch (CR
// 704.5h) SBAs of that pass have been performed. Before the fix the
// rotation ran between them, cleared the marks, and both creatures
// lived.
func TestActiveSeatSBALossStillDestroysMarkedCreatures(t *testing.T) {
	p := newSBALossWithMarkedBoard(t)
	g := p.g
	upkeepsBefore := countEvents(g, EventBeginUpkeep)

	g.WithWriteLock(func() { g.runStateChecksLocked() })

	if !g.Seats[0].Eliminated {
		t.Fatal("the active player at 0 life should have lost")
	}
	if findCard(g, p.lethal) != nil {
		t.Error("the 2/2 with 2 damage marked survived the pass that ended the turn (CR 704.5g)")
	}
	if findCard(g, p.deathtouched) != nil {
		t.Error("the 5/5 dealt deathtouch damage survived the pass that ended the turn (CR 704.5h)")
	}
	// The turn did end: the survivor's damage is swept, and play moved
	// on once, to seat 1's upkeep.
	if c := findCard(g, p.bruised); c == nil || c.DamageMarked != 0 {
		t.Errorf("the 3/3 should survive with its damage swept: %+v", c)
	}
	if g.State != StateActive {
		t.Fatalf("three players remain: %s", g.State)
	}
	if g.Turn.ActiveSeat != 1 || g.Turn.Step != StepUpkeep || g.Turn.PriorityHolder != 1 {
		t.Errorf("cursor: seat %d step %s priority %d, want seat 1's upkeep with priority", g.Turn.ActiveSeat, g.Turn.Step, g.Turn.PriorityHolder)
	}
	if got := countEvents(g, EventBeginUpkeep) - upkeepsBefore; got != 1 {
		t.Errorf("%d upkeeps began, want exactly one", got)
	}
	// The deaths belong to the turn that ended: they are logged before
	// the new upkeep, and the new turn's tally starts from zero.
	lastDeath, upkeep := -1, -1
	for i, ev := range g.Events {
		switch {
		case ev.Kind == EventLTB && ev.NewZone == ZoneGraveyard && (ev.CardID == p.lethal || ev.CardID == p.deathtouched):
			lastDeath = i
		case ev.Kind == EventBeginUpkeep:
			upkeep = i
		}
	}
	if lastDeath < 0 || lastDeath > upkeep {
		t.Errorf("deaths at event %d, new upkeep at %d: the destruction must come before the turn ends", lastDeath, upkeep)
	}
	if g.TurnTally.CreaturesDied != 0 {
		t.Errorf("the new turn's tally counted the ended turn's deaths: %d", g.TurnTally.CreaturesDied)
	}
}

// TestSBALossQueuesTheLegendRuleBeforeTheTurnEnds: the legend rule is
// performed in the same pass as the loss, so its prompt is queued on
// the ended turn's board and stays open into the next turn; the turn
// still moves on exactly once.
func TestSBALossQueuesTheLegendRuleBeforeTheTurnEnds(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[1].ID
	a := pushLegendForTest(g, owner, "Teferi, Temporal Pilgrim", "Legendary Planeswalker — Teferi")
	b := pushLegendForTest(g, owner, "Teferi, Temporal Pilgrim", "Legendary Planeswalker — Teferi")
	upkeepsBefore := countEvents(g, EventBeginUpkeep)
	g.WithWriteLock(func() {
		g.Seats[0].Life = 0
		g.runStateChecksLocked()
	})
	if !g.Seats[0].Eliminated {
		t.Fatal("the active player at 0 life should have lost")
	}
	if legendPromptFor(g, owner) == nil {
		t.Fatal("no legend-rule prompt was queued")
	}
	if !onBattlefieldByID(g, a) || !onBattlefieldByID(g, b) {
		t.Error("a legend left before the prompt was answered")
	}
	if g.Turn.ActiveSeat != 1 || g.Turn.Step != StepUpkeep {
		t.Errorf("cursor: seat %d step %s, want seat 1's upkeep", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if got := countEvents(g, EventBeginUpkeep) - upkeepsBefore; got != 1 {
		t.Errorf("%d upkeeps began, want exactly one", got)
	}
}

// A single SBA pass is insufficient when one death removes a toughness
// bonus: cleanup must preserve the other creature's damage until the next
// check destroys it too.
func TestActiveSeatSBALossSettlesChainedLethalDamageBeforeCleanup(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[2].ID
	withStaticAbilities(t, func(id string) []StaticAbility {
		if id != "rotation-lord" {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer7PT, SubLayer: SubLayer7C_Modify,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.InstanceID != source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power++
				c.Toughness++
			},
		}}
	})
	lord := pushTypedTestCard(g, Card{Name: "Lord", OracleID: "rotation-lord", TypeLine: "Creature", Power: 2, Toughness: 2, Owner: owner, Controller: owner, DamageMarked: 2})
	bear := pushTypedTestCard(g, Card{Name: "Bear", TypeLine: "Creature", Power: 2, Toughness: 2, Owner: owner, Controller: owner, DamageMarked: 2})
	upkeepsBefore := countEvents(g, EventBeginUpkeep)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if got := findCard(g, bear).CurrentToughness(); got != 3 {
			t.Fatalf("setup: bear toughness = %d, want 3", got)
		}
		g.Seats[0].Life = 0
		g.runStateChecksLocked()
	})
	for _, id := range []uuid.UUID{lord, bear} {
		if g.Battlefield.Contains(id) || !g.Seats[2].Graveyard.Contains(id) {
			t.Errorf("lethally damaged creature %s did not reach its owner's graveyard", id)
		}
	}
	if g.Turn.ActiveSeat != 1 || g.Turn.Step != StepUpkeep {
		t.Errorf("cursor: seat %d step %s, want seat 1's upkeep", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if got := countEvents(g, EventBeginUpkeep) - upkeepsBefore; got != 1 {
		t.Errorf("%d upkeeps began, want exactly one", got)
	}
	lordDeath, bearDeath, upkeep := -1, -1, -1
	for i, ev := range g.Events {
		switch {
		case ev.Kind == EventLTB && ev.NewZone == ZoneGraveyard && ev.CardID == lord:
			lordDeath = i
		case ev.Kind == EventLTB && ev.NewZone == ZoneGraveyard && ev.CardID == bear:
			bearDeath = i
		case ev.Kind == EventBeginUpkeep:
			upkeep = i
		}
	}
	if lordDeath < 0 || bearDeath <= lordDeath || upkeep <= bearDeath {
		t.Errorf("events: lord death %d, bear death %d, upkeep %d; both SBA passes must finish before the new turn", lordDeath, bearDeath, upkeep)
	}
	if g.TurnTally.CreaturesDied != 0 {
		t.Errorf("new turn inherited %d deaths from the ended turn", g.TurnTally.CreaturesDied)
	}
}
