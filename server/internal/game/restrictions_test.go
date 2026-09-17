package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// restrictions_test.go pins the S24 restriction vocabulary at the
// ENGINE gates: a declaration or an activation the rules forbid is
// refused, and one they allow is not.
//
// The enumerator side of the same contract — that internal/legal
// never OFFERS a move these gates would refuse — is pinned in
// internal/legal/restrictions_test.go, which is the half #544 is
// about. The two files are deliberately separate: this one can fail
// with a bad card, that one can only fail with a hung table.
//
// Everything here drives the real layer pipeline through a stubbed
// CatalogStaticAbilities hook, so it exercises the whole chain —
// static ability → Characteristic.Restrictions → gate — rather than
// hand-stamping a Card.effective the way the older combat tests do.

const restrictorOracle = "static-restrictor"

// withRestrictionOn installs a catalog static that puts `r` on every
// battlefield card named `victim`. The restriction arrives when a
// permanent carrying restrictorOracle enters, so a test can choose
// WHEN it lands relative to a declaration.
func withRestrictionOn(t *testing.T, r Restriction, victim string) {
	t.Helper()
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != restrictorOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
				return target.Name == victim
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Restrictions |= r
			},
		}}
	})
}

// pushRestrictor puts the source of the restriction onto the
// battlefield, which is what makes it take effect.
func pushRestrictor(g *Game, owner *Player) uuid.UUID {
	return pushTypedTestCard(g, Card{
		Name: "The Restrictor", TypeLine: "Enchantment", OracleID: restrictorOracle,
		Owner: owner.ID, Controller: owner.ID,
	})
}

// pushRestrictableCreature is a plain 2/2 whose NAME is what the
// stub keys on. It arrives without summoning sickness (CR 302.6) so
// the attack tests fail on the restriction they are about rather
// than on the creature being new.
func pushRestrictableCreature(g *Game, owner *Player, name string) uuid.UUID {
	id := pushTypedTestCard(g, Card{
		Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: owner.ID, Controller: owner.ID,
	})
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	return id
}

// TestCantAttackIsRefusedAtDeclaration is Pacifism's first half at
// the engine boundary: both attack verbs refuse, and an unrestricted
// creature beside it still swings.
func TestCantAttackIsRefusedAtDeclaration(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withRestrictionOn(t, CantAttack, "Pacified Bear")
	pacified := pushRestrictableCreature(g, me, "Pacified Bear")
	free := pushRestrictableCreature(g, me, "Free Bear")
	pushRestrictor(g, me)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttacker(pacified, them.ID); err != ErrCantAttack {
		t.Errorf("DeclareAttacker on a pacified creature = %v, want ErrCantAttack", err)
	}
	if c := findCard(g, pacified); c.AttackingTarget != uuid.Nil {
		t.Error("a refused declaration still set AttackingTarget")
	}
	if c := findCard(g, pacified); c.Tapped {
		t.Error("a refused declaration tapped the creature anyway — the tap is part of the declaration (CR 508.1f)")
	}

	// The bulk verb skips rather than errors, per its contract, and
	// must not report the whole batch as a no-op while a legal
	// attacker is in it.
	declared, err := g.DeclareAttackers([]AttackDeclaration{
		{Attacker: pacified, Target: them.ID},
		{Attacker: free, Target: them.ID},
	})
	if err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	if len(declared) != 1 || declared[0] != free {
		t.Errorf("bulk declaration = %v, want just the unrestricted creature", declared)
	}
}

// TestCantAttackAloneIsNoLegalAttackers — the batch that is nothing
// but restricted creatures is a no-op, and the room layer needs the
// error to leave the undo stack alone.
func TestCantAttackAloneIsNoLegalAttackers(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withRestrictionOn(t, CantAttack, "Pacified Bear")
	pacified := pushRestrictableCreature(g, me, "Pacified Bear")
	pushRestrictor(g, me)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if _, err := g.DeclareAttackers([]AttackDeclaration{
		{Attacker: pacified, Target: them.ID},
	}); err != ErrNoLegalAttackers {
		t.Errorf("all-restricted batch = %v, want ErrNoLegalAttackers", err)
	}
}

// TestCantBlockIsRefusedAtDeclaration is Pacifism's other half, and
// Carrion Feeder's whole drawback.
func TestCantBlockIsRefusedAtDeclaration(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withRestrictionOn(t, CantBlock, "Feeder")
	attacker := pushRestrictableCreature(g, me, "Attacker")
	feeder := pushRestrictableCreature(g, them, "Feeder")
	wall := pushRestrictableCreature(g, them, "Wall")
	pushRestrictor(g, me)

	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	if err := g.DeclareBlocker(feeder, attacker); !errors.Is(err, ErrIllegalBlock) {
		t.Errorf("DeclareBlocker with a can't-block creature = %v, want ErrIllegalBlock", err)
	}
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Errorf("an unrestricted blocker was refused: %v", err)
	}
}

// TestCantBeBlockedIsRefusedAtDeclaration is the Cloak. The
// restriction is on the ATTACKER and it is the DEFENDER's options
// that shrink, which is the distinction the vocabulary exists to
// keep straight.
func TestCantBeBlockedIsRefusedAtDeclaration(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withRestrictionOn(t, CantBeBlocked, "Cloaked")
	cloaked := pushRestrictableCreature(g, me, "Cloaked")
	plain := pushRestrictableCreature(g, me, "Plain")
	wall := pushRestrictableCreature(g, them, "Wall")
	pushRestrictor(g, me)

	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, a := range []uuid.UUID{cloaked, plain} {
		if err := g.DeclareAttacker(a, them.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	if err := g.DeclareBlocker(wall, cloaked); !errors.Is(err, ErrIllegalBlock) {
		t.Errorf("blocking an unblockable attacker = %v, want ErrIllegalBlock", err)
	}
	if err := g.DeclareBlocker(wall, plain); err != nil {
		t.Errorf("blocking the ordinary attacker was refused: %v", err)
	}
}

// TestSeatOwesNoBlockDecisionWhenEveryBlockIsRestricted keeps the
// #328 auto-pass guard honest: a defender whose only creature can't
// block is not being asked anything, and stopping their client on a
// window they cannot act in is the bug that signal exists to avoid.
func TestSeatOwesNoBlockDecisionWhenEveryBlockIsRestricted(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withRestrictionOn(t, CantBlock, "Feeder")
	attacker := pushRestrictableCreature(g, me, "Attacker")
	pushRestrictableCreature(g, them, "Feeder")
	pushRestrictor(g, me)

	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	if g.SeatOwesBlockDecision(them.ID) {
		t.Error("the defender owes a block decision they cannot legally make")
	}
}

// TestRestrictionIsCheckedOnlyAtDeclaration — CR 506.4 lists what
// removes a permanent from combat (leaving the battlefield, changing
// control, ceasing to be a creature) and "acquired a restriction" is
// not on it. A creature pacified after attackers were declared keeps
// attacking and deals its damage.
func TestRestrictionIsCheckedOnlyAtDeclaration(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withRestrictionOn(t, CantAttackOrBlock, "Late Bear")
	bear := pushRestrictableCreature(g, me, "Late Bear")

	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// The Pacifism lands AFTER the declaration.
	pushRestrictor(g, me)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !Restricted(findCard(g, bear), CantAttack) {
		t.Fatal("the restriction did not take effect at all")
	}
	if c := findCard(g, bear); c.AttackingTarget != them.ID {
		t.Error("a mid-combat restriction removed the creature from combat")
	}

	before := them.Life
	advanceIntoStep(t, g, StepCombatDamage)
	if them.Life != before-2 {
		t.Errorf("defender life %d → %d, want 2 damage from a creature that was already attacking",
			before, them.Life)
	}
}

// TestLosingAllAbilitiesDoesNotClearRestrictions is the interop test
// with layer 6's "loses all abilities" (Darksteel Mutation, Song of
// the Dryads). Pacifism's restriction belongs to the AURA, not to
// the creature, so stripping the creature's abilities must not lift
// it. Modelling restrictions as keyword strings would have got this
// backwards, silently.
func TestLosingAllAbilitiesDoesNotClearRestrictions(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const stripperOracle = "static-ability-stripper"
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case restrictorOracle:
			return []StaticAbility{{
				Layer: Layer6Ability,
				AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
					return target.Name == "Mutated Bear"
				},
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.Restrictions |= CantAttackOrBlock
				},
			}}
		case stripperOracle:
			return []StaticAbility{{
				Layer: Layer6Ability,
				AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
					return target.Name == "Mutated Bear"
				},
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.Abilities = nil
				},
			}}
		}
		return nil
	})
	bear := pushTypedTestCard(g, Card{
		Name: "Mutated Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	pushRestrictor(g, me)
	pushTypedTestCard(g, Card{
		Name: "The Stripper", TypeLine: "Enchantment", OracleID: stripperOracle,
		Owner: me.ID, Controller: me.ID,
	})

	c := findCard(g, bear)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if HasKeyword(c, "flying") {
		t.Fatal("the ability-stripping stub did not run")
	}
	if !Restricted(c, CantAttack) || !Restricted(c, CantBlock) {
		t.Error("losing all abilities cleared a restriction that belongs to the source, not the creature")
	}
}

// TestActivationRestrictionSplitsManaFromEverythingElse is the
// Arrest / Faith's Fetters difference, and the reason the vocabulary
// carries two activation bits rather than one.
func TestActivationRestrictionSplitsManaFromEverythingElse(t *testing.T) {
	for _, tc := range []struct {
		name      string
		r         Restriction
		wantOther error
		wantMana  error
	}{
		{"fetters_spares_mana", CantActivate, ErrCantActivate, nil},
		{"arrest_stops_both", CantActivate | CantActivateMana, ErrCantActivate, ErrCantActivate},
		{"unrestricted", 0, nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			withRestrictionOn(t, tc.r, "Fettered Rock")

			// One permanent with BOTH an ordinary activated ability
			// and a mana ability, so the two gates are compared on
			// the same card.
			rock := pushTypedTestCard(g, Card{
				Name: "Fettered Rock", TypeLine: "Artifact",
				Owner: me.ID, Controller: me.ID,
				ManaAbilities: []ManaAbilityShape{{
					Produced: "{C}",
					Label:    "Add {C}",
				}},
				ActivatedAbilities: []ActivatedAbilityShape{{
					Label:  "do a thing",
					Effect: func(*Game, *StackItem) error { return nil },
				}},
			})
			pushRestrictor(g, me)

			if err := g.ActivateCatalogAbility(me.ID, rock, 0, ActivateAbilityParams{}); err != tc.wantOther {
				t.Errorf("ActivateCatalogAbility = %v, want %v", err, tc.wantOther)
			}
			if err := g.ActivateManaAbility(me.ID, rock, 0, ManaAbilityParams{}); err != tc.wantMana {
				t.Errorf("ActivateManaAbility = %v, want %v", err, tc.wantMana)
			}
		})
	}
}

// TestFetteredWalkerCannotUseLoyaltyAbilities — a loyalty ability is
// an activated ability (CR 606.1), and Faith's Fetters enchants a
// PERMANENT, so the planeswalker case is reachable in real play. It
// covers both loyalty paths: the catalog activation and the sandbox
// ActivateLoyalty verb.
func TestFetteredWalkerCannotUseLoyaltyAbilities(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	withRestrictionOn(t, CantActivate, "Test Planeswalker")
	pw := pushLoyaltyWalker(g, me, 4, -3)
	// The walker predates the stub's source, so bring the layer
	// engine up to date the way a real entry would.
	pushRestrictor(g, me)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != ErrCantActivate {
		t.Errorf("catalog loyalty activation = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateLoyalty(me.ID, pw, "manual +1", 1); err != ErrCantActivate {
		t.Errorf("sandbox ActivateLoyalty = %v, want ErrCantActivate", err)
	}
	if got := loyaltyOf(g, pw); got != 4 {
		t.Errorf("loyalty = %d, want 4 — a refused activation must cost nothing", got)
	}
}

// TestAutoTapSkipsAPermanentWhoseManaAbilitiesAreRestricted. The
// auto-tapper PLANS a payment and the activation path executes it,
// so a source the executor will refuse must never be planned — the
// failure mode is a half-paid cost, with whatever came earlier in
// the plan already tapped. Same argument #352 made for gated
// abilities, now for Arrest.
func TestAutoTapSkipsAPermanentWhoseManaAbilitiesAreRestricted(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withRestrictionOn(t, CantActivateMana, "Arrested Bird")
	pushTypedTestCard(g, Card{
		Name: "Arrested Bird", TypeLine: "Creature — Bird",
		Owner: me.ID, Controller: me.ID,
		ManaAbilities: []ManaAbilityShape{{Produced: "{G}", Label: "Add {G}"}},
	})
	pushRestrictor(g, me)

	cost, err := ParseCost("{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if _, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Error("the auto-tapper planned a payment from a source whose mana abilities can't be activated")
	}
}
