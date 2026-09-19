package game

import (
	"testing"

	"github.com/google/uuid"
)

// protection_keywords_test.go — S23: hexproof and shroud at the
// targeting choke point. Before this, targets.go read no keywords at
// all: hexproof was a badge with no rule behind it, and The
// Wandering Rescuer shipped with a grant nothing enforced.
//
// The four cases that matter are the asymmetry (an opponent is
// refused, you are not), shroud's symmetry (nobody may target),
// the CR 608.2b re-check (a target that GAINS hexproof after
// announce fizzles the spell), and the boundary the gate must not
// cross (paying a cost is not targeting).

// pushProtectedCreature puts a creature with printed keywords onto the
// battlefield under `owner`. No layer recompute is forced, so
// HasKeyword reads the card's own Keywords slice — the printed road.
func pushProtectedCreature(g *Game, owner *Player, name string, keywords ...string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Test"
	c.ManaCost = "{1}{W}"
	c.Power, c.Toughness = 2, 2
	c.Keywords = keywords
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// anyCreatureSpec is the plainest possible targeting clause —
// "target creature", no narrowing predicate at all. The keyword gate
// has to come from the engine, not from the spec, or every catalog
// card would have to remember to opt in.
func anyCreatureSpec() *TargetSpec {
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

// castAtTarget puts a one-shot instant with `spec` into me's hand
// and casts it at `target`, returning the cast error and the spell's
// instance ID.
func castAtTarget(t *testing.T, g *Game, me *Player, oracle string, spec *TargetSpec, ref TargetRef) (uuid.UUID, error) {
	t.Helper()
	spell := NewCard("Test Removal", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	me.Hand.PushTop(spell)
	err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{Targets: []TargetRef{ref}})
	return spell.InstanceID, err
}

func advanceToMain(t *testing.T, g *Game) {
	t.Helper()
	for g.Turn.Step != StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
}

// TestHexproofRefusesOpponentsButNotItsController is CR 702.11b: a
// hexproof creature can't be the target of spells or abilities YOUR
// OPPONENTS control. Its controller's own Giant Growth still lands.
func TestHexproofRefusesOpponentsButNotItsController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	const oracle = "test-hexproof-gate"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return anyCreatureSpec()
		}
		return nil
	})

	theirs := pushProtectedCreature(g, opp, "Their Hexproof Bear", "hexproof")
	mine := pushProtectedCreature(g, me, "My Hexproof Bear", "hexproof")

	if _, err := castAtTarget(t, g, me, oracle, anyCreatureSpec(),
		TargetRef{Kind: TargetCard, ID: theirs}); err != ErrIllegalTarget {
		t.Fatalf("opponent's hexproof creature: got %v, want ErrIllegalTarget", err)
	}
	if _, err := castAtTarget(t, g, me, oracle, anyCreatureSpec(),
		TargetRef{Kind: TargetCard, ID: mine}); err != nil {
		t.Fatalf("your own hexproof creature must be targetable by you: %v", err)
	}
}

// TestShroudRefusesEveryone is CR 702.18a — shroud has no "your
// opponents" clause, so the controller is shut out too.
func TestShroudRefusesEveryone(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	const oracle = "test-shroud-gate"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return anyCreatureSpec()
		}
		return nil
	})

	theirs := pushProtectedCreature(g, opp, "Their Shrouded Bear", "shroud")
	mine := pushProtectedCreature(g, me, "My Shrouded Bear", "shroud")

	if _, err := castAtTarget(t, g, me, oracle, anyCreatureSpec(),
		TargetRef{Kind: TargetCard, ID: theirs}); err != ErrIllegalTarget {
		t.Fatalf("opponent's shrouded creature: got %v, want ErrIllegalTarget", err)
	}
	if _, err := castAtTarget(t, g, me, oracle, anyCreatureSpec(),
		TargetRef{Kind: TargetCard, ID: mine}); err != ErrIllegalTarget {
		t.Fatalf("your OWN shrouded creature: got %v, want ErrIllegalTarget", err)
	}
}

// TestLegalTargetEnumerationHonoursKeywords: the picker list the
// client and the bot's move enumerator both read must already be
// filtered, or the bot proposes moves the announce gate then
// bounces (the #347 drift class).
func TestLegalTargetEnumerationHonoursKeywords(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	plain := pushProtectedCreature(g, opp, "Plain Bear")
	theirHexproof := pushProtectedCreature(g, opp, "Their Hexproof Bear", "hexproof")
	myHexproof := pushProtectedCreature(g, me, "My Hexproof Bear", "hexproof")
	myShroud := pushProtectedCreature(g, me, "My Shrouded Bear", "shroud")

	lt := g.LegalTargetsForEffect(SourceChooser(me.ID), anyCreatureSpec())
	has := func(id uuid.UUID) bool {
		for _, c := range lt.Cards {
			if c == id {
				return true
			}
		}
		return false
	}
	if !has(plain) {
		t.Errorf("an unprotected creature must stay in the legal set")
	}
	if !has(myHexproof) {
		t.Errorf("your own hexproof creature must stay in YOUR legal set")
	}
	if has(theirHexproof) {
		t.Errorf("an opponent's hexproof creature must not be offered")
	}
	if has(myShroud) {
		t.Errorf("a shrouded creature must not be offered to anyone")
	}

	// The same board from the opponent's side: the asymmetry flips.
	lt = g.LegalTargetsForEffect(SourceChooser(opp.ID), anyCreatureSpec())
	hasOpp := func(id uuid.UUID) bool {
		for _, c := range lt.Cards {
			if c == id {
				return true
			}
		}
		return false
	}
	if !hasOpp(theirHexproof) || hasOpp(myHexproof) {
		t.Errorf("hexproof asymmetry did not flip for the other seat: %v", lt.Cards)
	}
}

// TestResolutionRecheckFizzlesOnGainedHexproof is the CR 608.2b
// half, and the reason this change is only one function: the spell
// was announced at a legal target, the target then GAINED hexproof
// from a Layer 6 grant, and the re-check must counter the spell by
// game rules. The grant goes through RegisterScopedStatic so the
// keyword is read out of Effective(), not off the printed card.
func TestResolutionRecheckFizzlesOnGainedHexproof(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	const oracle = "test-hexproof-recheck"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return anyCreatureSpec()
		}
		return nil
	})

	target := pushProtectedCreature(g, opp, "Naked Bear")
	spell, err := castAtTarget(t, g, me, oracle, anyCreatureSpec(),
		TargetRef{Kind: TargetCard, ID: target})
	if err != nil {
		t.Fatalf("announce at an unprotected creature: %v", err)
	}

	// "In response": the opponent's Swiftfoot Boots land on it.
	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer: Layer6Ability,
			AppliesTo: func(c *Card, _ *Game, _ *Card) bool {
				return c.InstanceID == target
			},
			Apply: func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
				ch.Abilities = append(ch.Abilities, "hexproof")
			},
		}, uuid.New(), "boots", g.UntilEndOfTurnDuration())
	})
	// Force the recompute so Effective() carries the grant, the same
	// way ReadSnapshot does before any consumer reads it.
	g.ReadSnapshot(func() {})

	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Fatalf("resolveTopOfStackLocked: %v", err)
		}
	})

	var fizzled bool
	for _, ev := range g.Events {
		if ev.Kind == EventFizzle && ev.CardID == spell {
			fizzled = true
		}
	}
	if !fizzled {
		t.Errorf("a spell whose target gained hexproof after announce must be countered by game rules")
	}
	if !me.Graveyard.Contains(spell) {
		t.Errorf("fizzled spell should be in its owner's graveyard")
	}
}

// TestPayingACostIgnoresProtectionKeywords is the boundary. CR
// 601.2f/h cost payments — convoke's tap list, an additional
// sacrifice cost — reuse TargetSpec as a predicate but do not
// target, so the gate must not reach them. Getting this wrong would
// make a shrouded creature unconvokable and un-sacrificeable, which
// is a rule nobody printed.
func TestPayingACostIgnoresProtectionKeywords(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	shrouded := pushProtectedCreature(g, me, "Shrouded Bear", "shroud")
	hexproof := pushProtectedCreature(g, me, "Hexproof Bear", "hexproof")

	spec := anyCreatureSpec()
	cands := g.SpecCandidatesForEffect(me.ID, spec)
	has := func(id uuid.UUID) bool {
		for _, c := range cands.Cards {
			if c == id {
				return true
			}
		}
		return false
	}
	if !has(shrouded) || !has(hexproof) {
		t.Errorf("cost payment must see protected permanents: %v", cands.Cards)
	}

	// And the single-ref form used by the convoke / sacrifice
	// validators.
	g.ReadSnapshot(func() {
		for _, id := range []uuid.UUID{shrouded, hexproof} {
			ref := TargetRef{Kind: TargetCard, ID: id}
			if !g.specMatchLocked(SourceChooser(me.ID), spec, ref, false) {
				t.Errorf("cost validator rejected a protected permanent %s", id)
			}
			if id == shrouded && g.specMatchLocked(SourceChooser(me.ID), spec, ref, true) {
				t.Errorf("targeting validator accepted a shrouded permanent")
			}
		}
	})
}

// TestProtectionKeywordsAreBattlefieldOnly: hexproof printed on a
// creature CARD in a graveyard must not stop Regrowth. The keywords
// are abilities of a permanent; the off-battlefield HasKeyword
// fallback would otherwise leak them into every zone.
func TestProtectionKeywordsAreBattlefieldOnly(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	dead := NewCard("Dead Hexproof Bear", opp.ID)
	dead.TypeLine = "Creature — Bear"
	dead.Keywords = []string{"hexproof"}
	opp.Graveyard.PushTop(dead)

	spec := &TargetSpec{
		Mode:  "card_in_graveyard",
		Label: "target creature card in a graveyard",
		Zones: []ZoneKind{ZoneGraveyard},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
	lt := g.LegalTargetsForEffect(SourceChooser(me.ID), spec)
	if len(lt.Cards) != 1 || lt.Cards[0] != dead.InstanceID {
		t.Errorf("graveyard card with printed hexproof was filtered out: %v", lt.Cards)
	}
}

// TestCanBeTargetedByIsTotal pins the helper's edge cases so a
// caller can hand it anything the zone walk produced.
func TestCanBeTargetedByIsTotal(t *testing.T) {
	me := uuid.New()
	if !CanBeTargetedBy(nil, ZoneBattlefield, SourceChooser(me)) {
		t.Error("nil card is an existence problem, not a targeting one")
	}
	c := Card{Controller: uuid.New(), Keywords: []string{"hexproof"}}
	if CanBeTargetedBy(&c, ZoneBattlefield, SourceChooser(me)) {
		t.Error("opponent's hexproof permanent should be refused")
	}
	if !CanBeTargetedBy(&c, ZoneStack, SourceChooser(me)) {
		t.Error("the gate is battlefield-only")
	}
	if !CanBeTargetedBy(&c, ZoneBattlefield, SourceChooser(c.Controller)) {
		t.Error("hexproof does not stop its own controller")
	}
}
