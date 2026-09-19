package game

import (
	"testing"

	"github.com/google/uuid"
)

// protection_player_test.go — #980, CR 702.16k: the one quality in the
// closed grammar that is not a characteristic of the source at all.
// "Protection from the chosen player" tests who CONTROLS the source,
// against a seat the protected permanent stored as it entered
// (CR 614.12, Card.ChosenPlayer).
//
// Every test here is the same assertion twice: the chosen player is
// refused, and SOMEBODY ELSE IS NOT. A player quality that degenerated
// into "refuse everything" would pass half of each of these, which is
// why no test below checks only the refusal.

// pushProtectedFromPlayer seeds a creature with CR 702.16k's printed
// ability and the seat it has already chosen. Stamped rather than
// prompted: the prompt has its own tests, and these are about the
// rules that read the answer.
func pushProtectedFromPlayer(g *Game, owner *Player, name string, chosen uuid.UUID) uuid.UUID {
	id := pushColouredCreature(g, owner, name, []string{"U"}, ProtectionFromChosenPlayer)
	g.WithWriteLock(func() {
		g.Battlefield.Cards[findCardOnBattlefield(g, id)].ChosenPlayer = chosen
		g.recomputeLayersLocked()
	})
	return id
}

// --- the reader ----------------------------------------------------

// TestTheChosenPlayerQualityIsResolvedByTheReader — the token names no
// seat, so a quality parsed on its own protects from nobody and one
// read off a card knows who it means. That split is what keeps a raw
// UUID out of every display string.
func TestTheChosenPlayerQualityIsResolvedByTheReader(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushProtectedFromPlayer(g, opp, "True-Name Nemesis", me.ID)

	bare, ok := ParseProtectionQuality(ProtectionFromChosenPlayer)
	if !ok {
		t.Fatal("the grammar accepts the token")
	}
	if bare.Player != uuid.Nil {
		t.Errorf("the parser resolves no seat: %v", bare.Player)
	}
	if bare.Matches(&Characteristic{Controller: me.ID}) {
		t.Error("an unbound player quality matches nothing — the seat is not in the token")
	}

	card, _ := battlefieldCardByID(g, id)
	qs := ProtectionQualities(&card)
	if len(qs) != 1 {
		t.Fatalf("one quality, got %+v", qs)
	}
	if qs[0].Kind != ProtectionQualityPlayer {
		t.Errorf("kind %v, want player", qs[0].Kind)
	}
	if qs[0].Player != me.ID {
		t.Errorf("the reader resolved %v, want the stored seat %v", qs[0].Player, me.ID)
	}
	if qs[0].Printed != "the chosen player" {
		t.Errorf("printed %q — the display string never holds a UUID", qs[0].Printed)
	}
	if !qs[0].Matches(&Characteristic{Controller: me.ID}) {
		t.Error("a source controlled by the chosen player has the quality")
	}
	if qs[0].Matches(&Characteristic{Controller: opp.ID}) {
		t.Error("a source controlled by anybody else does not")
	}
}

// TestAnUnansweredNemesisIsProtectedFromNobody — the window between the
// permanent entering and its controller answering. An unchosen player
// is nobody, not everybody; the whole engine errs that way and this is
// the one place where the other direction would make a creature
// untouchable by the entire table.
func TestAnUnansweredNemesisIsProtectedFromNobody(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushColouredCreature(g, opp, "Unanswered Nemesis", []string{"U"}, ProtectionFromChosenPlayer)
	card, _ := battlefieldCardByID(g, id)

	if !HasProtection(&card) {
		t.Fatal("the printed ability is there either way")
	}
	for _, who := range []uuid.UUID{me.ID, opp.ID, uuid.New()} {
		if ProtectedFrom(&card, &Characteristic{Controller: who}) {
			t.Errorf("nobody has been chosen, so %v is not the chosen player", who)
		}
	}
}

// --- T: CR 702.16b -------------------------------------------------

// TestTheChosenPlayerCannotTargetIt, and the half that matters: the
// quality reads the CONTROLLER of the source, so the colour of the
// spell is irrelevant and another seat's identical spell is legal.
func TestTheChosenPlayerCannotTargetIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, owner, other := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	const oracle = "test-protection-chosen-player"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return anyCreatureSpec()
		}
		return nil
	})

	nemesis := pushProtectedFromPlayer(g, owner, "True-Name Nemesis", me.ID)

	// Mine is refused whatever colour it is: the quality is a player.
	if err := castColouredSpellAt(t, g, me, oracle, []string{"R"},
		TargetRef{Kind: TargetCard, ID: nemesis}); err != ErrIllegalTarget {
		t.Fatalf("the chosen player's red spell: got %v, want ErrIllegalTarget", err)
	}
	if err := castColouredSpellAt(t, g, me, oracle, []string{"W"},
		TargetRef{Kind: TargetCard, ID: nemesis}); err != ErrIllegalTarget {
		t.Fatalf("the chosen player's white spell: got %v, want ErrIllegalTarget", err)
	}
	// Somebody else's identical spell is fine. Without this the test
	// passes for a check that refuses every source.
	if err := castColouredSpellAt(t, g, other, oracle, []string{"R"},
		TargetRef{Kind: TargetCard, ID: nemesis}); err != nil {
		t.Fatalf("another seat's red spell must be legal: %v", err)
	}
}

// TestTheQualityReadsTheAbilitySourcesController — an ABILITY is tested
// by who controls the permanent it came from, not by who activated it
// and not by any characteristic. Same shape as #662's colour version,
// pointed at the player quality.
func TestTheQualityReadsTheAbilitySourcesController(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, owner, other := g.Seats[0], g.Seats[1], g.Seats[2]

	nemesis := pushProtectedFromPlayer(g, owner, "True-Name Nemesis", me.ID)

	pinger := func(name string, controller uuid.UUID) *Card {
		c := NewCard(name, controller)
		c.TypeLine = "Artifact"
		c.Controller = controller
		g.Battlefield.PushTop(c)
		return &g.Battlefield.Cards[findCardOnBattlefield(g, c.InstanceID)]
	}
	mine := pinger("Chosen Player's Pinger", me.ID)
	theirs := pinger("Another Seat's Pinger", other.ID)

	var mineOK, theirsOK bool
	g.ReadSnapshot(func() {
		spec := anyCreatureSpec()
		ref := TargetRef{Kind: TargetCard, ID: nemesis}
		mineOK = g.targetLegalLocked(SourceObject(me.ID, mine), spec, ref)
		theirsOK = g.targetLegalLocked(SourceObject(other.ID, theirs), spec, ref)
	})
	if mineOK {
		t.Error("an ability from a permanent the chosen player controls may not target it")
	}
	if !theirsOK {
		t.Error("an ability from anybody else's permanent may")
	}
}

// --- D: CR 702.16e -------------------------------------------------

// TestDamageFromTheChosenPlayerIsPrevented — through the LKI on the
// event, which already carries the source's controller (ADR 0072 §3
// put it there for exactly this).
func TestDamageFromTheChosenPlayerIsPrevented(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, owner, other := g.Seats[0], g.Seats[1], g.Seats[2]

	nemesis := pushProtectedFromPlayer(g, owner, "True-Name Nemesis", me.ID)
	mineBolt := pushColouredCreature(g, me, "Chosen Player's Bear", []string{"R"})
	theirBolt := pushColouredCreature(g, other, "Another Seat's Bear", []string{"R"})

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(mineBolt, nemesis, 3); err != nil {
			t.Fatalf("damage from the chosen player: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(theirBolt, nemesis, 1); err != nil {
			t.Fatalf("damage from another seat: %v", err)
		}
	})
	if got := damageOn(g, nemesis); got != 1 {
		t.Errorf("marked %d damage, want 1 — the chosen player's 3 prevented, the other seat's 1 dealt", got)
	}
}

// --- B: CR 702.16f -------------------------------------------------

// TestTheChosenPlayersCreatureCannotBlockIt, and somebody else's can.
func TestTheChosenPlayersCreatureCannotBlockIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, owner, other := g.Seats[0], g.Seats[1], g.Seats[2]

	nemesis := pushProtectedFromPlayer(g, owner, "True-Name Nemesis", me.ID)
	mineBlocker := pushColouredCreature(g, me, "Chosen Player's Blocker", []string{"R"})
	theirBlocker := pushColouredCreature(g, other, "Another Seat's Blocker", []string{"R"})

	card := func(id uuid.UUID) *Card {
		return &g.Battlefield.Cards[findCardOnBattlefield(g, id)]
	}
	var refusal, allowed BlockRefusal
	g.ReadSnapshot(func() {
		refusal = g.BlockPairRefusalLocked(card(nemesis), card(mineBlocker))
		allowed = g.BlockPairRefusalLocked(card(nemesis), card(theirBlocker))
	})
	if refusal.Reason != BlockReasonProtection {
		t.Errorf("the chosen player's blocker: got %q, want protection", refusal.Reason)
	}
	if !allowed.Legal() {
		t.Errorf("another seat's blocker is legal: got %q", allowed.Reason)
	}
}

// --- E: CR 702.16c-d -----------------------------------------------

// TestAnAuraTheChosenPlayerControlsFallsOff, and one somebody else
// controls stays on. The test is against the ATTACHMENT's controller
// because that is what the player quality compares.
func TestAnAuraTheChosenPlayerControlsFallsOff(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, owner, other := g.Seats[0], g.Seats[1], g.Seats[2]

	nemesis := pushProtectedFromPlayer(g, owner, "True-Name Nemesis", me.ID)
	mineAura := pushAttachTestCard(g, me.ID, "Chosen Player's Aura", "Enchantment — Aura")
	theirAura := pushAttachTestCard(g, other.ID, "Another Seat's Aura", "Enchantment — Aura")

	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{mineAura, theirAura} {
			if err := g.AttachForEffect(id, TargetRef{Kind: TargetCard, ID: nemesis}); err != nil {
				t.Fatalf("attach %v: %v", id, err)
			}
		}
		g.runStateChecksLocked()
	})

	if _, ok := battlefieldCardByID(g, mineAura); ok {
		t.Error("CR 704.5m: an Aura the chosen player controls goes to its owner's graveyard")
	}
	if c, ok := battlefieldCardByID(g, theirAura); !ok || !c.IsAttachedTo(nemesis) {
		t.Error("an Aura anybody else controls is unaffected")
	}
}

// --- the field's lifecycle -----------------------------------------

// TestTheChosenPlayerIsClearedOnTheWayOut — CR 400.7. A bounced
// Nemesis is a new object when it comes back and chooses again; one in
// a graveyard is protected from nobody, which is also what stops the
// reader having to ask what zone it is in.
func TestTheChosenPlayerIsClearedOnTheWayOut(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushProtectedFromPlayer(g, opp, "True-Name Nemesis", me.ID)

	if got := g.ChosenPlayerOf(id); got != me.ID {
		t.Fatalf("stored %v, want %v", got, me.ID)
	}
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	for _, c := range opp.Hand.Cards {
		if c.Name == "True-Name Nemesis" && c.ChosenPlayer != uuid.Nil {
			t.Errorf("the card in hand still names %v (CR 400.7)", c.ChosenPlayer)
		}
	}
	if got := g.ChosenPlayerOf(id); got != uuid.Nil {
		t.Errorf("ChosenPlayerOf on a permanent that has left: %v, want nil", got)
	}
}

// TestTheChosenPlayerIsClearedByTheNewObjectReset — the other clearing
// site. An entry that mints a new object (CR 400.7) wipes it alongside
// NamedTribe and ChosenColor, so a blinked Nemesis chooses again even
// though it never passed through zone.go's exit.
func TestTheChosenPlayerIsClearedByTheNewObjectReset(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushProtectedFromPlayer(g, opp, "True-Name Nemesis", me.ID)

	var newID uuid.UUID
	g.WithWriteLock(func() { newID = g.resetAsNewObjectLocked(id) })
	if newID == uuid.Nil {
		t.Fatal("the reset ran")
	}
	if got := g.ChosenPlayerOf(newID); got != uuid.Nil {
		t.Errorf("the new object still names %v", got)
	}
}

// TestTheChosenPlayerIsNotACopiableValue — CR 707.2. It falls out of
// where the field lives rather than out of a rule anybody has to
// remember, and this test is what keeps it that way: a Clone of a
// Nemesis copies the printed protection ability and chooses its OWN
// player as it enters.
func TestTheChosenPlayerIsNotACopiableValue(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushProtectedFromPlayer(g, opp, "True-Name Nemesis", me.ID)
	card, _ := battlefieldCardByID(g, id)

	values := CopiableValuesOf(card)
	found := false
	for _, kw := range values.Keywords {
		if kw == ProtectionFromChosenPlayer {
			found = true
		}
	}
	if !found {
		t.Error("the printed ability IS copiable (CR 707.2 copies printed values)")
	}
	// And the answer is not. PrintedValues has no field for it, which
	// is the assertion — a copy has nowhere to put the chosen player
	// and so does not get one.
	blank := Card{InstanceID: uuid.New(), Owner: opp.ID, Controller: opp.ID}
	if blank.ChosenPlayer != uuid.Nil {
		t.Fatal("a fresh card names nobody")
	}
	if ProtectedFrom(&blank, &Characteristic{Controller: me.ID}) {
		t.Error("a copy with no chosen player of its own is protected from nobody")
	}
}

// TestTheStoredAnswersSurviveASnapshot is the OTHER carry, and it has
// to assert on the restored GAME rather than on a re-captured snapshot.
// TestSnapshotRoundTripIsExact compares capture → JSON → restore →
// capture, which is structurally blind to a field missing from both
// projections symmetrically: absent equals absent. Reading the restored
// card is what closes that.
//
// All three CR 614.12-family answers, because none of them was covered
// before #980 and they fail the same way — a player made the choice and
// nothing in the catalog can re-derive it.
//
// #1005 generalised the shape: snapshot_carried_test.go now runs this
// assertion for EVERY field the drift plan calls `carried`, so a fourth
// stored answer is covered the moment its row lands. This one stays as
// the named, readable case for the three the report was filed about.
func TestTheStoredAnswersSurviveASnapshot(t *testing.T) {
	g := newRestorableGame(t)
	enrich(t, g)

	want := map[uuid.UUID]Card{}
	for _, c := range g.Battlefield.Cards {
		want[c.InstanceID] = c
	}
	_, restored := roundTrip(t, g)

	seen := 0
	for _, got := range restored.Battlefield.Cards {
		orig, ok := want[got.InstanceID]
		if !ok || orig.ChosenPlayer == uuid.Nil {
			continue
		}
		seen++
		if got.ChosenPlayer != orig.ChosenPlayer {
			t.Errorf("ChosenPlayer = %v, want %v — a restored Nemesis is protected from nobody",
				got.ChosenPlayer, orig.ChosenPlayer)
		}
		if got.ChosenColor != orig.ChosenColor {
			t.Errorf("ChosenColor = %q, want %q", got.ChosenColor, orig.ChosenColor)
		}
		if got.NamedTribe != orig.NamedTribe {
			t.Errorf("NamedTribe = %q, want %q", got.NamedTribe, orig.NamedTribe)
		}
	}
	if seen == 0 {
		t.Fatal("the fixture carries no stored answers; this test measured nothing")
	}
}

// TestTheChosenPlayerSurvivesAnUndo — carried by clone and the
// snapshot, classified `carried` in snapshot_drift_test.go. A restore
// that lost it would bring the Nemesis back protected from nobody,
// silently.
func TestTheChosenPlayerSurvivesAnUndo(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushProtectedFromPlayer(g, opp, "True-Name Nemesis", me.ID)

	snap := g.Clone()
	g.WithWriteLock(func() {
		g.Battlefield.Cards[findCardOnBattlefield(g, id)].ChosenPlayer = uuid.Nil
	})
	g.RestoreFrom(snap)
	if got := g.ChosenPlayerOf(id); got != me.ID {
		t.Errorf("after the restore the Nemesis names %v, want %v", got, me.ID)
	}
}
