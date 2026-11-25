package game

import (
	"testing"

	"github.com/google/uuid"
)

// cant_be_enchanted_test.go — #6203 (CR 303.4, 303.4c, 303.4j, 701.12b):
// "can't be enchanted" (Tetravite), "can't be enchanted by other Auras"
// (Consecrate Land, Anti-Magic Aura), Guardian Beast's new-attachments-only
// form, and "other players can't gain control of them".

const (
	cbeConsecrate = "test-6203-consecrate"
	cbeTetravite  = "test-6203-tetravite"
	cbeGuardian   = "test-6203-guardian"
)

func withCantBeEnchantedStatics(t *testing.T) {
	t.Helper()
	self := func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID }
	// Every test Aura is "Enchant permanent", so the only thing that can
	// make one illegal is the effect under test.
	withCatalogTargetSpec(t, func(string) *TargetSpec {
		return &TargetSpec{Mode: "permanent", Zones: []ZoneKind{ZoneBattlefield},
			CardOK: func(_ *Game, _ uuid.UUID, _ Card, _ ZoneKind) bool { return true }, Min: 1, Max: 1}
	})
	withStaticAbilities(t, func(key string) []StaticAbility {
		switch key {
		case cbeConsecrate:
			return []StaticAbility{{Layer: Layer6Ability, AppliesTo: self,
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.ExtraRestrictions |= EnchantedCantBeEnchantedByOthers
				}}}
		case cbeTetravite:
			return []StaticAbility{{Layer: Layer6Ability, AppliesTo: self,
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.ExtraRestrictions |= CantBeEnchanted }}}
		case cbeGuardian:
			return []StaticAbility{{Layer: Layer6Ability,
				AppliesTo: func(target *Card, _ *Game, source *Card) bool {
					return !source.Tapped && target.Controller == source.Controller && target.IsArtifact() && !target.IsCreature()
				},
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.ExtraRestrictions |= CantBecomeEnchanted | OthersCantGainControl
				}}}
		}
		return nil
	})
}

func cbeAura(g *Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	if oracle == "" {
		oracle = "test-6203-plain-aura"
	}
	return pushTypedTestCard(g, Card{Name: name, TypeLine: "Enchantment — Aura", OracleID: oracle, Owner: owner, Controller: owner})
}

func cbeAttach(t *testing.T, g *Game, aura, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if err := g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
}

func cbeAttachedTo(g *Game, aura, host uuid.UUID) bool {
	ok := false
	g.ReadSnapshot(func() {
		if c := findBattlefieldCard(g, aura); c != nil {
			ok = c.IsAttachedTo(host)
		}
	})
	return ok
}

func cbeOnBattlefield(g *Game, id uuid.UUID) bool {
	ok := false
	g.ReadSnapshot(func() { ok = findBattlefieldCard(g, id) != nil })
	return ok
}

func cbeCanTarget(g *Game, host uuid.UUID, aura *Card, by uuid.UUID) bool {
	ok := false
	g.ReadSnapshot(func() { ok = canBeTargetedBy(g, findBattlefieldCard(g, host), ZoneBattlefield, SourceCast(by, aura)) })
	return ok
}

// Consecrate Land's acceptance: an Aura already on the land when it
// arrives goes to the graveyard, and another Aura can't target the land
// afterwards; the Consecrate Land itself stays.
func TestCantBeEnchantedByOtherAuras(t *testing.T) {
	withCantBeEnchantedStatics(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	land := pushTypedTestCard(g, Card{Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me, Controller: me})
	first := cbeAura(g, me, "Wild Growth", "")
	cbeAttach(t, g, first, land)
	consecrate := cbeAura(g, me, "Consecrate", cbeConsecrate)
	cbeAttach(t, g, consecrate, land)
	g.RunStateChecksForTest()
	if cbeOnBattlefield(g, first) {
		t.Error("the Aura already on the land stayed (CR 303.4c)")
	}
	if !cbeAttachedTo(g, consecrate, land) {
		t.Error("Consecrate Land fell off its own land")
	}
	next := Card{InstanceID: uuid.New(), Name: "Utopia Sprawl", TypeLine: "Enchantment — Aura", Owner: me, Controller: me}
	if cbeCanTarget(g, land, &next, me) {
		t.Error("another Aura spell can target the land")
	}
	instant := Card{InstanceID: uuid.New(), Name: "Stone Rain", TypeLine: "Sorcery", Owner: me, Controller: me}
	if !cbeCanTarget(g, land, &instant, me) {
		t.Error("a non-Aura spell can't target the land")
	}
	later := cbeAura(g, me, "Fertile Ground", "")
	cbeAttach(t, g, later, land)
	if cbeAttachedTo(g, later, land) {
		t.Error("another Aura became attached to the land (CR 303.4j)")
	}
}

// Tetravite: "This token can't be enchanted." Nothing attaches, and an
// Aura spell can't target it.
func TestCantBeEnchanted(t *testing.T) {
	withCantBeEnchantedStatics(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	tok := pushTypedTestCard(g, Card{Name: "Tetravite", TypeLine: "Token Artifact Creature — Tetravite", OracleID: cbeTetravite, Owner: me, Controller: me, Power: 1, Toughness: 1})
	aura := cbeAura(g, me, "Holy Strength", "")
	cbeAttach(t, g, aura, tok)
	if cbeAttachedTo(g, aura, tok) {
		t.Error("an Aura became attached to a permanent that can't be enchanted")
	}
	spell := Card{InstanceID: uuid.New(), Name: "Pacifism", TypeLine: "Enchantment — Aura", Owner: me, Controller: me}
	if cbeCanTarget(g, tok, &spell, me) {
		t.Error("an Aura spell can target a permanent that can't be enchanted")
	}
}

// Guardian Beast: while it is untapped, a noncreature artifact its
// controller controls can't become enchanted, an Aura already on it
// stays, and another player can't gain control of it; tapped, both stop.
func TestGuardianBeastNewAurasAndControl(t *testing.T) {
	withCantBeEnchantedStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	rock := pushTypedTestCard(g, Card{Name: "Mind Stone", TypeLine: "Artifact", Owner: me, Controller: me})
	old := cbeAura(g, opp, "Old Aura", "")
	cbeAttach(t, g, old, rock)
	beast := pushTypedTestCard(g, Card{Name: "Guardian Beast", TypeLine: "Creature — Beast", OracleID: cbeGuardian, Owner: me, Controller: me, Power: 2, Toughness: 4})
	g.RunStateChecksForTest()
	if !cbeAttachedTo(g, old, rock) {
		t.Error("an Aura already attached was removed")
	}
	fresh := cbeAura(g, opp, "New Aura", "")
	cbeAttach(t, g, fresh, rock)
	if cbeAttachedTo(g, fresh, rock) {
		t.Error("a new Aura became attached under the Beast")
	}
	var stolen bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		stolen = g.GainControlForEffect(uuid.New(), rock, opp, IndefiniteDuration(), "steal")
	})
	if stolen {
		t.Error("an opponent gained control of the artifact under the Beast")
	}
	g.WithWriteLock(func() { findBattlefieldCard(g, beast).Tapped = true; g.layerVersion.Add(1) })
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		stolen = g.GainControlForEffect(uuid.New(), rock, opp, IndefiniteDuration(), "steal")
	})
	if !stolen {
		t.Error("with the Beast tapped, the control change was still refused")
	}
}

// CR 701.12b: an exchange one half of which can't happen does not happen.
func TestExchangeRefusedWhenOneHalfIsBarred(t *testing.T) {
	withCantBeEnchantedStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	rock := pushTypedTestCard(g, Card{Name: "Mind Stone", TypeLine: "Artifact", Owner: me, Controller: me})
	theirs := pushTypedTestCard(g, Card{Name: "Their Stone", TypeLine: "Artifact", Owner: opp, Controller: opp})
	pushTypedTestCard(g, Card{Name: "Guardian Beast", TypeLine: "Creature — Beast", OracleID: cbeGuardian, Owner: me, Controller: me})
	var ok bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		ok = g.ExchangeControlForEffect(uuid.New(), rock, theirs, "swap")
	})
	if ok {
		t.Error("the exchange happened although one half is barred")
	}
}
