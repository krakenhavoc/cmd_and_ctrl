package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// cast_timing_test.go — #1195, ADR 0066's 2026-09-22 amendment. The
// per-PLAYER half of CR 307.1: a grant that opens instant speed for
// everything a player casts, and the restriction that shuts it again.
//
// Every assertion goes through the ONE predicate the cast path, the
// bot enumerator and the view all read — CastTimingOpenLocked — or
// through CastSpell itself, so a test cannot pass against a copy of
// the rule that the other two callers do not share.

// withCatalogCastTimings stubs the per-permanent timing declaration
// for one oracle ID, chaining to whatever the catalog already
// answered so a fixture cannot blank the real one.
func withCatalogCastTimings(t *testing.T, oracle string, timings ...CastTimingRule) {
	t.Helper()
	prev := CatalogCastTimings
	CatalogCastTimings = func(key string) []CastTimingRule {
		if key == oracle {
			return timings
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogCastTimings = prev })
}

// timingSource puts a permanent declaring `oracle` onto the
// battlefield under `controller`.
func timingSource(g *Game, controller *Player, name, oracle string) uuid.UUID {
	c := NewCard(name, controller.ID)
	c.TypeLine = "Artifact"
	c.OracleID = oracle
	c.Controller = controller.ID
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// handCard seeds one card of a given type line into a seat's hand and
// returns the instance ID.
func timingHandCard(p *Player, name, typeLine string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = "{1}"
	p.Hand.PushTop(c)
	return c.InstanceID
}

// timingOpenFor asks the one predicate about a card sitting in a
// seat's hand.
func timingOpenFor(t *testing.T, g *Game, p *Player, id uuid.UUID) bool {
	t.Helper()
	var out bool
	g.ReadSnapshot(func() {
		c, ok := g.cardInZoneLocked(p.Hand, id)
		if !ok {
			t.Fatalf("card %s is not in %s's hand", id, p.Name)
		}
		out = g.CastTimingOpenLocked(p.ID, c, ZoneHand, nil)
	})
	return out
}

// storedTimings counts the cast-timing statements stored on a player.
// Since #1195 folded onto #1197's registry they are PlayerStatics
// carrying a rule and no Keyword, which is also what keeps them out of
// playerAbilityTokensLocked — see TestAStoredTimingGrantIsNotAnAbilityToken.
func storedTimings(p *Player) int {
	n := 0
	for _, st := range p.Statics {
		if st.Timing.Timing != TimingNormal {
			n++
		}
	}
	return n
}

// --- the grant ------------------------------------------------------

// Vedalken Orrery: "You may cast spells as though they had flash." A
// CREATURE spell at instant speed, driven all the way through
// CastSpell so the assertion is about the announce path and not only
// about the predicate.
func TestOrreryCastsACreatureAtInstantSpeed(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-timing-orrery"
	advanceTo(t, g, StepUpkeep)

	bear := timingHandCard(me, "Test Bear", "Creature — Bear")
	if timingOpenFor(t, g, me, bear) {
		t.Fatal("setup: a creature spell is castable in an upkeep step with nothing granting anything")
	}

	withCatalogCastTimings(t, oracle, CastTimingRule{
		Timing: TimingFlash,
		Label:  "You may cast spells as though they had flash.",
	})
	timingSource(g, me, "Test Orrery", oracle)

	if !timingOpenFor(t, g, me, bear) {
		t.Fatal("the Orrery did not open the window")
	}
	if err := g.CastSpell(me.ID, bear, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell under an Orrery: %v", err)
	}
	if me.Hand.Contains(bear) {
		t.Error("the creature spell never left the hand")
	}
}

// Leyline of Anticipation on an OPPONENT's turn, with a sorcery. The
// other half of the sorcery-speed gate: not just "the stack is empty
// and it is a main phase" but "it is not even your turn".
func TestLeylineCastsASorceryOnAnOpponentsTurn(t *testing.T) {
	g := newActiveGame(t)
	me, active := g.Seats[1], g.Seats[0]
	const oracle = "test-timing-leyline"
	advanceTo(t, g, StepPrecombatMain)
	if g.Seats[g.Turn.ActiveSeat].ID != active.ID {
		t.Fatalf("setup: active seat is %d, want seat 0", g.Turn.ActiveSeat)
	}

	bolt := timingHandCard(me, "Test Ritual", "Sorcery")
	if timingOpenFor(t, g, me, bolt) {
		t.Fatal("setup: a sorcery is castable on somebody else's turn with nothing granting anything")
	}

	withCatalogCastTimings(t, oracle, CastTimingRule{
		Timing: TimingFlash,
		Label:  "You may cast spells as though they had flash.",
	})
	timingSource(g, me, "Test Leyline", oracle)

	if !timingOpenFor(t, g, me, bolt) {
		t.Fatal("the Leyline did not open an opponent's turn")
	}
	if err := g.CastSpell(me.ID, bolt, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell on an opponent's turn under a Leyline: %v", err)
	}
}

// A grant reaches only the SEAT its Affects clause names. "You may
// cast spells as though they had flash" is the source controller's
// permission and nobody else's.
func TestAFlashGrantReachesOnlyItsController(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	const oracle = "test-timing-mine-only"
	advanceTo(t, g, StepUpkeep)
	withCatalogCastTimings(t, oracle, CastTimingRule{Timing: TimingFlash, Label: "You may cast spells as though they had flash."})
	timingSource(g, me, "Test Orrery", oracle)

	mine := timingHandCard(me, "My Bear", "Creature — Bear")
	theirs := timingHandCard(them, "Their Bear", "Creature — Bear")
	if !timingOpenFor(t, g, me, mine) {
		t.Error("the controller of the grant cannot use it")
	}
	if timingOpenFor(t, g, them, theirs) {
		t.Error("a \"you may cast\" grant reached an opponent")
	}
}

// Yeva, Nature's Herald: "You may cast CREATURE spells as though they
// had flash." The filter is the whole of the difference, and the
// sorcery beside the creature is what proves it narrows.
func TestACreatureOnlyGrantDoesNotOpenASorcery(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-timing-yeva"
	advanceTo(t, g, StepUpkeep)
	withCatalogCastTimings(t, oracle, CastTimingRule{
		Timing: TimingFlash,
		Filter: PermissionFilter{CreatureOnly: true},
		Label:  "You may cast creature spells as though they had flash.",
	})
	timingSource(g, me, "Test Yeva", oracle)

	bear := timingHandCard(me, "Test Bear", "Creature — Bear")
	ritual := timingHandCard(me, "Test Ritual", "Sorcery")
	if !timingOpenFor(t, g, me, bear) {
		t.Error("a creature-only grant did not open a creature spell")
	}
	if timingOpenFor(t, g, me, ritual) {
		t.Error("a creature-only grant opened a sorcery")
	}
	if err := g.CastSpell(me.ID, ritual, CastSpellParams{}); !errors.Is(err, ErrSorcerySpeedRequired) {
		t.Errorf("CastSpell of the sorcery = %v, want ErrSorcerySpeedRequired", err)
	}
}

// A permanent that has lost its abilities (CR 613.1f) stops saying
// it. The derivation reads CatalogAbilityKey for exactly this reason,
// and an Orrery under a Merfolk Trickster is the shape that proves it.
func TestAGrantDiesWithItsSourcesAbilities(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-timing-stripped-orrery"
	advanceTo(t, g, StepUpkeep)
	withCatalogCastTimings(t, oracle, CastTimingRule{Timing: TimingFlash, Label: "You may cast spells as though they had flash."})
	src := timingSource(g, me, "Test Orrery", oracle)

	bear := timingHandCard(me, "Test Bear", "Creature — Bear")
	if !timingOpenFor(t, g, me, bear) {
		t.Fatal("setup: the Orrery is not granting")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != src {
				continue
			}
			eff := c.printedCharacteristic()
			eff.AbilitiesRemoved = true
			c.effective = &eff
		}
	})
	if timingOpenFor(t, g, me, bear) {
		t.Error("an Orrery with no abilities is still granting flash")
	}
}

// --- the inverse ----------------------------------------------------

// Teferi, Time Raveler: "Each opponent can cast spells only any time
// they could cast a sorcery." An opponent's INSTANT on your turn is
// the case the clause exists for.
func TestTeferiStopsAnOpponentsInstantOnYourTurn(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	const oracle = "test-timing-teferi"
	advanceTo(t, g, StepPrecombatMain)
	withCatalogCastTimings(t, oracle, CastTimingRule{
		Timing:  TimingSorcery,
		Affects: TimingAffectsEachOpponent,
		Label:   "Each opponent can cast spells only any time they could cast a sorcery.",
	})

	bolt := timingHandCard(them, "Their Bolt", "Instant")
	mine := timingHandCard(me, "My Bolt", "Instant")
	if !timingOpenFor(t, g, them, bolt) {
		t.Fatal("setup: an instant is not castable on an opponent's main phase")
	}
	timingSource(g, me, "Test Teferi", oracle)

	if timingOpenFor(t, g, them, bolt) {
		t.Error("Teferi did not stop an opponent's instant on your turn")
	}
	if err := g.CastSpell(them.ID, bolt, CastSpellParams{}); !errors.Is(err, ErrSorcerySpeedRequired) {
		t.Errorf("CastSpell of the opponent's instant = %v, want ErrSorcerySpeedRequired", err)
	}
	// And the clause says "each OPPONENT": Teferi's own controller
	// keeps instant speed.
	if !timingOpenFor(t, g, me, mine) {
		t.Error("Teferi restricted his own controller")
	}
}

// CR 101.2, and the one ordering decision in the whole predicate: a
// restriction beats a grant, whoever controls which. An opponent's
// Vedalken Orrery does not get them past your Teferi.
func TestARestrictionBeatsAGrantTheOpponentControls(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	const teferi, orrery = "test-timing-teferi-beats", "test-timing-orrery-beaten"
	advanceTo(t, g, StepPrecombatMain)

	prev := CatalogCastTimings
	CatalogCastTimings = func(key string) []CastTimingRule {
		switch key {
		case teferi:
			return []CastTimingRule{{
				Timing: TimingSorcery, Affects: TimingAffectsEachOpponent,
				Label: "Each opponent can cast spells only any time they could cast a sorcery.",
			}}
		case orrery:
			return []CastTimingRule{{Timing: TimingFlash, Label: "You may cast spells as though they had flash."}}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogCastTimings = prev })

	timingSource(g, them, "Their Orrery", orrery)
	bolt := timingHandCard(them, "Their Bolt", "Instant")
	if !timingOpenFor(t, g, them, bolt) {
		t.Fatal("setup: the opponent's own Orrery does not open their instant")
	}
	timingSource(g, me, "My Teferi", teferi)
	if timingOpenFor(t, g, them, bolt) {
		t.Error("a grant beat a restriction — CR 101.2 says \"can't\" wins")
	}
}

// Dosan the Falling Leaf: "Players can cast spells only during their
// own turns." NOT the sorcery-speed gate — it leaves every
// instant-speed window on your OWN turn open and shuts every window
// on anybody else's, including the one an Orrery would open.
func TestYourTurnOnlyIsNotTheSorceryGate(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	const oracle = "test-timing-dosan"
	withCatalogCastTimings(t, oracle, CastTimingRule{
		Timing:  TimingYourTurnOnly,
		Affects: TimingAffectsEachPlayer,
		Label:   "Players can cast spells only during their own turns.",
	})
	timingSource(g, me, "Test Dosan", oracle)
	advanceTo(t, g, StepUpkeep)

	mine := timingHandCard(me, "My Bolt", "Instant")
	theirs := timingHandCard(them, "Their Bolt", "Instant")
	// The active seat keeps its instant speed in a step the
	// sorcery-speed gate is shut in — which is the whole distinction.
	if !timingOpenFor(t, g, me, mine) {
		t.Error("Dosan shut the active player's own upkeep")
	}
	if timingOpenFor(t, g, them, theirs) {
		t.Error("Dosan let a non-active player cast an instant")
	}
}

// A land PLAY is a special action (CR 116.2a), not a cast, so no
// timing statement reaches it in either direction: an Orrery does not
// open a land drop outside a main phase, and a Dosan does not close
// the active player's.
func TestTimingStatementsDoNotReachALandPlay(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-timing-orrery-land"
	advanceTo(t, g, StepUpkeep)
	withCatalogCastTimings(t, oracle, CastTimingRule{Timing: TimingFlash, Label: "You may cast spells as though they had flash."})
	timingSource(g, me, "Test Orrery", oracle)

	land := NewCard("Test Wastes", me.ID)
	land.TypeLine = "Land"
	me.Hand.PushTop(land)
	if err := g.CastSpell(me.ID, land.InstanceID, CastSpellParams{}); !errors.Is(err, ErrSorcerySpeedRequired) {
		t.Errorf("a land play in an upkeep under an Orrery = %v, want ErrSorcerySpeedRequired", err)
	}
}

// --- the stored half, and its durations -----------------------------

// Emergence Zone: "You may cast spells this turn as though they had
// flash." The statement outlives the land that sacrificed itself to
// make it, and it ends at that turn's cleanup step (CR 514.2) —
// through the real rotation, so the sweep and the query agree.
func TestAStoredFlashGrantEndsAtEndOfTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, StepPrecombatMain)
	g.WithWriteLock(func() {
		g.GrantCastTimingForEffect(me.ID, CastTimingRule{Timing: TimingFlash},
			"You may cast spells this turn as though they had flash.", uuid.Nil, Duration{})
	})
	bear := timingHandCard(me, "Test Bear", "Creature — Bear")
	advanceTo(t, g, StepEnd)
	if !timingOpenFor(t, g, me, bear) {
		t.Fatal("the grant died before the turn did")
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if timingOpenFor(t, g, me, bear) {
		t.Error("an until-end-of-turn grant outlived its turn")
	}
	if n := storedTimings(me); n != 0 {
		t.Errorf("the cleanup sweep left %d statements behind", n)
	}
}

// Teferi, Time Raveler's +1: "Until your next turn, you may cast
// SORCERY spells as though they had flash." Four seats, so the window
// has to ride out three opponents' turns and then end as the holder's
// own turn BEGINS (CR 500.1) rather than at its cleanup.
func TestAStoredGrantUntilYourNextTurnEndsAsThatTurnBegins(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		g.GrantCastTimingForEffect(me.ID, CastTimingRule{
			Timing: TimingFlash,
			Filter: PermissionFilter{SorceryOnly: true},
		}, "Until your next turn, you may cast sorcery spells as though they had flash.",
			uuid.Nil, g.UntilYourNextTurnDuration(me.ID))
	})
	ritual := timingHandCard(me, "Test Ritual", "Sorcery")
	bear := timingHandCard(me, "Test Bear", "Creature — Bear")

	for i, who := range []string{"my own turn", "opponent 1", "opponent 2", "opponent 3"} {
		if !timingOpenFor(t, g, me, ritual) {
			t.Fatalf("the sorcery window died during %s", who)
		}
		// The filter still narrows for the whole of the window.
		if timingOpenFor(t, g, me, bear) && g.Seats[g.Turn.ActiveSeat].ID != me.ID {
			t.Fatalf("the sorcery-only window opened a creature spell during %s", who)
		}
		if err := g.PassTurn(); err != nil {
			t.Fatalf("PassTurn %d: %v", i, err)
		}
	}
	if g.Seats[g.Turn.ActiveSeat].ID != me.ID {
		t.Fatalf("setup: expected to be back on my own turn, active seat is %d", g.Turn.ActiveSeat)
	}
	if n := storedTimings(me); n != 0 {
		t.Errorf("an until-your-next-turn statement survived the turn that ends it: %d left", n)
	}
	advanceTo(t, g, StepEnd)
	if timingOpenFor(t, g, me, ritual) {
		t.Error("the window outlived the beginning of your next turn")
	}
}

// A stored grant with no stated window is stamped "until end of turn"
// on the way in rather than lasting forever — the same posture
// GrantCastPermissionForEffect takes, so a card file that forgets the
// duration gets the narrowest real one.
func TestAStoredGrantWithNoDurationIsUntilEndOfTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.GrantCastTimingForEffect(me.ID, CastTimingRule{Timing: TimingFlash}, "", uuid.Nil, Duration{})
	})
	if storedTimings(me) != 1 {
		t.Fatalf("the grant did not land: %+v", me.Statics)
	}
	if k := me.Statics[0].Duration.Kind; k != UntilEndOfTurn {
		t.Errorf("unstamped duration kind = %v, want UntilEndOfTurn", k)
	}
}

// --- the permission's own timing still wins its own branch ----------

// ADR 0066 Decision 6 is untouched: a permission that carries
// TimingFlash (madness, suspend's free cast) opens the window with
// nothing on the battlefield saying anything, and one that carries
// TimingSorcery shuts it for an instant.
func TestThePermissionsOwnTimingStillReadsThroughTheOnePredicate(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	ritual := timingHandCard(me, "Test Ritual", "Sorcery")
	bolt := timingHandCard(me, "Test Bolt", "Instant")

	g.mu.RLock()
	defer g.mu.RUnlock()
	sorcery, _ := g.cardInZoneLocked(me.Hand, ritual)
	instant, _ := g.cardInZoneLocked(me.Hand, bolt)

	if g.CastTimingOpenLocked(me.ID, sorcery, ZoneHand, nil) {
		t.Error("a sorcery is open in an upkeep with no permission")
	}
	if !g.CastTimingOpenLocked(me.ID, sorcery, ZoneHand, &CastPermission{Player: me.ID, Timing: TimingFlash}) {
		t.Error("TimingFlash did not open the sorcery")
	}
	if !g.CastTimingOpenLocked(me.ID, instant, ZoneHand, nil) {
		t.Error("an instant is shut in an upkeep")
	}
	if g.CastTimingOpenLocked(me.ID, instant, ZoneHand, &CastPermission{Player: me.ID, Timing: TimingSorcery}) {
		t.Error("TimingSorcery did not shut the instant")
	}
}

// A statement that names a zone is about casts out of THAT zone.
// Nothing on the seam row prints one, and the field exists so that
// "you may cast spells from your graveyard as though they had flash"
// is not a second read when it arrives.
func TestAZoneScopedGrantStaysInItsZone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	g.WithWriteLock(func() {
		g.GrantCastTimingForEffect(me.ID, CastTimingRule{Timing: TimingFlash, FromZone: ZoneGraveyard},
			"You may cast spells from your graveyard as though they had flash.", uuid.Nil, Duration{})
	})
	ritual := timingHandCard(me, "Test Ritual", "Sorcery")
	g.mu.RLock()
	defer g.mu.RUnlock()
	c, _ := g.cardInZoneLocked(me.Hand, ritual)
	if g.CastTimingOpenLocked(me.ID, c, ZoneHand, nil) {
		t.Error("a graveyard-scoped grant opened a cast from hand")
	}
	if !g.CastTimingOpenLocked(me.ID, c, ZoneGraveyard, nil) {
		t.Error("a graveyard-scoped grant did not open a cast from the graveyard")
	}
}

// THE FOLD, asserted: #1195's stored statements ride #1197's one
// player-level registry, and the two kinds of entry do not see each
// other. A timing grant is not an ability token — it must not reach
// playerAbilityTokensLocked, where it would be offered to the
// protection parser as an empty keyword — and an ability grant must
// not reach the timing read.
func TestAStoredTimingGrantIsNotAnAbilityToken(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	ritual := timingHandCard(me, "Test Ritual", "Sorcery")

	g.WithWriteLock(func() {
		g.GrantCastTimingForEffect(me.ID, CastTimingRule{Timing: TimingFlash},
			"You may cast spells this turn as though they had flash.", uuid.Nil, Duration{})
		g.GrantPlayerStaticForEffect(me.ID, KeywordHexproof, "Test hexproof", uuid.Nil, Duration{})
	})
	if n := len(me.Statics); n != 2 {
		t.Fatalf("one registry, two entries: got %d", n)
	}

	// The ability reader sees the hexproof and nothing else.
	g.mu.RLock()
	toks := g.PlayerAbilitiesForEffect(me)
	hasHexproof := g.PlayerHasKeywordLocked(me, KeywordHexproof)
	g.mu.RUnlock()
	if len(toks) != 1 || toks[0] != KeywordHexproof {
		t.Errorf("player ability tokens = %v, want just [hexproof]", toks)
	}
	if !hasHexproof {
		t.Error("the hexproof grant was lost")
	}

	// The timing read sees the flash grant and is not confused by the
	// hexproof entry sitting beside it.
	if !timingOpenFor(t, g, me, ritual) {
		t.Error("the timing grant did not open the window from the shared registry")
	}
}
