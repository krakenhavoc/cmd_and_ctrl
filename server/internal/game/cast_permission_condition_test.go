package game

import (
	"testing"

	"github.com/google/uuid"
)

// cast_permission_condition_test.go — #3192, ADR 0066's second
// 2026-10-13 amendment. The MODEL of a conditional exile permission
// against hand-built boards; the cards that use it are pinned in
// cards/effects.

// conditionalExile exiles the top card of `owner`'s library and grants
// `holder` a permission over it that is live only while cond holds.
func conditionalExile(t *testing.T, g *Game, owner, holder *Player, cond *PermissionCondition) uuid.UUID {
	t.Helper()
	seedLibraryTop(owner, "Held Bolt", "Instant")
	return impulseExile(t, g, owner, holder, CastPermission{
		Duration:  WhileInZoneDuration(),
		Condition: cond,
	})
}

// playableNow is the engine's own answer: the permission query the cast
// path, the view and the enumerator all read.
func playableNow(g *Game, seat *Player, id uuid.UUID) bool {
	return grantOn(g, seat.ID, id, ZoneExile) != nil
}

// declareAttack declares `ids` as attackers against the first opponent
// and locks the declaration in, which is what emits EventAttack.
func condDeclareAttack(t *testing.T, g *Game, ids ...uuid.UUID) {
	t.Helper()
	for _, id := range ids {
		if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttackDeclaration(t, g)
}

// retype gives a battlefield test creature a type line and re-derives
// its cached characteristics from it, as the layer pass would.
func retype(g *Game, id uuid.UUID, typeLine string) {
	c := findCard(g, id)
	c.TypeLine = typeLine
	c.effective = printedEffectiveWith(*c)
}

func attackReady(t *testing.T) (*Game, *Player) {
	t.Helper()
	g := newFourPlayerActiveGame(t)
	return g, g.Seats[0]
}

// The permission persists across the turns the condition is false: the
// grant is there, the card is exiled, and the card is simply not
// playable until the condition holds.
func TestConditionalPermissionIsDormantNotGone(t *testing.T) {
	g, me := attackReady(t)
	tok := pushPlainCreature(t, g, me, "Goblin Token")
	g.WithWriteLock(func() { retype(g, tok, "Token Creature — Goblin") })
	toMainPhase(t, g)
	id := conditionalExile(t, g, me, me, &PermissionCondition{Key: CondAttackedWithToken})

	if playableNow(g, me, id) {
		t.Fatal("playable before any token attacked")
	}
	var unmet bool
	g.WithWriteLock(func() {
		perm := g.CastPermissionOnCardByIDForEffect(id)
		if perm == nil {
			t.Fatal("the dormant grant disappeared; it should persist")
		}
		unmet = g.PermissionConditionUnmetLocked(perm)
		if PermissionConditionLabel(perm) == "" {
			t.Error("a conditional grant has no label for the client")
		}
	})
	if !unmet {
		t.Error("the view's condition_unmet is false while the condition is false")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile"}); err != ErrNoPlayPermission {
		t.Fatalf("cast while the condition is false = %v, want ErrNoPlayPermission", err)
	}

	advanceIntoStep(t, g, StepDeclareAttackers)
	condDeclareAttack(t, g, tok)
	if !playableNow(g, me, id) {
		t.Fatal("not playable the turn a token attacked")
	}
	g.WithWriteLock(func() {
		if g.PermissionConditionUnmetLocked(g.CastPermissionOnCardByIDForEffect(id)) {
			t.Error("condition_unmet still true once the condition holds")
		}
	})
}

// The tally is per turn: the same card is playable on a later turn only
// if the condition holds again on that turn.
func TestConditionalPermissionFlipsAcrossTurns(t *testing.T) {
	g, me := attackReady(t)
	tok := pushPlainCreature(t, g, me, "Goblin Token")
	g.WithWriteLock(func() { retype(g, tok, "Token Creature — Goblin") })
	toMainPhase(t, g)
	id := conditionalExile(t, g, me, me, &PermissionCondition{Key: CondAttackedWithToken})
	advanceIntoStep(t, g, StepDeclareAttackers)
	condDeclareAttack(t, g, tok)
	if !playableNow(g, me, id) {
		t.Fatal("setup: not playable the turn a token attacked")
	}

	// The token dies in combat. The answer to "did I attack with a token
	// this turn" must not.
	removeFromBattlefieldForTest(t, g, tok, false)
	if !playableNow(g, me, id) {
		t.Error("the condition lapsed when the token that attacked died")
	}

	// A later turn with no attack: dormant again. Reset the tally the
	// way the turn boundary does.
	g.WithWriteLock(func() { g.resetTurnTallyLocked() })
	if playableNow(g, me, id) {
		t.Error("playable on a later turn on which no token attacked")
	}
	g.WithWriteLock(func() {
		if g.CastPermissionOnCardByIDForEffect(id) == nil {
			t.Error("the grant was dropped at the turn boundary")
		}
	})
}

func TestConditionKindsReadWhatTheyClaim(t *testing.T) {
	setup := func(t *testing.T) (*Game, *Player, uuid.UUID, uuid.UUID, uuid.UUID) {
		g, me := attackReady(t)
		a := pushPlainCreature(t, g, me, "Alpha")
		b := pushPlainCreature(t, g, me, "Bravo")
		c := pushPlainCreature(t, g, me, "Charlie")
		toMainPhase(t, g)
		return g, me, a, b, c
	}

	t.Run("a commander", func(t *testing.T) {
		g, me, a, b, _ := setup(t)
		id := conditionalExile(t, g, me, me, &PermissionCondition{Key: CondAttackedWithCommander})
		advanceIntoStep(t, g, StepDeclareAttackers)
		condDeclareAttack(t, g, a)
		if playableNow(g, me, id) {
			t.Fatal("a non-commander attacked and opened it")
		}
		g.WithWriteLock(func() { findCard(g, b).IsCommander = true })
		condDeclareAttack(t, g, b)
		if !playableNow(g, me, id) {
			t.Error("a commander attacked and it stayed shut")
		}
	})

	t.Run("a creature type, changelings included", func(t *testing.T) {
		g, me, a, b, _ := setup(t)
		g.WithWriteLock(func() {
			retype(g, a, "Creature — Human Rogue")
			retype(g, b, "Creature — Elf")
		})
		id := conditionalExile(t, g, me, me, &PermissionCondition{Key: CondAttackedWithSubtype, Subtype: "Rogue"})
		advanceIntoStep(t, g, StepDeclareAttackers)
		condDeclareAttack(t, g, b)
		if playableNow(g, me, id) {
			t.Fatal("an Elf attacked and opened a Rogue condition")
		}
		condDeclareAttack(t, g, a)
		if !playableNow(g, me, id) {
			t.Error("a Rogue attacked and it stayed shut")
		}
	})

	t.Run("N or more different creatures", func(t *testing.T) {
		g, me, a, b, c := setup(t)
		id := conditionalExile(t, g, me, me, &PermissionCondition{Key: CondAttackedWithAtLeast, N: 3})
		advanceIntoStep(t, g, StepDeclareAttackers)
		condDeclareAttack(t, g, a, b)
		if playableNow(g, me, id) {
			t.Fatal("two attackers opened a three-attacker condition")
		}
		removeFromBattlefieldForTest(t, g, b, false)
		// A creature re-declared in a second combat is still one creature.
		g.WithWriteLock(func() {
			g.EmitEvent(Event{Kind: EventAttack, Actor: me.ID, CardID: a, Target: g.Seats[1].ID})
		})
		if playableNow(g, me, id) {
			t.Fatal("a repeated declaration of one creature counted twice")
		}
		condDeclareAttack(t, g, c)
		if !playableNow(g, me, id) {
			t.Error("three different creatures attacked and it stayed shut")
		}
	})

	t.Run("one specific object", func(t *testing.T) {
		g, me, a, b, _ := setup(t)
		var epoch int
		g.WithWriteLock(func() { epoch = findCard(g, a).ObjectEpoch })
		id := conditionalExile(t, g, me, me, &PermissionCondition{Key: CondAttackedWithObject, ObjectID: a, ObjectEpoch: epoch})
		advanceIntoStep(t, g, StepDeclareAttackers)
		condDeclareAttack(t, g, b)
		if playableNow(g, me, id) {
			t.Fatal("another creature attacked and opened an object condition")
		}
		condDeclareAttack(t, g, a)
		if !playableNow(g, me, id) {
			t.Error("the named creature attacked and it stayed shut")
		}
	})

	t.Run("a different epoch is a different object", func(t *testing.T) {
		g, me, a, _, _ := setup(t)
		var epoch int
		g.WithWriteLock(func() { epoch = findCard(g, a).ObjectEpoch })
		id := conditionalExile(t, g, me, me, &PermissionCondition{Key: CondAttackedWithObject, ObjectID: a, ObjectEpoch: epoch + 1})
		advanceIntoStep(t, g, StepDeclareAttackers)
		condDeclareAttack(t, g, a)
		if playableNow(g, me, id) {
			t.Error("a creature of another epoch satisfied the object condition (CR 400.7)")
		}
	})

	t.Run("any attack", func(t *testing.T) {
		g, me, a, _, _ := setup(t)
		id := conditionalExile(t, g, me, me, &PermissionCondition{Key: CondAttackedThisTurn})
		if playableNow(g, me, id) {
			t.Fatal("open before anything attacked")
		}
		advanceIntoStep(t, g, StepDeclareAttackers)
		condDeclareAttack(t, g, a)
		if !playableNow(g, me, id) {
			t.Error("an attack happened and it stayed shut")
		}
	})

	t.Run("control a subtype", func(t *testing.T) {
		g, me, a, _, _ := setup(t)
		id := conditionalExile(t, g, me, me, &PermissionCondition{Key: CondControlSubtype, Subtype: "Kavu"})
		if playableNow(g, me, id) {
			t.Fatal("open with no Kavu")
		}
		g.WithWriteLock(func() { retype(g, a, "Creature — Kavu") })
		if !playableNow(g, me, id) {
			t.Error("a Kavu is controlled and it stayed shut")
		}
		g.WithWriteLock(func() { findCard(g, a).Controller = g.Seats[1].ID })
		if playableNow(g, me, id) {
			t.Error("an opponent's Kavu opened it")
		}
	})

	t.Run("a lore counter on a saga", func(t *testing.T) {
		g, me, a, _, _ := setup(t)
		var epoch int
		g.WithWriteLock(func() {
			retype(g, a, "Enchantment Creature — Saga Knight")
			epoch = findCard(g, a).ObjectEpoch
		})
		id := conditionalExile(t, g, me, me, &PermissionCondition{Key: CondLoreCounterOnObject, ObjectID: a, ObjectEpoch: epoch})
		if playableNow(g, me, id) {
			t.Fatal("open before a lore counter")
		}
		g.WithWriteLock(func() { g.loreCountersPutLocked(a, 0, 1) })
		if !playableNow(g, me, id) {
			t.Error("a lore counter went on the saga and it stayed shut")
		}
		g.WithWriteLock(func() { g.resetTurnTallyLocked() })
		if playableNow(g, me, id) {
			t.Error("a lore counter from a previous turn still opens it")
		}
	})
}

// An unregistered key fails closed (dormant), and a snapshot naming one
// is refused rather than restored as an unconditional grant.
func TestUnknownConditionFailsClosed(t *testing.T) {
	g, me := attackReady(t)
	toMainPhase(t, g)
	id := conditionalExile(t, g, me, me, &PermissionCondition{Key: "no_such_condition"})
	if playableNow(g, me, id) {
		t.Fatal("an unknown condition opened the permission")
	}
	if _, err := g.CaptureSnapshot().Restore(); err == nil {
		t.Error("restore accepted a permission whose condition this binary cannot read")
	}
}

// Clone and snapshot carry a pending conditional permission AND the
// tally it reads, so an undo or a server restart mid-turn neither opens
// nor shuts a card.
func TestConditionalPermissionSurvivesCloneAndSnapshot(t *testing.T) {
	g, me := attackReady(t)
	tok := pushPlainCreature(t, g, me, "Goblin Token")
	g.WithWriteLock(func() { retype(g, tok, "Token Creature — Goblin") })
	toMainPhase(t, g)
	id := conditionalExile(t, g, me, me, &PermissionCondition{Key: CondAttackedWithToken})

	clone := g.Clone()
	if playableNow(clone, me, id) {
		t.Fatal("the clone opened a dormant permission")
	}
	snap := g.CaptureSnapshot()
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if playableNow(restored, me, id) {
		t.Fatal("the restore opened a dormant permission")
	}
	if !restored.Seats[0].CastPermissions[0].equalCondition(&PermissionCondition{Key: CondAttackedWithToken}) {
		t.Fatalf("the condition did not survive the snapshot: %+v", restored.Seats[0].CastPermissions[0].Condition)
	}

	advanceIntoStep(t, g, StepDeclareAttackers)
	condDeclareAttack(t, g, tok)
	// Independent backing: a clone taken now carries the open state, and
	// resetting the original's tally must not reach it.
	live := g.Clone()
	liveSnap, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore after the attack: %v", err)
	}
	g.WithWriteLock(func() { g.resetTurnTallyLocked() })
	if !playableNow(live, me, id) {
		t.Error("the clone lost the attack tally")
	}
	if !playableNow(liveSnap, me, id) {
		t.Error("the restore lost the attack tally")
	}
	if playableNow(g, me, id) {
		t.Error("setup: the reset did not close the original")
	}
}

func (p CastPermission) equalCondition(c *PermissionCondition) bool {
	return p.Condition != nil && c != nil && *p.Condition == *c
}

// The spell-type gate is the permission's Filter, applied to a named
// card too: "you may cast it if it's a creature spell".
func TestScopeCardsPermissionHonoursItsFilter(t *testing.T) {
	g, me := attackReady(t)
	toMainPhase(t, g)
	creature := seedLibraryTop(me, "Held Bear", "Creature — Bear")
	g.WithWriteLock(func() {
		_, err := g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, CastPermission{
			Duration: WhileInZoneDuration(), Filter: PermissionFilter{CreatureOnly: true},
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	instant := seedLibraryTop(me, "Held Bolt", "Instant")
	g.WithWriteLock(func() {
		_, err := g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, CastPermission{
			Duration: WhileInZoneDuration(), Filter: PermissionFilter{CreatureOnly: true},
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	if !playableNow(g, me, creature) {
		t.Error("the creature card is not castable under a creature-only grant")
	}
	if playableNow(g, me, instant) {
		t.Error("the instant is castable under a creature-only grant")
	}
}

// "For as long as you control <source>" is already a source-bound
// duration on the permission; this pins that it reaches a stored grant
// so the cards that print it need no condition.
func TestForAsLongAsYouControlDurationEndsAStoredGrant(t *testing.T) {
	g, me := attackReady(t)
	src := pushPlainCreature(t, g, me, "Lightning")
	toMainPhase(t, g)
	var d Duration
	var ok bool
	g.WithWriteLock(func() { d, ok = g.ForAsLongAsYouControlDuration(src, me.ID) })
	if !ok {
		t.Fatal("setup: no duration")
	}
	seedLibraryTop(me, "Held Card", "Sorcery")
	var id uuid.UUID
	g.WithWriteLock(func() {
		ids, err := g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, CastPermission{Duration: d})
		if err != nil || len(ids) != 1 {
			t.Fatalf("exile: %v %v", ids, err)
		}
		id = ids[0]
	})
	if !playableNow(g, me, id) {
		t.Fatal("not playable while its source is controlled")
	}
	removeFromBattlefieldForTest(t, g, src, false)
	if playableNow(g, me, id) {
		t.Error("still playable after the source left the battlefield")
	}
}
