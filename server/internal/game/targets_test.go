package game

import (
	"testing"

	"github.com/google/uuid"
)

// targets_test.go — S20 sub-PR 1: TargetSpec legality at announce
// and resolution, legal-target enumeration, colour derivation.

// withCatalogTargetSpec swaps the CatalogTargetSpec hook for the
// test's duration.
func withCatalogTargetSpec(t *testing.T, fn func(oracleID string) *TargetSpec) {
	t.Helper()
	prev := CatalogTargetSpec
	CatalogTargetSpec = fn
	t.Cleanup(func() { CatalogTargetSpec = prev })
}

// nonBlackCreatureSpec is Doom Blade's clause.
func nonBlackCreatureSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "creature",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature() && !c.HasColor("B")
		},
		Min: 1, Max: 1,
	}
}

func pushColoredCreature(g *Game, owner *Player, name, manaCost string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Test"
	c.ManaCost = manaCost
	c.Power, c.Toughness = 2, 2
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func TestEffectiveColorsFromManaCostAndStamped(t *testing.T) {
	c := Card{ManaCost: "{1}{W/U}{B}"}
	got := c.EffectiveColors()
	if len(got) != 3 || !c.HasColor("W") || !c.HasColor("U") || !c.HasColor("B") || c.HasColor("R") {
		t.Errorf("derived colours = %v", got)
	}
	stamped := Card{ManaCost: "{2}", Colors: []string{"G"}}
	if !stamped.HasColor("G") || stamped.IsColorless() {
		t.Errorf("stamped Colors should win over the mana-cost derivation")
	}
	if !(Card{ManaCost: "{3}"}).IsColorless() {
		t.Errorf("{3} should be colourless")
	}
}

func TestLegalTargetsFiltersByPredicateAndZone(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := pushColoredCreature(g, opp, "White Knight", "{W}{W}")
	black := pushColoredCreature(g, opp, "Black Knight", "{B}{B}")
	mine := pushColoredCreature(g, me, "My Bear", "{1}{G}")
	rock := NewCard("Sol Ring", opp.ID)
	rock.TypeLine = "Artifact"
	g.Battlefield.PushTop(rock)
	// A creature in a graveyard must not show up for a battlefield spec.
	dead := NewCard("Dead Guy", opp.ID)
	dead.TypeLine = "Creature — Zombie"
	opp.Graveyard.PushTop(dead)

	lt := g.LegalTargetsForEffect(me.ID, nonBlackCreatureSpec())
	has := func(id uuid.UUID) bool {
		for _, c := range lt.Cards {
			if c == id {
				return true
			}
		}
		return false
	}
	if !has(white) || !has(mine) {
		t.Errorf("non-black creatures missing from legal set: %v", lt.Cards)
	}
	if has(black) || has(rock.InstanceID) || has(dead.InstanceID) {
		t.Errorf("illegal candidates present: %v", lt.Cards)
	}
	if len(lt.Players) != 0 {
		t.Errorf("creature spec must not list players")
	}
}

func TestCastSpellRejectsIllegalTargetAndAcceptsLegal(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for g.Turn.Step != StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	white := pushColoredCreature(g, opp, "White Knight", "{W}{W}")
	black := pushColoredCreature(g, opp, "Black Knight", "{B}{B}")
	const oracle = "test-doom-blade"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return nonBlackCreatureSpec()
		}
		return nil
	})
	blade := NewCard("Doom Blade", me.ID)
	blade.TypeLine = "Instant"
	blade.OracleID = oracle
	me.Hand.PushTop(blade)

	err := g.CastSpell(me.ID, blade.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: black}},
	})
	if err != ErrIllegalTarget {
		t.Fatalf("black creature: got %v, want ErrIllegalTarget", err)
	}
	if !me.Hand.Contains(blade.InstanceID) {
		t.Fatalf("rejected cast must leave the card in hand")
	}
	err = g.CastSpell(me.ID, blade.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetPlayer, ID: opp.ID}},
	})
	if err != ErrIllegalTarget {
		t.Fatalf("player for a creature spec: got %v, want ErrIllegalTarget", err)
	}
	err = g.CastSpell(me.ID, blade.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: white}},
	})
	if err != nil {
		t.Fatalf("legal target rejected: %v", err)
	}
}

func TestCastSpellRejectsWrongTargetCount(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for g.Turn.Step != StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	a := pushColoredCreature(g, opp, "A", "{W}")
	b := pushColoredCreature(g, opp, "B", "{W}")
	const oracle = "test-single-target"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return nonBlackCreatureSpec()
		}
		return nil
	})
	spell := NewCard("Single", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	me.Hand.PushTop(spell)
	err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: a}, {Kind: TargetCard, ID: b}},
	})
	if err != ErrInvalidParam {
		t.Fatalf("two targets for Max 1: got %v, want ErrInvalidParam", err)
	}
}

// TestResolutionRecheckUsesPredicate: the target was legal at
// announce and still exists at resolution, but no longer satisfies
// the predicate (it stopped being a creature) — CR 608.2b counters
// the spell by game rules.
func TestResolutionRecheckUsesPredicate(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for g.Turn.Step != StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	target := pushColoredCreature(g, opp, "Mutable", "{W}")
	const oracle = "test-recheck"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return nonBlackCreatureSpec()
		}
		return nil
	})
	spell := NewCard("Recheck", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	me.Hand.PushTop(spell)
	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// "In response": the target becomes an artifact (loses creature
	// type). Still on the battlefield, so the existence check alone
	// would let the spell resolve.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == target {
				g.Battlefield.Cards[i].TypeLine = "Artifact"
			}
		}
	})
	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Fatalf("resolveTopOfStackLocked: %v", err)
		}
	})
	var fizzled bool
	for _, ev := range g.Events {
		if ev.Kind == EventFizzle && ev.CardID == spell.InstanceID {
			fizzled = true
		}
	}
	if !fizzled {
		t.Errorf("spell whose target failed the predicate at resolution was not countered by game rules")
	}
	if !me.Graveyard.Contains(spell.InstanceID) {
		t.Errorf("fizzled spell should be in its owner's graveyard")
	}
}

func TestSpecWithoutPredicateKeepsExistenceCheck(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src, _ := me.Hand.Top()
	// Free-form (nil spec) announce still accepts any card ID.
	if err := g.ActivateAbility(me.ID, src.InstanceID, AbilityParams{
		Label:   "free-form",
		Targets: []TargetRef{{Kind: TargetCard, ID: src.InstanceID}},
	}); err != nil {
		t.Fatalf("free-form ability with an arbitrary target: %v", err)
	}
}
