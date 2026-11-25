package game

import (
	"testing"

	"github.com/google/uuid"
)

// source_targeting_bar_test.go — #6196: "can't be the target of" a
// spell or ability WITH A QUALITY (Bartel Runeaxe's Aura spells, Raiding
// Party's white spells and abilities from white sources, Artifact Ward's
// abilities from artifact sources, Anti-Magic Aura's spells, Lurker's
// until it attacked or blocked). CR 702.16b is the matcher the bar
// borrows; the cards are pinned in cards/effects when their batches
// build them.

// barredCreature puts a creature carrying `bars` on the battlefield and
// returns a value copy, as the targeting gate is asked about one.
func barredCreature(t *testing.T, g *Game, owner uuid.UUID, bars ...SourceTargetingBar) Card {
	t.Helper()
	id := pushScopedTestCreature(g, owner, 2, 2)
	var out Card
	g.WithWriteLock(func() {
		c, ok := g.battlefieldCardLocked(id)
		if !ok {
			t.Fatalf("card %s not on the battlefield", id)
		}
		eff := c.Effective()
		eff.TargetingBars = append(eff.TargetingBars, bars...)
		c.effective = &eff
		out = *c
	})
	return out
}

// auraSpell is an Aura enchantment spell's source, whiteInstant a white
// instant's, and artifactAbility an artifact's activated ability.
func auraSpell(controller uuid.UUID) TargetSource {
	return SourceCast(controller, &Card{
		InstanceID: uuid.New(), Name: "Test Aura", TypeLine: "Enchantment — Aura",
		Controller: controller, Owner: controller,
	})
}

func whiteInstant(controller uuid.UUID) TargetSource {
	return SourceCast(controller, &Card{
		InstanceID: uuid.New(), Name: "Test Instant", TypeLine: "Instant",
		ManaCost: "{W}", Controller: controller, Owner: controller,
	})
}

func whiteAbility(controller uuid.UUID) TargetSource {
	return SourceObject(controller, &Card{
		InstanceID: uuid.New(), Name: "White Source", TypeLine: "Creature — Human",
		ManaCost: "{W}", Controller: controller, Owner: controller,
	})
}

func artifactAbility(controller uuid.UUID) TargetSource {
	return SourceObject(controller, &Card{
		InstanceID: uuid.New(), Name: "Test Artifact", TypeLine: "Artifact",
		Controller: controller, Owner: controller,
	})
}

// TestBarOnAuraSpellsRefusesAnAuraAndNothingElse — Bartel Runeaxe: an
// Aura spell is refused, an instant is not, and an Aura's own activated
// ability (an ability, not a spell) is not.
func TestBarOnAuraSpellsRefusesAnAuraAndNothingElse(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	card := barredCreature(t, g, g.Seats[1].ID, SourceTargetingBar{
		Label:   "Bartel Runeaxe can't be the target of Aura spells.",
		Quality: ProtectionQuality{Kind: ProtectionQualitySubtype, Value: "Aura", Printed: "Aura"},
		Spells:  true,
	})
	g.ReadSnapshot(func() {
		if g.canBeTargetedByLocked(&card, ZoneBattlefield, auraSpell(me)) {
			t.Error("an Aura spell targeted it")
		}
		if !g.canBeTargetedByLocked(&card, ZoneBattlefield, whiteInstant(me)) {
			t.Error("an instant was refused")
		}
		if !g.canBeTargetedByLocked(&card, ZoneBattlefield, SourceObject(me, &Card{
			InstanceID: uuid.New(), TypeLine: "Enchantment — Aura", Controller: me, Owner: me,
		})) {
			t.Error("an Aura's ABILITY was refused by a bar on Aura spells")
		}
	})
}

// TestBarOnAColourRefusesSpellsAndAbilities — Raiding Party: "white
// spells or abilities from white sources" is CR 702.16b's pair, so both
// halves are refused and a colourless source passes.
func TestBarOnAColourRefusesSpellsAndAbilities(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	card := barredCreature(t, g, g.Seats[1].ID, SourceTargetingBar{
		Label:     "This enchantment can't be the target of white spells or abilities from white sources.",
		Quality:   ProtectionQuality{Kind: ProtectionQualityColor, Value: "W", Printed: "white"},
		Spells:    true,
		Abilities: true,
	})
	g.ReadSnapshot(func() {
		if g.canBeTargetedByLocked(&card, ZoneBattlefield, whiteInstant(me)) {
			t.Error("a white spell targeted it")
		}
		if g.canBeTargetedByLocked(&card, ZoneBattlefield, whiteAbility(me)) {
			t.Error("an ability from a white source targeted it")
		}
		if !g.canBeTargetedByLocked(&card, ZoneBattlefield, artifactAbility(me)) {
			t.Error("a colourless artifact's ability was refused")
		}
	})
}

// TestBarOnAbilitiesLeavesSpells — Artifact Ward: "abilities from
// artifact sources" says abilities, so an artifact SPELL still targets.
func TestBarOnAbilitiesLeavesSpells(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	card := barredCreature(t, g, g.Seats[1].ID, SourceTargetingBar{
		Label:     "Enchanted creature can't be the target of abilities from artifact sources.",
		Quality:   ProtectionQuality{Kind: ProtectionQualityCardType, Value: "artifact", Printed: "artifact"},
		Abilities: true,
	})
	g.ReadSnapshot(func() {
		if g.canBeTargetedByLocked(&card, ZoneBattlefield, artifactAbility(me)) {
			t.Error("an ability from an artifact source targeted it")
		}
		if !g.canBeTargetedByLocked(&card, ZoneBattlefield, SourceCast(me, &Card{
			InstanceID: uuid.New(), TypeLine: "Artifact", Controller: me, Owner: me,
		})) {
			t.Error("an artifact SPELL was refused by a bar on abilities")
		}
	})
}

// TestBarOnEverySpell — Anti-Magic Aura: every spell is refused, every
// ability passes.
func TestBarOnEverySpell(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	card := barredCreature(t, g, g.Seats[1].ID, SourceTargetingBar{
		Label:   "Enchanted creature can't be the target of spells.",
		Quality: ProtectionQuality{Kind: ProtectionQualityEverything, Printed: "everything"},
		Spells:  true,
	})
	g.ReadSnapshot(func() {
		if g.canBeTargetedByLocked(&card, ZoneBattlefield, whiteInstant(me)) {
			t.Error("a spell targeted it")
		}
		if !g.canBeTargetedByLocked(&card, ZoneBattlefield, artifactAbility(me)) {
			t.Error("an ability was refused by a bar on spells")
		}
	})
}

// TestLurkerIsTargetableOnceItAttacked — the condition: barred while it
// has not attacked or blocked this turn, targetable once its attack is
// locked in, and still targetable after combat, because the declaration
// is history (the turn's record, not the creature's combat state).
func TestLurkerIsTargetableOnceItAttacked(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bar := SourceTargetingBar{
		Label:                           "Lurker can't be the target of spells unless it attacked or blocked this turn.",
		Quality:                         ProtectionQuality{Kind: ProtectionQualityEverything, Printed: "everything"},
		Spells:                          true,
		UnlessAttackedOrBlockedThisTurn: true,
	}
	lurker := pushCombatant(t, g, me, "Lurker", 2, 2)
	// The bar is stamped on the effective characteristic before each
	// read: a layer recompute rebuilds it from the catalog, and the
	// catalog cannot be reached from this package.
	barred := func() bool {
		var ok bool
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			c, found := g.battlefieldCardLocked(lurker)
			if !found {
				t.Fatal("the Lurker left the battlefield")
			}
			eff := c.Effective()
			eff.TargetingBars = append(eff.TargetingBars, bar)
			c.effective = &eff
			card := *c
			ok = !g.canBeTargetedByLocked(&card, ZoneBattlefield, whiteInstant(them.ID))
		})
		return ok
	}
	if !barred() {
		t.Fatal("a Lurker that has not attacked was targetable")
	}
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(lurker, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// CR 508.1: the attack is RECORDED as the declaration is locked in,
	// which is where "attacked this turn" starts being true.
	lockInAttacksByPriorityWrap(t, g)
	if barred() {
		t.Fatal("a Lurker that attacked is still barred")
	}
	// And it stays targetable once combat is over.
	advanceIntoStep(t, g, StepPostcombatMain)
	if barred() {
		t.Error("a Lurker that attacked earlier this turn is barred again after combat")
	}
}

// TestLurkerIsTargetableOnceItBlocked — the other half of the
// condition, and the reason it is read off the turn's per-pair block
// record (#6193) rather than the creature's current block: a Lurker that
// blocked stays targetable once combat is over, which is what "blocked
// this turn" says.
func TestLurkerIsTargetableOnceItBlocked(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bar := SourceTargetingBar{
		Label:                           "Lurker can't be the target of spells unless it attacked or blocked this turn.",
		Quality:                         ProtectionQuality{Kind: ProtectionQualityEverything, Printed: "everything"},
		Spells:                          true,
		UnlessAttackedOrBlockedThisTurn: true,
	}
	// A small attacker and a fat Lurker, so the Lurker is still on the
	// battlefield after combat damage and the bar can be asked again.
	attacker := pushCombatant(t, g, me, "Attacker", 1, 1)
	lurker := pushCombatant(t, g, them, "Lurker", 2, 5)
	barred := func() bool {
		var ok bool
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			c, found := g.battlefieldCardLocked(lurker)
			if !found {
				t.Fatal("the Lurker left the battlefield")
			}
			eff := c.Effective()
			eff.TargetingBars = append(eff.TargetingBars, bar)
			c.effective = &eff
			card := *c
			ok = !g.canBeTargetedByLocked(&card, ZoneBattlefield, whiteInstant(me.ID))
		})
		return ok
	}
	if !barred() {
		t.Fatal("a Lurker that has not blocked was targetable")
	}
	attackAndStepToBlocks(t, g, them.ID, attacker)
	if err := g.DeclareBlocker(lurker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	// A STAGED block is not yet a block: the declaration is announced
	// when the defender finishes it (CR 509.1), and that is where the
	// turn records it.
	if !barred() {
		t.Fatal("a staged block lifted the bar before the declaration was announced")
	}
	if err := g.FinishBlocks(them.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	if barred() {
		t.Fatal("a Lurker that blocked is still barred")
	}
	advanceIntoStep(t, g, StepPostcombatMain)
	if barred() {
		t.Error("a Lurker that blocked earlier this turn is barred again after combat")
	}
}

// TestBarIsNotHexproofSoAWaiverDoesNotLiftIt — the bars sit ahead of
// hexproof, so a "may be targeted as though it didn't have hexproof"
// waiver (hexproof_bypass.go) cannot reach them. Here as a unit on the
// predicate: a bar with neither half named is inert rather than total.
func TestBarWithNoHalfNamedIsInert(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	card := barredCreature(t, g, g.Seats[1].ID, SourceTargetingBar{
		Label:   "a card-file bug: no half named",
		Quality: ProtectionQuality{Kind: ProtectionQualityEverything},
	})
	g.ReadSnapshot(func() {
		if !g.canBeTargetedByLocked(&card, ZoneBattlefield, whiteInstant(me)) {
			t.Error("a bar naming neither spells nor abilities barred a spell")
		}
		if !g.canBeTargetedByLocked(&card, ZoneBattlefield, artifactAbility(me)) {
			t.Error("a bar naming neither spells nor abilities barred an ability")
		}
	})
}
