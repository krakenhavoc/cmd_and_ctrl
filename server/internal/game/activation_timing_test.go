package game

import (
	"testing"

	"github.com/google/uuid"
)

// activation_timing_test.go — #1208, the ACTIVATION twin of
// cast_timing_test.go. CR 602.5d and CR 606.3 asked once, through the
// one predicate the activation path, the bot enumerator and the view
// all read (ActivationTimingOpenLocked), so a test cannot pass
// against a copy of the rule the other callers do not share.

// withCatalogActivationTimings stubs the per-permanent declaration
// for one oracle ID, chaining to whatever the catalog already
// answered so a fixture cannot blank the real one.
func withCatalogActivationTimings(t *testing.T, oracle string, timings ...ActivationTiming) {
	t.Helper()
	prev := CatalogActivationTimings
	CatalogActivationTimings = func(key string) []ActivationTiming {
		if key == oracle {
			return timings
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogActivationTimings = prev })
}

// activationTimingSource puts a permanent declaring `oracle` onto the
// battlefield under `controller` and returns it.
func activationTimingSource(g *Game, controller *Player, name, oracle string) Card {
	c := NewCard(name, controller.ID)
	c.TypeLine = "Artifact"
	c.OracleID = oracle
	c.Controller = controller.ID
	g.Battlefield.PushTop(c)
	return c
}

// activationOpen asks the one predicate about an ability of `card`.
func activationOpen(t *testing.T, g *Game, p *Player, card Card, ab ActivationAbility) bool {
	t.Helper()
	var out bool
	g.ReadSnapshot(func() {
		out = g.ActivationTimingOpenLocked(p.ID, card, ZoneBattlefield, ab)
	})
	return out
}

// alwaysCovers is the "this statement is about every activation"
// predicate, for the tests that are about the FOLD rather than about
// a card's narrowing.
func alwaysCovers(ActivationQuery) bool { return true }

// --- the ability's own timing ---------------------------------------

// Step 2 of the fold, with nothing declared: an ordinary activated
// ability is instant-speed (CR 117.1b) and a sorcery-speed one is
// not (CR 602.5d).
func TestActivationTimingWithoutStatements(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	src := activationTimingSource(g, me, "Test Source", "test-activation-timing-none")

	if !activationOpen(t, g, me, src, ActivationAbility{Label: "{T}: Add {C}."}) {
		t.Error("a plain activated ability is instant-speed (CR 117.1b)")
	}
	if activationOpen(t, g, me, src, ActivationAbility{Label: "Equip {2}", SorcerySpeed: true}) {
		t.Error("a sorcery-speed ability is not activatable in an upkeep step")
	}
	if activationOpen(t, g, me, src, ActivationAbility{Label: "+1: …", Loyalty: true}) {
		t.Error("CR 606.3: a loyalty ability is not activatable in an upkeep step")
	}
}

// CR 605.3a: a mana ability has its own window and this read has
// nothing to say about one — not even when a statement claims to
// cover everything. The fast POSITIVE, taken before the walk.
func TestActivationTimingNeverTouchesManaAbilities(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	const oracle = "test-activation-timing-mana"

	withCatalogActivationTimings(t, oracle, ActivationTiming{
		Label:  "Test — activations are sorcery-speed.",
		Timing: TimingSorcery,
		Covers: alwaysCovers,
	})
	src := activationTimingSource(g, me, "Test Lock", oracle)

	if !activationOpen(t, g, me, src, ActivationAbility{Label: "{T}: Add {G}.", Mana: true}) {
		t.Error("a mana ability must stay open: CR 605.3a is not the CR 602.5d window")
	}
	if activationOpen(t, g, me, src, ActivationAbility{Label: "{T}: Draw a card."}) {
		t.Error("the same statement must still shut a non-mana ability")
	}
}

// --- the grant ------------------------------------------------------

// Leonin Shikari's shape: a statement that opens the sorcery window
// on the abilities it names, and on nothing else.
func TestActivationTimingGrantOpensTheWindow(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	const oracle = "test-activation-timing-grant"

	equip := ActivationAbility{Label: "Equip {2}", SorcerySpeed: true, Equip: true}
	other := ActivationAbility{Label: "Cycling {2}", SorcerySpeed: true}

	src := activationTimingSource(g, me, "Test Shikari", oracle)
	if activationOpen(t, g, me, src, equip) {
		t.Fatal("setup: equip is sorcery-speed with nothing granting anything")
	}

	withCatalogActivationTimings(t, oracle, ActivationTiming{
		Label:  "You may activate equip abilities any time you could cast an instant.",
		Timing: TimingFlash,
		Covers: func(q ActivationQuery) bool {
			return q.Ability.Equip && q.Controller == q.Source.Controller
		},
	})

	if !activationOpen(t, g, me, src, equip) {
		t.Error("the grant did not open the equip ability's window")
	}
	if activationOpen(t, g, me, src, other) {
		t.Error("the grant reached an ability it does not name")
	}
	if activationOpen(t, g, g.Seats[1], src, equip) {
		t.Error("the grant reached a player it does not name")
	}
}

// CR 101.1: a card beats a rule, which is the only reason a printed
// clause can reach CR 606.3's sorcery half at all. The Wandering
// Emperor and Teferi, Master of Time are the two cards.
func TestActivationTimingGrantReachesLoyaltyAbilities(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	const oracle = "test-activation-timing-loyalty"

	withCatalogActivationTimings(t, oracle, ActivationTiming{
		Label:  "You may activate loyalty abilities of this any time you could cast an instant.",
		Timing: TimingFlash,
		Covers: func(q ActivationQuery) bool {
			return q.Ability.Loyalty && q.Card.InstanceID == q.Source.InstanceID
		},
	})
	walker := activationTimingSource(g, me, "Test Walker", oracle)

	if !activationOpen(t, g, me, walker, ActivationAbility{Label: "+1: …", Loyalty: true}) {
		t.Error("CR 101.1: a printed clause overrides CR 606.3's sorcery window")
	}
	// The same statement says nothing about a DIFFERENT object's
	// loyalty ability — the self-reference half of CR 201.5.
	other := activationTimingSource(g, me, "Other Walker", "test-activation-timing-other")
	if activationOpen(t, g, me, other, ActivationAbility{Label: "+1: …", Loyalty: true}) {
		t.Error("the statement opened another permanent's loyalty ability")
	}
}

// CR 613.1f: a source under an ability-removing effect stops saying
// it, and that falls out of CatalogAbilityKey rather than out of a
// rule this file has to remember. The same argument
// ActivationRestrictionsForCard makes.
func TestActivationTimingStopsWhenTheSourceLosesItsAbilities(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	const oracle = "test-activation-timing-removed"

	withCatalogActivationTimings(t, oracle, ActivationTiming{
		Label:  "Test — equip at instant speed.",
		Timing: TimingFlash,
		Covers: alwaysCovers,
	})
	src := activationTimingSource(g, me, "Test Shikari", oracle)
	equip := ActivationAbility{Label: "Equip {2}", SorcerySpeed: true, Equip: true}
	if !activationOpen(t, g, me, src, equip) {
		t.Fatal("setup: the grant is not live")
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID != src.InstanceID {
				continue
			}
			eff := g.Battlefield.Cards[i].Effective()
			eff.AbilitiesRemoved = true
			g.Battlefield.Cards[i].effective = &eff
		}
	})
	if activationOpen(t, g, me, src, equip) {
		t.Error("a source that has lost its abilities still granted")
	}
}

// --- the restrictions, and CR 101.2 ---------------------------------

// Nothing in the catalog declares the restricting values today (a
// printed activation restriction says "can't be activated" and goes
// through ActivationGateLocked). The fold is tested anyway, because
// the placement IS CR 101.2 and a grant-only read would have to be
// rewritten to state it.
func TestActivationTimingRestrictionBeatsGrant(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)

	const grantOracle = "test-activation-timing-open"
	const shutOracle = "test-activation-timing-shut"
	prev := CatalogActivationTimings
	CatalogActivationTimings = func(key string) []ActivationTiming {
		switch key {
		case grantOracle:
			return []ActivationTiming{{
				Label:  "Test — everything at instant speed.",
				Timing: TimingFlash,
				Covers: alwaysCovers,
			}}
		case shutOracle:
			return []ActivationTiming{{
				Label:  "Test — activations only at sorcery speed.",
				Timing: TimingSorcery,
				Covers: alwaysCovers,
			}}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogActivationTimings = prev })

	src := activationTimingSource(g, me, "Test Source", "test-activation-timing-plain")
	plain := ActivationAbility{Label: "{T}: Draw a card."}
	if !activationOpen(t, g, me, src, plain) {
		t.Fatal("setup: a plain ability is instant-speed")
	}

	activationTimingSource(g, me, "Test Lock", shutOracle)
	if activationOpen(t, g, me, src, plain) {
		t.Error("a TimingSorcery statement did not shut an instant-speed ability")
	}
	// The grant lands second and must NOT win: CR 101.2 says "can't"
	// beats "can", and the read applies the restrictions last.
	activationTimingSource(g, me, "Test Opener", grantOracle)
	if activationOpen(t, g, me, src, plain) {
		t.Error("CR 101.2: a grant got past a restriction")
	}
}

// TimingYourTurnOnly refuses outright rather than narrowing to the
// sorcery window — the same reading CastTimingOpenLocked gives it,
// because Dosan's clause is a gate on WHOSE turn it is.
func TestActivationTimingYourTurnOnly(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-activation-timing-your-turn"
	withCatalogActivationTimings(t, oracle, ActivationTiming{
		Label:  "Test — players activate abilities only during their own turns.",
		Timing: TimingYourTurnOnly,
		Covers: alwaysCovers,
	})
	src := activationTimingSource(g, me, "Test Dosan", oracle)
	plain := ActivationAbility{Label: "{T}: Draw a card."}

	advanceTo(t, g, StepPrecombatMain)
	if !activationOpen(t, g, me, src, plain) {
		t.Error("the active player's own instant-speed window is left open")
	}
	if activationOpen(t, g, g.Seats[1], src, plain) {
		t.Error("a non-active player was not refused")
	}
}

// A statement whose Covers is nil, or whose Timing says nothing,
// changes nothing. effects.Register refuses both, so this is about
// the read being safe rather than about a card that can exist.
func TestActivationTimingIgnoresEmptyStatements(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	const oracle = "test-activation-timing-empty"

	withCatalogActivationTimings(t, oracle,
		ActivationTiming{Label: "no timing", Covers: alwaysCovers},
		ActivationTiming{Label: "no predicate", Timing: TimingFlash},
	)
	src := activationTimingSource(g, me, "Test Blank", oracle)
	if activationOpen(t, g, me, src, ActivationAbility{Label: "Equip {2}", SorcerySpeed: true}) {
		t.Error("an under-declared statement opened a window")
	}
}

// ActivationAbilityOf is the one place a shape becomes an identity,
// so CR 606.3 cannot be spelled differently in the three callers.
func TestActivationAbilityOfCarriesTheLoyaltyRule(t *testing.T) {
	loyalty := 1
	got := ActivationAbilityOf(ActivatedAbilityShape{
		Label: "+1: …",
		Cost:  AbilityCost{Loyalty: &loyalty},
	})
	if !got.Loyalty {
		t.Error("a loyalty cost did not make the identity a loyalty ability")
	}
	if got.SorcerySpeed {
		t.Error("CR 606.3 rides Loyalty, not the printed SorcerySpeed flag")
	}
	if got.Mana {
		t.Error("an activated ability is never a mana ability")
	}
	equip := ActivationAbilityOf(ActivatedAbilityShape{Label: "Equip {2}", SorcerySpeed: true, Equip: true})
	if !equip.Equip || !equip.SorcerySpeed || equip.Loyalty {
		t.Errorf("equip identity: got %+v", equip)
	}
}

// The sandbox manual loyalty verb reads the same function, so a
// statement about loyalty abilities reaches a planeswalker the
// catalog has never heard of (#1208's fourth caller).
func TestSandboxActivateLoyaltyReadsTheTimingStatement(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	const oracle = "test-activation-timing-sandbox"

	pwID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: pwID,
		Name:       "Unknown Walker",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleID:   oracle,
		Owner:      me.ID,
		Controller: me.ID,
		Counters:   map[string]int{"loyalty": 3},
	})
	if err := g.ActivateLoyalty(me.ID, pwID, "+1", 1); err != ErrSorcerySpeedRequired {
		t.Fatalf("setup: upkeep activation: got %v, want ErrSorcerySpeedRequired", err)
	}

	withCatalogActivationTimings(t, oracle, ActivationTiming{
		Label:  "You may activate loyalty abilities of this any time you could cast an instant.",
		Timing: TimingFlash,
		Covers: func(q ActivationQuery) bool { return q.Ability.Loyalty },
	})
	if err := g.ActivateLoyalty(me.ID, pwID, "+1", 1); err != nil {
		t.Errorf("with the statement live: got %v, want nil", err)
	}
}
