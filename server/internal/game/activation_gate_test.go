package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// activation_gate_test.go — #1210, ADR 0073's amendment of
// 2026-09-22. The gate has four callers and the tests here pin the
// three inside this package (the enumerator's and the view's live in
// their own packages): both activation paths, and the auto-tapper,
// which is the one that would silently SUCCEED rather than refuse.

const (
	totemOracle  = "test-activation-totem"
	needleOracle = "test-activation-needle"
	elfOracle    = "test-activation-elf"
	rockOracle   = "test-activation-rock"
)

// stubActivationRestrictions wires CatalogActivationRestrictions for
// one oracle ID and restores the previous hook at the end of the
// test.
func stubActivationRestrictions(t *testing.T, oracleID string, rules []ActivationRestriction) {
	t.Helper()
	prev := CatalogActivationRestrictions
	CatalogActivationRestrictions = func(id string) []ActivationRestriction {
		if id == oracleID {
			return rules
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { CatalogActivationRestrictions = prev })
}

// stubActivatedFor wires CatalogActivatedAbilities for one oracle ID.
func stubActivatedFor(t *testing.T, oracleID string, abilities ...ActivatedAbilityShape) {
	t.Helper()
	prev := CatalogActivatedAbilities
	CatalogActivatedAbilities = func(key string) []ActivatedAbilityShape {
		if key == oracleID {
			return abilities
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogActivatedAbilities = prev })
}

// pushGateCard puts a permanent keyed on `oracle` on the battlefield
// under `controller` and returns its instance ID.
func pushGateCard(g *Game, name, typeLine, oracle string, controller uuid.UUID) uuid.UUID {
	c := NewCard(name, controller)
	c.TypeLine = typeLine
	c.OracleID = oracle
	c.Controller = controller
	// Old enough to pay a {T} cost (CR 302.6) — the gate is what
	// these tests are about, not summoning sickness.
	c.SummonedThisTurn = false
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// cursedTotemRule is "Activated abilities of creatures can't be
// activated" — NO mana exemption, which is the half of the printed
// card that makes it more than a Linvala.
func cursedTotemRule() ActivationRestriction {
	return ActivationRestriction{
		Label:   "Activated abilities of creatures can't be activated.",
		Forbids: func(q ActivationQuery) bool { return q.Card.IsCreature() },
	}
}

// pithingNeedleRule is "Activated abilities of sources named X can't
// be activated unless they're mana abilities" — the OTHER half, keyed
// on a name rather than on the chosen-name field, which lands with
// the prompt.
func pithingNeedleRule(named string) ActivationRestriction {
	return ActivationRestriction{
		Label: "Activated abilities of sources with the chosen name can't be activated unless they're mana abilities.",
		Forbids: func(q ActivationQuery) bool {
			if q.Ability.Mana {
				return false
			}
			return q.Card.Name == named
		},
	}
}

// TestCursedTotemStopsACreatureAbilityAndItsManaAbilityToo is the
// headline pair: one restriction, two activation paths, and the
// restriction — not the call site — decides that the mana one is
// caught as well.
func TestCursedTotemStopsACreatureAbilityAndItsManaAbilityToo(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	stubActivationRestrictions(t, totemOracle, []ActivationRestriction{cursedTotemRule()})
	stubActivatedFor(t, elfOracle, ActivatedAbilityShape{
		Label:  "{T}: Draw a card.",
		Cost:   AbilityCost{Tap: true},
		Effect: func(*Game, *StackItem) error { return nil },
	})
	withCatalogHook(t, func(id string) []ManaAbilityShape {
		if id == elfOracle {
			return []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}}
		}
		return nil
	})

	elf := pushGateCard(g, "Test Elf", "Creature — Elf", elfOracle, me.ID)
	pushGateCard(g, "Cursed Totem", "Artifact", totemOracle, me.ID)

	err := g.ActivateCatalogAbility(me.ID, elf, 0, ActivateAbilityParams{})
	if !errors.Is(err, ErrCantActivate) {
		t.Fatalf("non-mana activation: err = %v, want ErrCantActivate", err)
	}
	var cant *CantActivateError
	if !errors.As(err, &cant) {
		t.Fatalf("err = %T, want *CantActivateError carrying the printed clause", err)
	}
	if cant.Reason != "Activated abilities of creatures can't be activated." {
		t.Errorf("Reason = %q, want the printed clause", cant.Reason)
	}

	// CR 605.1a: the mana ability is an activated ability, and Cursed
	// Totem prints no exemption. Same function, same answer.
	if err := g.ActivateManaAbility(me.ID, elf, 0, ManaAbilityParams{}); !errors.Is(err, ErrCantActivate) {
		t.Fatalf("mana activation: err = %v, want ErrCantActivate", err)
	}

	// And nothing was paid on either path.
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == elf && c.Tapped {
				t.Error("the elf is tapped — a refused activation pays nothing")
			}
		}
	})
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want empty — the refused mana ability produced nothing", me.ManaPool)
	}
}

// TestAnActivationRestrictionThatExemptsManaAbilitiesLeavesThemAlone
// is the other half of the same decision: Pithing Needle's clause
// prints "unless they're mana abilities", and because the exemption
// lives on the RESTRICTION rather than on the call site, the very
// same gate lets the mana ability through.
func TestAnActivationRestrictionThatExemptsManaAbilitiesLeavesThemAlone(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	stubActivationRestrictions(t, needleOracle, []ActivationRestriction{pithingNeedleRule("Test Elf")})
	stubActivatedFor(t, elfOracle, ActivatedAbilityShape{
		Label:  "{T}: Draw a card.",
		Cost:   AbilityCost{Tap: true},
		Effect: func(*Game, *StackItem) error { return nil },
	})
	withCatalogHook(t, func(id string) []ManaAbilityShape {
		if id == elfOracle {
			return []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}}
		}
		return nil
	})

	elf := pushGateCard(g, "Test Elf", "Creature — Elf", elfOracle, me.ID)
	pushGateCard(g, "Pithing Needle", "Artifact", needleOracle, me.ID)

	if err := g.ActivateCatalogAbility(me.ID, elf, 0, ActivateAbilityParams{}); !errors.Is(err, ErrCantActivate) {
		t.Fatalf("non-mana activation: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateManaAbility(me.ID, elf, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("mana activation: err = %v, want nil — the clause exempts mana abilities", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %v, want one {G} — the mana ability was not restricted", me.ManaPool)
	}
}

// TestAnActivationRestrictionBindsEveryPlayerAndOnlyTheNamedSources
// is the "and nothing else" half of the Pithing Needle assertion: the
// named source is stopped for EVERY player (the clause says "sources",
// not "sources your opponents control"), and a source with another
// name is stopped for nobody.
func TestAnActivationRestrictionBindsEveryPlayerAndOnlyTheNamedSources(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]

	stubActivationRestrictions(t, needleOracle, []ActivationRestriction{pithingNeedleRule("Test Elf")})
	probe := ActivatedAbilityShape{
		Label:  "{T}: Draw a card.",
		Cost:   AbilityCost{Tap: true},
		Effect: func(*Game, *StackItem) error { return nil },
	}
	prev := CatalogActivatedAbilities
	CatalogActivatedAbilities = func(key string) []ActivatedAbilityShape {
		if key == elfOracle || key == rockOracle {
			return []ActivatedAbilityShape{probe}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogActivatedAbilities = prev })

	// The Needle belongs to seat 0; the named source to seat 1. A
	// "can't be activated" static is not a "your opponents" clause.
	pushGateCard(g, "Pithing Needle", "Artifact", needleOracle, me.ID)
	theirElf := pushGateCard(g, "Test Elf", "Creature — Elf", elfOracle, them.ID)
	myElf := pushGateCard(g, "Test Elf", "Creature — Elf", elfOracle, me.ID)
	myRock := pushGateCard(g, "Test Rock", "Artifact", rockOracle, me.ID)

	if err := g.ActivateCatalogAbility(them.ID, theirElf, 0, ActivateAbilityParams{}); !errors.Is(err, ErrCantActivate) {
		t.Errorf("opponent's named source: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, myElf, 0, ActivateAbilityParams{}); !errors.Is(err, ErrCantActivate) {
		t.Errorf("own named source: err = %v, want ErrCantActivate — the clause names sources, not opponents", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, myRock, 0, ActivateAbilityParams{}); err != nil {
		t.Errorf("a differently named source: err = %v, want nil", err)
	}
}

// TestAnActivationRestrictionStopsWhenItsSourceLosesItsAbilities is
// CR 613.1f, and it falls out of ActivationRestrictionsForCard going
// through CatalogAbilityKey rather than out of any rule written here
// — the same property the cast gate has. A Darksteel Mutation'd
// Cursed Totem restricts nobody.
func TestAnActivationRestrictionStopsWhenItsSourceLosesItsAbilities(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	stubActivationRestrictions(t, totemOracle, []ActivationRestriction{cursedTotemRule()})
	stubActivatedFor(t, elfOracle, ActivatedAbilityShape{
		Label:  "{T}: Draw a card.",
		Cost:   AbilityCost{Tap: true},
		Effect: func(*Game, *StackItem) error { return nil },
	})

	elf := pushGateCard(g, "Test Elf", "Creature — Elf", elfOracle, me.ID)
	totem := pushGateCard(g, "Cursed Totem", "Artifact", totemOracle, me.ID)

	if err := g.ActivateCatalogAbility(me.ID, elf, 0, ActivateAbilityParams{}); !errors.Is(err, ErrCantActivate) {
		t.Fatalf("with the Totem's abilities intact: err = %v, want ErrCantActivate", err)
	}

	// Silence the Totem the way layer 6 does.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID != totem {
				continue
			}
			eff := g.Battlefield.Cards[i].Effective()
			eff.AbilitiesRemoved = true
			g.Battlefield.Cards[i].effective = &eff
		}
	})
	g.ReadSnapshot(func() {
		var src Card
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == totem {
				src = c
			}
		}
		if got := ActivationRestrictionsForCard(src); len(got) != 0 {
			t.Fatalf("a silenced Totem still contributes %d restrictions, want 0", len(got))
		}
	})
}

// TestTheAutoTapperSkipsASourceUnderAnActivationRestriction is the
// caller the CAST gate has no equivalent of, and the one whose
// failure mode is not a refusal. materializePlanLocked taps the
// permanent and mints its mana directly — it never goes through
// ActivateManaAbility — so without the gate in gatherTapSources the
// plan would SUCCEED and hand the player mana Cursed Totem forbids.
func TestTheAutoTapperSkipsASourceUnderAnActivationRestriction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	withCatalogHook(t, func(id string) []ManaAbilityShape {
		if id == elfOracle {
			return []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}}
		}
		return nil
	})
	elf := pushGateCard(g, "Test Elf", "Creature — Elf", elfOracle, me.ID)

	cost, err := ParseCost("{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok || len(plan) != 1 || plan[0] != elf {
		t.Fatalf("without a restriction the elf should be the plan; ok=%v plan=%v", ok, plan)
	}

	stubActivationRestrictions(t, totemOracle, []ActivationRestriction{cursedTotemRule()})
	pushGateCard(g, "Cursed Totem", "Artifact", totemOracle, me.ID)

	if _, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Fatal("the auto-tapper planned a Cursed-Totem'd creature — the plan would have SUCCEEDED and produced forbidden mana")
	}
}

// TestActivationGateIgnoresARestrictionWithNoForbids pins the
// under-declared card file: nil Forbids is a no-op, never a lockout.
// Register panics on it at boot, so this only ever fires for a
// restriction built in code — but "the safe direction" is worth
// asserting rather than assuming.
func TestActivationGateIgnoresARestrictionWithNoForbids(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	stubActivationRestrictions(t, totemOracle, []ActivationRestriction{{Label: "Says nothing."}})
	pushGateCard(g, "Quiet Totem", "Artifact", totemOracle, me.ID)

	var err error
	g.ReadSnapshot(func() {
		err = g.ActivationGateLocked(me.ID, Card{Name: "Anything"}, ZoneBattlefield, ActivationAbility{Label: "x"})
	})
	if err != nil {
		t.Fatalf("ActivationGateLocked = %v, want nil", err)
	}
}
