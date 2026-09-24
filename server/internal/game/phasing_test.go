package game

import (
	"testing"

	"github.com/google/uuid"
)

// phasing_test.go — CR 702.26, #1199, ADR 0084.
//
// One test per SHAPE of the rule rather than one per card, because
// the design's whole claim is that the shapes are the engine's and
// the cards say nothing. Each test names the rule it pins, and each
// fails for a different reason if a piece of phasing.go is reverted.

// pushPhasingTestCard parks an untapped permanent on the battlefield
// under `controller` and announces the entry, so the layer listener
// stamps EnteredBattlefieldAt the way a real entry does.
func pushPhasingTestCard(g *Game, controller uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Owner:      controller,
		Controller: controller,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			CardID:  id,
			OldZone: ZoneHand,
			NewZone: ZoneBattlefield,
		})
	})
	return id
}

// phasedOutCardByID returns a copy of the named card from the holding
// slice.
func phasedOutCardByID(g *Game, id uuid.UUID) (Card, bool) {
	var out Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.PhasedOut.Cards {
			if c.InstanceID == id {
				out, found = c, true
				return
			}
		}
	})
	return out, found
}

func eventsOfKind(g *Game, kind EventKind) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}

// seatOf returns the seat index of a player ID.
func seatOf(g *Game, player uuid.UUID) int {
	for i, p := range g.Seats {
		if p.ID == player {
			return i
		}
	}
	return -1
}

// --- the round trip -------------------------------------------------

// CR 702.26b then CR 502.1: a permanent phases out, is gone from the
// battlefield, and comes back during its controller's untap step —
// before that player untaps (CR 502.3).
func TestPhaseOutAndInAcrossTheControllersUntapStep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	bear := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, bear) })

	if _, ok := battlefieldCardByID(g, bear); ok {
		t.Fatal("a phased-out permanent is out of the battlefield slice (CR 702.26b)")
	}
	if _, ok := phasedOutCardByID(g, bear); !ok {
		t.Fatal("it is in the holding slice")
	}

	// Somebody else's untap step does nothing: CR 502.1 phases in the
	// permanents the ACTIVE player controlled when they phased out.
	g.WithWriteLock(func() { g.performPhasingLocked(g.Seats[1].ID) })
	if _, ok := battlefieldCardByID(g, bear); ok {
		t.Fatal("it does not phase in during another player's untap step (CR 502.1)")
	}

	g.WithWriteLock(func() { g.performPhasingLocked(me) })
	if _, ok := battlefieldCardByID(g, bear); !ok {
		t.Fatal("it phases in during its controller's untap step (CR 502.1)")
	}
	if _, ok := phasedOutCardByID(g, bear); ok {
		t.Fatal("and leaves the holding slice")
	}
}

// CR 502.1 runs BEFORE CR 502.3, so a permanent that phased out
// tapped phases in tapped and then untaps in the same step.
func TestAPermanentPhasesInBeforeTheUntapStepUntaps(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	rock := pushTappedPermanent(g, me, "Tapped Rock", "", "Artifact", true)

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, rock) })
	if c, ok := phasedOutCardByID(g, rock); !ok || !c.Tapped {
		t.Fatal("phasing out does not untap it (CR 702.26d, CR 110.5c)")
	}

	g.WithWriteLock(func() { g.performUntapStepLocked(seatOf(g, me)) })

	c, ok := battlefieldCardByID(g, rock)
	if !ok {
		t.Fatal("it phased in")
	}
	if c.Tapped {
		t.Error("CR 502.1 is ahead of CR 502.3, so the permanent it brought back untaps in the same step")
	}
}

// CR 702.26d: "Zone-change triggers don't trigger when a permanent
// phases in or out." No zone move, no LTB, no ETB, at either end.
func TestPhasingEmitsNoZoneChangeEvents(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	bear := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")

	before := map[EventKind]int{
		EventZoneMove: eventsOfKind(g, EventZoneMove),
		EventLTB:      eventsOfKind(g, EventLTB),
		EventETB:      eventsOfKind(g, EventETB),
	}

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, bear) })
	g.WithWriteLock(func() { g.performPhasingLocked(me) })

	for kind, was := range before {
		if now := eventsOfKind(g, kind); now != was {
			t.Errorf("%s fired across a phase cycle (%d → %d); CR 702.26d says none does", kind, was, now)
		}
	}
	if n := eventsOfKind(g, EventPhaseOut); n != 1 {
		t.Errorf("one phase_out event, got %d", n)
	}
	if n := eventsOfKind(g, EventPhaseIn); n != 1 {
		t.Errorf("one phase_in event, got %d", n)
	}
}

// CR 702.26d, and the promise activation_tally.go has been making
// since S38: phasing is not a zone change, so the object epoch — the
// key every per-object registry in the engine hangs on — does not
// move. A flicker refreshes an exhaust ability; a phase-out does not.
func TestPhasingDoesNotBumpTheObjectEpoch(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	bear := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")
	before, _ := battlefieldCardByID(g, bear)

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, bear) })
	g.WithWriteLock(func() { g.performPhasingLocked(me) })

	after, ok := battlefieldCardByID(g, bear)
	if !ok {
		t.Fatal("it phased back in")
	}
	if after.ObjectEpoch != before.ObjectEpoch {
		t.Errorf("ObjectEpoch moved %d → %d across a phase cycle (CR 702.26d)",
			before.ObjectEpoch, after.ObjectEpoch)
	}
	if after.InstanceID != before.InstanceID {
		t.Error("it is the same object")
	}
}

// CR 702.26d: "Counters and stickers remain on a permanent while it's
// phased out", and so does its marked damage, its tapped state, its
// entry timestamp and its summoning sickness — none of which survives
// a real battlefield exit.
func TestPhasingKeepsEverythingABattlefieldExitWouldClear(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	bear := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != bear {
				continue
			}
			c.Counters = map[string]int{"+1/+1": 2}
			c.DamageMarked = 1
			c.SummonedThisTurn = true
		}
	})
	before, _ := battlefieldCardByID(g, bear)

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, bear) })
	out, ok := phasedOutCardByID(g, bear)
	if !ok {
		t.Fatal("it phased out")
	}
	if out.Counters["+1/+1"] != 2 {
		t.Errorf("counters stay while phased out (CR 702.26d): %v", out.Counters)
	}
	if out.DamageMarked != 1 {
		t.Errorf("marked damage stays: %d", out.DamageMarked)
	}
	if out.EnteredBattlefieldAt != before.EnteredBattlefieldAt {
		t.Error("the CR 613.7 timestamp stays — nothing entered anything")
	}

	// Phase in on somebody else's turn, so the CR 302.6 clear in the
	// untap step cannot be what cleared the flag.
	g.WithWriteLock(func() { g.performPhasingLocked(me) })
	in, ok := battlefieldCardByID(g, bear)
	if !ok {
		t.Fatal("it phased in")
	}
	if in.Counters["+1/+1"] != 2 || in.DamageMarked != 1 {
		t.Errorf("and comes back with them: counters=%v damage=%d", in.Counters, in.DamageMarked)
	}
	if !in.SummonedThisTurn {
		t.Error("CR 702.26d: phasing in is not entering, so it does not reset summoning sickness either way")
	}
}

// A phased-out permanent is absent from the battlefield slice when its
// controller's next turn begins, but CR 702.26d still treats it as the
// same permanent. The turn boundary must retire last turn's sickness
// marker before the untap action phases it back in.
func TestTurnBoundaryClearsSummoningSicknessWhilePermanentIsPhasedOut(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	bear := pushPhasingTestCard(g, me, "Patient Bear", "Creature — Bear")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].SummonedThisTurn = true
			}
		}
		_ = g.PhaseOutForEffect(uuid.Nil, bear)
		g.Turn.Seq++
		g.onTurnBeganLocked()
		g.performPhasingLocked(me)
	})

	in, ok := battlefieldCardByID(g, bear)
	if !ok {
		t.Fatal("permanent did not phase in during its controller's untap action")
	}
	if in.SummonedThisTurn {
		t.Fatal("phased-out permanent retained last turn's summoning-sickness marker")
	}
}

// CR 702.26d again, from the other side: control and ownership do not
// change, and CR 702.26a's phase-in reads the controller AT PHASE-OUT
// rather than the live one.
func TestPhasedOutByIsTheControllerAtPhaseOutNotTheLiveOne(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[0].ID
	thief := g.Seats[1].ID
	bear := pushPhasingTestCard(g, owner, "Grizzly Bears", "Creature — Bear")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Controller = thief
			}
		}
	})

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, bear) })
	out, ok := phasedOutCardByID(g, bear)
	if !ok {
		t.Fatal("it phased out")
	}
	if out.Owner != owner {
		t.Errorf("ownership is unchanged: %v", out.Owner)
	}
	if out.Controller != thief || out.PhasedOutBy != thief {
		t.Errorf("control is unchanged and recorded: controller=%v phasedOutBy=%v", out.Controller, out.PhasedOutBy)
	}

	// CR 702.26f: the control-changing effect ends while it is away.
	// CR 702.26a still phases it in during the THIEF's untap step,
	// which is why PhasedOutBy is stored rather than derived.
	g.WithWriteLock(func() {
		for i := range g.PhasedOut.Cards {
			if g.PhasedOut.Cards[i].InstanceID == bear {
				g.PhasedOut.Cards[i].Controller = owner
			}
		}
	})
	g.WithWriteLock(func() { g.performPhasingLocked(owner) })
	if _, ok := battlefieldCardByID(g, bear); ok {
		t.Fatal("it does not phase in under the player who merely controls it now (CR 702.26a)")
	}
	g.WithWriteLock(func() { g.performPhasingLocked(thief) })
	if _, ok := battlefieldCardByID(g, bear); !ok {
		t.Error("it phases in under the player who controlled it when it phased out (CR 702.26a)")
	}
}

// --- attachments ----------------------------------------------------

// CR 702.26g: "When a permanent phases out, any Auras, Equipment, or
// Fortifications attached to that permanent phase out at the same
// time … won't phase in by itself, but instead phases in along with
// the permanent it's attached to." Transitively, and still attached.
func TestAttachmentsPhaseOutWithTheirHostAndComeBackAttached(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")
	sword := pushPhasingTestCard(g, me, "Test Sword", "Artifact — Equipment")
	pacifism := pushPhasingTestCard(g, me, "Test Aura", "Enchantment — Aura")
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
			t.Fatalf("attach sword: %v", err)
		}
		// An Aura on the EQUIPMENT, so the walk has to be transitive.
		if err := g.AttachForEffect(pacifism, TargetRef{Kind: TargetCard, ID: sword}); err != nil {
			t.Fatalf("attach aura: %v", err)
		}
	})

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, bear) })

	for _, id := range []uuid.UUID{bear, sword, pacifism} {
		if _, ok := battlefieldCardByID(g, id); ok {
			t.Fatalf("%v is off the battlefield (CR 702.26g, transitively)", id)
		}
	}
	att, _ := phasedOutCardByID(g, sword)
	if !att.PhasedOutIndirect {
		t.Error("the Equipment phased out INDIRECTLY (CR 702.26g)")
	}
	host, _ := phasedOutCardByID(g, bear)
	if host.PhasedOutIndirect {
		t.Error("the host phased out directly")
	}

	// CR 702.26g: the attachment does not phase in by itself, so
	// asking for the host is enough and asking for the attachment
	// alone would be wrong. performPhasingLocked names only the
	// direct ones and grows the set.
	g.WithWriteLock(func() { g.performPhasingLocked(me) })
	for _, id := range []uuid.UUID{bear, sword, pacifism} {
		if _, ok := battlefieldCardByID(g, id); !ok {
			t.Fatalf("%v phased back in with its host", id)
		}
	}
	back, _ := battlefieldCardByID(g, sword)
	if !back.IsAttachedTo(bear) {
		t.Error("and is still attached: phasing is not a zone change, so nothing unattached it")
	}
	backAura, _ := battlefieldCardByID(g, pacifism)
	if !backAura.IsAttachedTo(sword) {
		t.Error("and so is the Aura on the Equipment")
	}
}

// CR 702.26h: "If an object would simultaneously phase out directly
// and indirectly, it just phases out indirectly." Clever Concealment
// naming both a creature and the Equipment on it is the printed case.
func TestNamedAndAttachedPhasesOutIndirectly(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")
	sword := pushPhasingTestCard(g, me, "Test Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, bear, sword) })

	att, ok := phasedOutCardByID(g, sword)
	if !ok {
		t.Fatal("the Equipment phased out")
	}
	if !att.PhasedOutIndirect {
		t.Error("named AND attached is indirect (CR 702.26h), so it cannot phase in without its host")
	}
	if n := eventsOfKind(g, EventPhaseOut); n != 2 {
		t.Errorf("each permanent phases out once, got %d events", n)
	}
}

// --- "treated as though it does not exist" --------------------------

// CR 702.26b, the whole point of ADR 0084 Decision 1: every
// battlefield walk in the engine is right about a phased-out
// permanent without being told about phasing.
func TestAPhasedOutPermanentIsInvisibleToTheBoard(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	gone := pushPhasingTestCard(g, me, "Phased Bear", "Creature — Bear")
	stays := pushPhasingTestCard(g, me, "Present Bear", "Creature — Bear")

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, gone) })

	// Not in the walk every catalog wrath uses.
	g.WithWriteLock(func() {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.InstanceID == gone {
				t.Error("BattlefieldCardsForEffect does not see it")
			}
		}
	})
	// Not targetable: the one targeting choke point walks the same
	// slice, so the bot's enumerator and the client's legal_targets
	// follow by construction.
	spec := &TargetSpec{Mode: "creature", Label: "target creature", Zones: []ZoneKind{ZoneBattlefield}}
	g.WithWriteLock(func() {
		lt := g.LegalTargetsForEffect(TargetSource{Controller: me}, spec)
		for _, id := range lt.Cards {
			if id == gone {
				t.Error("it is not a legal target (CR 702.26b)")
			}
		}
	})
	// Not where the game thinks it is: "where is this object" has no
	// answer, because as far as the game is concerned there is none.
	g.WithWriteLock(func() {
		if z := g.FindCardZoneForEffect(gone); z != nil {
			t.Errorf("FindCardZoneForEffect answers nil, got %q", z.Kind)
		}
		if !g.IsPhasedOutForEffect(gone) {
			t.Error("IsPhasedOutForEffect is the rule's own read surface and says yes")
		}
	})
	// Not destroyed by a board wipe that names every creature the
	// wipe could see.
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.IsCreature() {
				ids = append(ids, c.InstanceID)
			}
		}
		g.DestroyPermanentsForEffect(ids)
	})
	if _, ok := phasedOutCardByID(g, gone); !ok {
		t.Error("a Wrath does not reach it (CR 702.26b)")
	}
	if _, ok := battlefieldCardByID(g, stays); ok {
		t.Error("and does reach the creature that is still there")
	}
}

// CR 702.26b's own last sentence: "A permanent that phases out is
// removed from combat. (See rule 506.4.)"
func TestPhasingOutRemovesFromCombat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	bear := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].AttackingTarget = g.Seats[1].ID
			}
		}
		g.noteAttackAnnouncedLocked(bear)
	})

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, bear) })

	out, ok := phasedOutCardByID(g, bear)
	if !ok {
		t.Fatal("it phased out")
	}
	if out.AttackingTarget != uuid.Nil {
		t.Error("it stopped being an attacking creature (CR 506.4)")
	}
	g.ReadSnapshot(func() {
		if g.announcedAttacks[bear] {
			t.Error("and the declaration that named it went with it")
		}
	})
}

// --- the keyword ----------------------------------------------------

// CR 702.26a's static-ability half: a permanent WITH phasing phases
// out during its controller's untap step, and phases back in during
// the next one. Read off the effective characteristics, so a grant
// works the same as a printing.
func TestThePhasingKeywordPhasesOutEveryUntapStep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Tolarian Serpent",
		TypeLine:   "Creature — Serpent",
		Owner:      me,
		Controller: me,
		Keywords:   []string{KeywordPhasing},
	})
	plain := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")

	g.WithWriteLock(func() { g.performPhasingLocked(me) })
	if _, ok := phasedOutCardByID(g, id); !ok {
		t.Fatal("a permanent with phasing phases out during its controller's untap step (CR 702.26a)")
	}
	if _, ok := battlefieldCardByID(g, plain); !ok {
		t.Error("a permanent without it does not")
	}

	// Somebody else's untap step leaves it out.
	g.WithWriteLock(func() { g.performPhasingLocked(g.Seats[1].ID) })
	if _, ok := phasedOutCardByID(g, id); !ok {
		t.Error("and stays out through everyone else's")
	}

	g.WithWriteLock(func() { g.performPhasingLocked(me) })
	if _, ok := battlefieldCardByID(g, id); !ok {
		t.Error("and phases back in during the next one")
	}
}

// The keyword is in the CLOSED table, which is the only thing that
// makes a printed-phasing permanent work with no catalog entry: the
// deck importer filters Scryfall's keywords array through
// CanonicalKeyword before stamping Card.Keywords, so a token that is
// not in the table is dropped on the way in and the permanent phases
// nowhere.
func TestPhasingIsACanonicalKeyword(t *testing.T) {
	got, ok := CanonicalKeyword("Phasing")
	if !ok || got != KeywordPhasing {
		t.Fatalf("CanonicalKeyword(%q) = (%q, %v), want (%q, true)", "Phasing", got, ok, KeywordPhasing)
	}
}

// Both halves of CR 502.1 happen simultaneously, which means the set
// is chosen before either is applied: a permanent with phasing that
// phases IN this step does not phase straight back out.
func TestPhasingInAndOutAreSimultaneous(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Tolarian Serpent",
		TypeLine:   "Creature — Serpent",
		Owner:      me,
		Controller: me,
		Keywords:   []string{KeywordPhasing},
	})

	g.WithWriteLock(func() { g.performPhasingLocked(me) })
	g.WithWriteLock(func() { g.performPhasingLocked(me) })

	if _, ok := battlefieldCardByID(g, id); !ok {
		t.Error("it phased in, and was not chosen to phase out again in the same step (CR 502.1)")
	}
}

// --- "phases out until ~ leaves the battlefield" --------------------

// Oubliette / Out of Time. The permanent is off the untap step's
// clock entirely, and comes back the moment its jailer leaves.
func TestPhaseOutUntilLeavesIgnoresTheUntapStepAndReturnsWhenTheSourceGoes(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	them := g.Seats[1].ID
	jail := pushPhasingTestCard(g, me, "Oubliette", "Enchantment")
	prisoner := pushPhasingTestCard(g, them, "Grizzly Bears", "Creature — Bear")

	g.WithWriteLock(func() {
		_ = g.PhaseOutUntilLeavesForEffect(jail, jail, true, prisoner)
	})
	if _, ok := phasedOutCardByID(g, prisoner); !ok {
		t.Fatal("it phased out")
	}

	// Its controller's untap step comes and goes and it stays out.
	g.WithWriteLock(func() { g.performUntapStepLocked(seatOf(g, them)) })
	if _, ok := phasedOutCardByID(g, prisoner); !ok {
		t.Fatal("an untap step does not release it — the duration is the jailer's presence, not the step")
	}

	g.WithWriteLock(func() {
		g.DestroyPermanentsForEffect([]uuid.UUID{jail})
		g.runStateChecksLocked()
	})

	back, ok := battlefieldCardByID(g, prisoner)
	if !ok {
		t.Fatal("it phases in the moment the source leaves the battlefield")
	}
	if !back.Tapped {
		t.Error("\"Tap that creature as it phases in this way\" — it comes back tapped")
	}
	if back.TapOnPhaseIn {
		t.Error("and the rider is consumed as it fires")
	}
}

// --- persistence ----------------------------------------------------

// An undo across a phase-out is exact, because the state is a Card
// value in a slice rather than a flag several subsystems agree about.
func TestUndoAcrossAPhaseOut(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	bear := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")

	before := g.Clone()

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, bear) })
	if _, ok := phasedOutCardByID(g, bear); !ok {
		t.Fatal("it phased out")
	}

	g.WithWriteLock(func() { g.RestoreFrom(before) })

	if _, ok := battlefieldCardByID(g, bear); !ok {
		t.Error("an undo puts it back on the battlefield")
	}
	g.ReadSnapshot(func() {
		if len(g.PhasedOut.Cards) != 0 {
			t.Errorf("and empties the holding slice: %d left", len(g.PhasedOut.Cards))
		}
	})
}

// The snapshot carries it, with every field of the status: a restore
// that dropped PhasedOutBy would bring a board back under the wrong
// player's untap step.
func TestSnapshotRoundTripsAPhasedOutPermanent(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	bear := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")
	sword := pushPhasingTestCard(g, me, "Test Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
			t.Fatalf("attach: %v", err)
		}
		_ = g.PhaseOutForEffect(uuid.Nil, bear)
	})

	snap := g.CaptureSnapshot()
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}

	host, ok := phasedOutCardByID(restored, bear)
	if !ok {
		t.Fatal("the phased-out permanent survives a restore")
	}
	if host.PhasedOutBy != me {
		t.Errorf("and so does whose untap step brings it back: %v", host.PhasedOutBy)
	}
	att, ok := phasedOutCardByID(restored, sword)
	if !ok || !att.PhasedOutIndirect || !att.IsAttachedTo(bear) {
		t.Error("and so does the attachment's indirect mark and its link")
	}

	restored.WithWriteLock(func() { restored.performPhasingLocked(me) })
	if _, ok := battlefieldCardByID(restored, bear); !ok {
		t.Error("and the restored game phases it in on the right untap step")
	}
}

// CR 702.26k: "Phased-out permanents owned by a player who leaves the
// game also leave the game."
func TestAPhasedOutPermanentLeavesWithItsOwner(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	bear := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, bear) })

	if err := g.Concede(me); err != nil {
		t.Fatalf("concede: %v", err)
	}

	if _, ok := phasedOutCardByID(g, bear); ok {
		t.Error("it left the game with its owner (CR 702.26k)")
	}
}
