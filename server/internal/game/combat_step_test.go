package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// combat_step_test.go pins Event.CombatStep (#187, ADR 0053 Decision 1
// and sub-PR 1): every combat EventDealDamage names the combat damage
// step that dealt it — "first_strike" or "regular" (CR 510.4) — but
// ONLY when the first-strike pass ran. A combat with no first strike or
// double strike anywhere stays untagged, including damage that lands
// later from a CR 510.1c assignment prompt.
//
// The scenarios are the ADR's acceptance criteria, with its corrected
// arithmetic: Fencing Ace is a 1/1 double strike, Youthful Knight a 2/1
// first strike, Grizzly Bears a 2/2.
//
// Two criteria depend on open engine bugs and are written as skipped
// tests that assert the RULES outcome, so they fail loudly the day the
// skip is removed without the fix — they never pin today's wrong
// behaviour:
//
//   - #715 (a blocked attacker whose blockers are gone is treated as
//     unblocked): acceptance criterion 2.
//   - #702 (a first-strike multi-blocker prompt resolves after the
//     regular pass): the ORDER half of the first-strike prompt case.
//     Its tag half is tested now.
//
// #716 (keywords read after the recompute between passes) changes
// which creatures deal damage in the regular pass, not how that damage
// is tagged, so nothing here depends on it.

// pushCombatant puts a named creature on the battlefield with printed
// keywords. Unlike pushKeywordCreature the keywords are on
// Card.Keywords, the printed-data road, so they survive the layer
// recompute resolveCombatDamageLocked runs between the two passes when
// something died in the first one.
func pushCombatant(t *testing.T, g *Game, owner *Player, name string, power, toughness int, keywords ...string) uuid.UUID {
	t.Helper()
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Test"
	c.Power = power
	c.Toughness = toughness
	c.Keywords = append([]string(nil), keywords...)
	g.Battlefield.PushTop(c)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == c.InstanceID {
			g.Battlefield.Cards[i].effective = printedEffectiveWith(c)
		}
	}
	return c.InstanceID
}

// combatDamageSince returns the combat EventDealDamage events emitted
// after seq, in emission order.
func combatDamageSince(g *Game, seq uint64) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Seq > seq && ev.Kind == EventDealDamage && ev.Combat {
			out = append(out, ev)
		}
	}
	return out
}

// lastSeq is the Seq of the newest event, the watermark combatDamageSince
// reads from.
func lastSeq(g *Game) uint64 {
	if len(g.Events) == 0 {
		return 0
	}
	return g.Events[len(g.Events)-1].Seq
}

// wantDamage is one expected combat damage event.
type wantDamage struct {
	source, target uuid.UUID
	amount         int
	step           string
}

// assertCombatDamage checks got against want, in order.
func assertCombatDamage(t *testing.T, got []Event, want []wantDamage) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%d combat damage events, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		ev := got[i]
		if ev.Source != w.source || ev.Target != w.target || ev.Amount != w.amount {
			t.Errorf("event %d: %s -> %s for %d, want %s -> %s for %d",
				i, ev.Source, ev.Target, ev.Amount, w.source, w.target, w.amount)
		}
		if ev.CombatStep != w.step {
			t.Errorf("event %d (%s -> %s for %d): CombatStep = %q, want %q",
				i, ev.Source, ev.Target, ev.Amount, ev.CombatStep, w.step)
		}
	}
}

// declareCombat drives a game from the start of turn into the combat
// damage step: attackers declared against seat 1, then blocks as
// blocker -> attacker pairs. Returns the event watermark taken just
// before the combat damage step began.
func declareCombat(t *testing.T, g *Game, attackers []uuid.UUID, blocks map[uuid.UUID]uuid.UUID) uint64 {
	t.Helper()
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	for blk, atk := range blocks {
		if err := g.DeclareBlocker(blk, atk); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	seq := lastSeq(g)
	advanceThroughDamageSteps(t, g)
	return seq
}

// advanceThroughDamageSteps walks the cursor from declare_blockers
// through both combat damage steps (CR 510.4, #717) and stops at the
// regular one — or earlier, at whichever damage step queued a prompt
// that blocks the table. A pending CR 510.1c assignment or CR 616
// ordering prompt holds the cursor where it is, which is the #702
// fix: the regular step cannot run while the first step still owes a
// damage assignment.
func advanceThroughDamageSteps(t *testing.T, g *Game) {
	t.Helper()
	for g.Turn.Step != StepCombatDamage {
		before := g.Turn
		if _, err := g.AdvanceStep(); err != nil {
			if errors.Is(err, ErrChoicePending) {
				return
			}
			t.Fatalf("AdvanceStep to %v: %v", StepCombatDamage, err)
		}
		if g.Turn == before {
			t.Fatalf("advanceThroughDamageSteps stuck at %v", before)
		}
	}
}

// pendingOfKind returns the first queued choice of kind, or fails.
func pendingOfKind(t *testing.T, g *Game, kind PendingChoiceKind) *PendingChoice {
	t.Helper()
	for _, pc := range g.PendingChoices {
		if pc.Kind == kind {
			return pc
		}
	}
	t.Fatalf("no %q prompt queued (have %d choices)", kind, len(g.PendingChoices))
	return nil
}

// Acceptance criterion 3 — first strike. Youthful Knight kills the
// Bears in the first-strike step; the regular step runs and deals
// nothing, so there is exactly one tagged event and no regular one.
func TestCombatStepFirstStrikeKillsBlocker(t *testing.T) {
	g := newActiveGame(t)
	knight := pushCombatant(t, g, g.Seats[0], "Youthful Knight", 2, 1, "first strike")
	bears := pushCombatant(t, g, g.Seats[1], "Grizzly Bears", 2, 2)
	seq := declareCombat(t, g, []uuid.UUID{knight}, map[uuid.UUID]uuid.UUID{bears: knight})

	assertCombatDamage(t, combatDamageSince(g, seq), []wantDamage{
		{knight, bears, 2, CombatStepFirstStrike},
	})
	if findCard(g, bears) != nil {
		t.Error("Grizzly Bears survived 2 first-strike damage")
	}
	if k := findCard(g, knight); k == nil || k.DamageMarked != 0 {
		t.Errorf("Youthful Knight should survive with 0 damage, got %+v", k)
	}
}

// Acceptance criterion 1 — double strike, the blocker survives the
// first step. Ace deals 1 (first_strike); then Ace deals 1 more and the
// Bears deal 2 back (both regular), and both die.
func TestCombatStepDoubleStrikeBlockerSurvivesFirstStep(t *testing.T) {
	g := newActiveGame(t)
	ace := pushCombatant(t, g, g.Seats[0], "Fencing Ace", 1, 1, "double strike")
	bears := pushCombatant(t, g, g.Seats[1], "Grizzly Bears", 2, 2)
	seq := declareCombat(t, g, []uuid.UUID{ace}, map[uuid.UUID]uuid.UUID{bears: ace})

	assertCombatDamage(t, combatDamageSince(g, seq), []wantDamage{
		{ace, bears, 1, CombatStepFirstStrike},
		{ace, bears, 1, CombatStepRegular},
		{bears, ace, 2, CombatStepRegular},
	})
	if findCard(g, bears) != nil || findCard(g, ace) != nil {
		t.Error("Fencing Ace and Grizzly Bears should both die in the regular step")
	}
}

// Acceptance criterion 4 — double strike to a player. Two events to the
// defending seat, 1 each, first_strike then regular; 40 -> 38.
func TestCombatStepDoubleStrikeUnblocked(t *testing.T) {
	g := newActiveGame(t)
	ace := pushCombatant(t, g, g.Seats[0], "Fencing Ace", 1, 1, "double strike")
	seq := declareCombat(t, g, []uuid.UUID{ace}, nil)

	def := g.Seats[1].ID
	assertCombatDamage(t, combatDamageSince(g, seq), []wantDamage{
		{ace, def, 1, CombatStepFirstStrike},
		{ace, def, 1, CombatStepRegular},
	})
	if got := g.Seats[1].Life; got != StartingLife-2 {
		t.Errorf("defender life = %d, want %d", got, StartingLife-2)
	}
}

// Acceptance criterion 5 — mixed combat. Double-strike, first-strike
// and vanilla attackers, each unblocked or blocked by one creature that
// is still there at damage. Every first-strike event comes before every
// regular event, and the vanilla creatures' damage is tagged regular
// because the combat had a first-strike step.
//
// Multi-blocker first strike (#702) and blockers removed before damage
// (#715) are excluded, as the criterion says.
func TestCombatStepMixedCombatOrdersFirstStrikeFirst(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	ace := pushCombatant(t, g, atk, "Fencing Ace", 1, 1, "double strike")
	knight := pushCombatant(t, g, atk, "Youthful Knight", 2, 1, "first strike")
	vanilla := pushCombatant(t, g, atk, "Vanilla Attacker", 3, 3)
	wall := pushCombatant(t, g, def, "Wall", 0, 5)
	bears := pushCombatant(t, g, def, "Grizzly Bears", 2, 2)
	// Ace unblocked; Knight blocked by a 0/5 that survives; the vanilla
	// 3/3 blocked by the Bears, which survive to deal regular damage.
	seq := declareCombat(t, g, []uuid.UUID{ace, knight, vanilla},
		map[uuid.UUID]uuid.UUID{wall: knight, bears: vanilla})

	got := combatDamageSince(g, seq)
	var lastFirst, firstRegular uint64
	counts := map[string]int{}
	for _, ev := range got {
		counts[ev.CombatStep]++
		switch ev.CombatStep {
		case CombatStepFirstStrike:
			lastFirst = ev.Seq
		case CombatStepRegular:
			if firstRegular == 0 {
				firstRegular = ev.Seq
			}
		default:
			t.Errorf("untagged combat damage %+v in a combat with a first-strike step", ev)
		}
	}
	if lastFirst == 0 || firstRegular == 0 || lastFirst > firstRegular {
		t.Errorf("first_strike events must all precede regular ones: last first_strike seq %d, first regular seq %d",
			lastFirst, firstRegular)
	}
	// first_strike: Ace -> defender 1, Knight -> Wall 2.
	// regular: Ace -> defender 1, vanilla -> Bears 3, Bears -> vanilla 2.
	if counts[CombatStepFirstStrike] != 2 || counts[CombatStepRegular] != 3 {
		t.Errorf("tag counts = %v, want 2 first_strike and 3 regular (events %+v)", counts, got)
	}
	for _, ev := range got {
		if (ev.Source == vanilla || ev.Source == bears) && ev.CombatStep != CombatStepRegular {
			t.Errorf("vanilla creature's damage %+v tagged %q, want regular", ev, ev.CombatStep)
		}
		if ev.Source == knight && ev.CombatStep != CombatStepFirstStrike {
			t.Errorf("first striker's damage %+v tagged %q, want first_strike", ev, ev.CombatStep)
		}
	}
}

// Acceptance criterion 6 (direct half) — no first strike anywhere.
// Unblocked and blocked damage, attacker and blocker side, all
// untagged: the wire says there is nothing to sequence.
func TestCombatStepVanillaCombatIsUntagged(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	unblocked := pushCombatant(t, g, atk, "Vanilla Attacker", 3, 3)
	blocked := pushCombatant(t, g, atk, "Blocked Attacker", 2, 2)
	blocker := pushCombatant(t, g, def, "Blocker", 4, 4)
	seq := declareCombat(t, g, []uuid.UUID{unblocked, blocked}, map[uuid.UUID]uuid.UUID{blocker: blocked})

	got := combatDamageSince(g, seq)
	if len(got) != 3 {
		t.Fatalf("%d combat damage events, want 3: %+v", len(got), got)
	}
	for _, ev := range got {
		if ev.CombatStep != "" {
			t.Errorf("combat with no first strike tagged %+v with %q, want untagged", ev, ev.CombatStep)
		}
	}
}

// Non-combat damage never carries a tag, even dealt during a combat
// damage step that had a first-strike pass.
func TestCombatStepNeverOnNonCombatDamage(t *testing.T) {
	g := newActiveGame(t)
	knight := pushCombatant(t, g, g.Seats[0], "Youthful Knight", 2, 1, "first strike")
	_ = declareCombat(t, g, []uuid.UUID{knight}, nil)

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(knight, g.Seats[1].ID, 1)
	})
	last := g.Events[len(g.Events)-1]
	if last.Kind != EventDealDamage || last.Combat || last.CombatStep != "" {
		t.Errorf("effect damage event: kind=%s combat=%v step=%q, want deal_damage / false / \"\"",
			last.Kind, last.Combat, last.CombatStep)
	}
}

// Prompt resume, regular pass with the first-strike pass run: a
// trampling 6/6 blocked by two creatures, in a combat that also has an
// unblocked first striker. The frame stores "regular" (FirstStrike
// false), and both frame paths — damage to each blocker and the
// trample spill to the player — come back tagged regular.
func TestCombatStepAssignmentPromptResumesRegular(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	knight := pushCombatant(t, g, atk, "Youthful Knight", 2, 1, "first strike")
	wurm := pushCombatant(t, g, atk, "Trampling Wurm", 6, 6, "trample")
	bear1 := pushCombatant(t, g, def, "Bear One", 2, 2)
	bear2 := pushCombatant(t, g, def, "Bear Two", 3, 3)
	_ = declareCombat(t, g, []uuid.UUID{knight, wurm}, map[uuid.UUID]uuid.UUID{bear1: wurm, bear2: wurm})

	prompt := pendingOfKind(t, g, PendingChoiceDamageAssignment)
	frame := prompt.DamageAssignment
	if frame.FirstStrike {
		t.Error("frame.FirstStrike = true for a regular-pass prompt")
	}
	if frame.CombatStep != CombatStepRegular {
		t.Fatalf("frame.CombatStep = %q, want %q", frame.CombatStep, CombatStepRegular)
	}

	seq := lastSeq(g)
	if err := g.ResolveDamageAssignment(prompt.ID, atk.ID, []DamageAssignmentEntry{
		{BlockerID: bear1, Amount: 2},
		{BlockerID: bear2, Amount: 3},
	}, 1); err != nil {
		t.Fatalf("ResolveDamageAssignment: %v", err)
	}
	assertCombatDamage(t, combatDamageSince(g, seq), []wantDamage{
		{wurm, bear1, 2, CombatStepRegular},
		{wurm, bear2, 3, CombatStepRegular},
		{wurm, def.ID, 1, CombatStepRegular},
	})
}

// Acceptance criterion 6 (prompt half) — a vanilla 5/5 blocked by two
// 2/2s in a combat with no first strike. The frame stores "" and the
// resumed damage is untagged. Tagging from FirstStrike (false) would
// wrongly say "regular" here, which is why the frame has its own field.
func TestCombatStepAssignmentPromptWithoutFirstStrikeIsUntagged(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	big := pushCombatant(t, g, atk, "Vanilla 5/5", 5, 5)
	bear1 := pushCombatant(t, g, def, "Bear One", 2, 2)
	bear2 := pushCombatant(t, g, def, "Bear Two", 2, 2)
	before := declareCombat(t, g, []uuid.UUID{big}, map[uuid.UUID]uuid.UUID{bear1: big, bear2: big})

	prompt := pendingOfKind(t, g, PendingChoiceDamageAssignment)
	if got := prompt.DamageAssignment.CombatStep; got != "" {
		t.Fatalf("frame.CombatStep = %q, want untagged", got)
	}
	if err := g.ResolveDamageAssignment(prompt.ID, atk.ID, []DamageAssignmentEntry{
		{BlockerID: bear1, Amount: 2},
		{BlockerID: bear2, Amount: 3},
	}, 0); err != nil {
		t.Fatalf("ResolveDamageAssignment: %v", err)
	}
	got := combatDamageSince(g, before)
	// Two blockers -> attacker before the prompt, two attacker ->
	// blocker from the resume.
	if len(got) != 4 {
		t.Fatalf("%d combat damage events, want 4: %+v", len(got), got)
	}
	for _, ev := range got {
		if ev.CombatStep != "" {
			t.Errorf("event %+v tagged %q in a combat with no first strike", ev, ev.CombatStep)
		}
	}
}

// Prompt resume, first-strike step: a 5/5 first striker blocked by two
// 2/2s. The TAG half — the frame says first_strike and the resumed
// damage is tagged first_strike. The ORDER half is the test below.
func TestCombatStepFirstStrikeAssignmentPromptTag(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	striker := pushCombatant(t, g, atk, "First Striker", 5, 5, "first strike")
	bear1 := pushCombatant(t, g, def, "Bear One", 2, 2)
	bear2 := pushCombatant(t, g, def, "Bear Two", 2, 2)
	_ = declareCombat(t, g, []uuid.UUID{striker}, map[uuid.UUID]uuid.UUID{bear1: striker, bear2: striker})

	prompt := pendingOfKind(t, g, PendingChoiceDamageAssignment)
	frame := prompt.DamageAssignment
	if !frame.FirstStrike || frame.CombatStep != CombatStepFirstStrike {
		t.Fatalf("frame FirstStrike=%v CombatStep=%q, want true / %q",
			frame.FirstStrike, frame.CombatStep, CombatStepFirstStrike)
	}
	seq := lastSeq(g)
	if err := g.ResolveDamageAssignment(prompt.ID, atk.ID, []DamageAssignmentEntry{
		{BlockerID: bear1, Amount: 2},
		{BlockerID: bear2, Amount: 3},
	}, 0); err != nil {
		t.Fatalf("ResolveDamageAssignment: %v", err)
	}
	assertCombatDamage(t, combatDamageSince(g, seq), []wantDamage{
		{striker, bear1, 2, CombatStepFirstStrike},
		{striker, bear2, 3, CombatStepFirstStrike},
	})
}

// The ORDER half of the first-strike prompt case, and the #702
// regression. By CR 510.4 the first-strike step's damage is all dealt
// before the regular step begins, so every first_strike event precedes
// every regular one. The engine used to run both passes back to back
// inside one cursor step, so the regular pass ran while the assignment
// prompt was still open and the attacker could die before assigning.
// Now the passes are two steps and the prompt blocks the table, so the
// cursor cannot leave the first one until it is answered.
func TestCombatStepFirstStrikeAssignmentPromptOrder(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	striker := pushCombatant(t, g, atk, "First Striker", 5, 5, "first strike")
	bear1 := pushCombatant(t, g, def, "Bear One", 2, 2)
	bear2 := pushCombatant(t, g, def, "Bear Two", 3, 3)
	before := declareCombat(t, g, []uuid.UUID{striker}, map[uuid.UUID]uuid.UUID{bear1: striker, bear2: striker})

	// The cursor is held in the first combat damage step by the
	// unanswered prompt — the regular step has not begun.
	if g.Turn.Step != StepFirstStrikeDamage {
		t.Fatalf("step with the assignment prompt open: got %q, want %q", g.Turn.Step, StepFirstStrikeDamage)
	}
	if got := combatDamageSince(g, before); len(got) != 0 {
		t.Fatalf("damage landed before the assignment prompt was answered: %+v", got)
	}
	prompt := pendingOfKind(t, g, PendingChoiceDamageAssignment)
	if err := g.ResolveDamageAssignment(prompt.ID, atk.ID, []DamageAssignmentEntry{
		{BlockerID: bear1, Amount: 2},
		{BlockerID: bear2, Amount: 3},
	}, 0); err != nil {
		t.Fatalf("ResolveDamageAssignment: %v", err)
	}
	advanceThroughDamageSteps(t, g)
	// Both blockers die to first-strike damage and never deal regular
	// damage, so the whole combat is the two first_strike events.
	assertCombatDamage(t, combatDamageSince(g, before), []wantDamage{
		{striker, bear1, 2, CombatStepFirstStrike},
		{striker, bear2, 3, CombatStepFirstStrike},
	})
}

// Acceptance criterion 2 — double strike, the blocker dies in the first
// step. By the rules (CR 509.1h, 510.1c) Ace is still blocked and deals
// no regular damage: one first_strike event, the defender stays at 40.
// Today the engine treats Ace as unblocked in the regular pass and hits
// the defender for 1 (#715). ADR 0053 says sub-PR 1 does not pin that
// wrong behaviour, so this asserts the rules outcome behind a skip.
func TestCombatStepDoubleStrikeBlockerDiesInFirstStep(t *testing.T) {
	t.Skip("#715: a blocked attacker whose blockers are all gone is treated as unblocked; " +
		"remove this skip with the fix")
	g := newActiveGame(t)
	ace := pushCombatant(t, g, g.Seats[0], "Fencing Ace", 1, 1, "double strike")
	chump := pushCombatant(t, g, g.Seats[1], "Vanilla 1/1", 1, 1)
	seq := declareCombat(t, g, []uuid.UUID{ace}, map[uuid.UUID]uuid.UUID{chump: ace})

	assertCombatDamage(t, combatDamageSince(g, seq), []wantDamage{
		{ace, chump, 1, CombatStepFirstStrike},
	})
	if got := g.Seats[1].Life; got != StartingLife {
		t.Errorf("defender life = %d, want %d", got, StartingLife)
	}
}

// A frame restored from a snapshot written before
// DamageAssignmentFrame.CombatStep existed decodes with "" and resumes
// untagged, even when it was queued by a regular pass that followed a
// first-strike pass. The cue is lost; the board is right. (The JSON
// half — that such a file really decodes to "" — is
// TestSnapshotWithoutCombatStepRestoresUntagged.)
func TestCombatStepPreFieldAssignmentPromptResumesUntagged(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	knight := pushCombatant(t, g, atk, "Youthful Knight", 2, 1, "first strike")
	big := pushCombatant(t, g, atk, "Vanilla 5/5", 5, 5)
	bear1 := pushCombatant(t, g, def, "Bear One", 2, 2)
	bear2 := pushCombatant(t, g, def, "Bear Two", 3, 3)
	_ = declareCombat(t, g, []uuid.UUID{knight, big}, map[uuid.UUID]uuid.UUID{bear1: big, bear2: big})

	prompt := pendingOfKind(t, g, PendingChoiceDamageAssignment)
	// What a pre-field snapshot decodes to.
	prompt.DamageAssignment.CombatStep = ""

	seq := lastSeq(g)
	if err := g.ResolveDamageAssignment(prompt.ID, atk.ID, []DamageAssignmentEntry{
		{BlockerID: bear1, Amount: 2},
		{BlockerID: bear2, Amount: 3},
	}, 0); err != nil {
		t.Fatalf("ResolveDamageAssignment: %v", err)
	}
	assertCombatDamage(t, combatDamageSince(g, seq), []wantDamage{
		{big, bear1, 2, ""},
		{big, bear2, 3, ""},
	})
}

// A CR 616 ordering prompt in the first-strike step: two replacements
// apply to the first striker's damage, so it pauses, and it lands when
// the order is answered. The prompt blocks the table, so the cursor is
// still in the first-strike step when the damage lands — and the tag
// rides the damage tail on the paused event either way, so it comes
// back first_strike and not whatever step is current.
func TestCombatStepReplacementOrderPromptKeepsFirstStrike(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	knight := pushCombatant(t, g, atk, "Youthful Knight", 2, 1, "first strike")
	g.WithWriteLock(func() {
		for _, label := range []string{"Prevent 1", "Double"} {
			g.RegisterReplacementForTest(ReplacementEffect{
				Watches: []EventKind{EventDealDamage},
				AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
					return ev.Kind == RepEventDamage && ev.IsCombatDamage
				},
				Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
					if label == "Prevent 1" {
						ev.DamageAmount--
					} else {
						ev.DamageAmount *= 2
					}
					return nil
				},
				Label: label,
			})
		}
	})
	before := declareCombat(t, g, []uuid.UUID{knight}, nil)

	if got := combatDamageSince(g, before); len(got) != 0 {
		t.Fatalf("damage landed before the CR 616 prompt was answered: %+v", got)
	}
	prompt := pendingOfKind(t, g, PendingChoiceReplacementOrder)
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	// (2 - 1) * 2 = 2 in the order offered.
	assertCombatDamage(t, combatDamageSince(g, before), []wantDamage{
		{knight, def.ID, 2, CombatStepFirstStrike},
	})
}
