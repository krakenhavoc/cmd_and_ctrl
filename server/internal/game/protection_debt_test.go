package game

import (
	"testing"

	"github.com/google/uuid"
)

// protection_debt_test.go — #662, three of the four checks CR 702.16
// makes: Targeting (702.16b), Enchanting/Equipping (702.16c-d) and
// Blocking (702.16f). Damage (702.16e) has its own file, because its
// whole difficulty is the source's last-known information.
//
// Every test here is really one assertion about the SOURCE. The
// controller is deliberately the wrong answer in each of them — a
// white player's red spell, a red player's colourless Equipment — so
// a regression that reads the controller instead of the object fails
// loudly rather than passing by coincidence.

// pushColouredCreature seeds a creature with an explicit colour and
// optional printed keywords. Colours are stamped rather than derived
// from a mana cost, because the whole point of these tests is which
// colour the RULE reads.
func pushColouredCreature(g *Game, owner *Player, name string, colors []string, keywords ...string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Test"
	c.Power, c.Toughness = 2, 2
	c.Colors = colors
	c.Keywords = keywords
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// castColouredSpellAt casts a one-shot instant of the given colour at
// `ref`. The CASTER is `me`; the spell's own colour is what CR
// 702.16b tests, which is why the two are always different in this
// file.
func castColouredSpellAt(t *testing.T, g *Game, me *Player, oracle string, colors []string, ref TargetRef) error {
	t.Helper()
	spell := NewCard("Test Bolt", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	spell.Colors = colors
	me.Hand.PushTop(spell)
	return g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{Targets: []TargetRef{ref}})
}

// --- T: CR 702.16b -------------------------------------------------

// TestProtectionRefusesASpellOfTheQualityAndOnlyThat is the headline:
// a red spell cannot target a pro-red creature, a white one can, and
// the creature's own controller is not a party to the question (which
// is where protection differs from hexproof).
func TestProtectionRefusesASpellOfTheQualityAndOnlyThat(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	const oracle = "test-protection-targeting"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return anyCreatureSpec()
		}
		return nil
	})

	theirs := pushColouredCreature(g, opp, "Their Pro-Red Bear", []string{"W"}, "protection from red")
	mine := pushColouredCreature(g, me, "My Pro-Red Bear", []string{"W"}, "protection from red")

	if err := castColouredSpellAt(t, g, me, oracle, []string{"R"},
		TargetRef{Kind: TargetCard, ID: theirs}); err != ErrIllegalTarget {
		t.Fatalf("red spell at an opponent's pro-red creature: got %v, want ErrIllegalTarget", err)
	}
	// Unlike hexproof, protection does not care whose creature it is:
	// you cannot bolt your OWN pro-red creature either.
	if err := castColouredSpellAt(t, g, me, oracle, []string{"R"},
		TargetRef{Kind: TargetCard, ID: mine}); err != ErrIllegalTarget {
		t.Fatalf("red spell at your own pro-red creature: got %v, want ErrIllegalTarget", err)
	}
	// A white spell is fine. This is the assertion that fails if the
	// check degenerates into "has protection, refuse".
	if err := castColouredSpellAt(t, g, me, oracle, []string{"W"},
		TargetRef{Kind: TargetCard, ID: theirs}); err != nil {
		t.Fatalf("white spell at a pro-red creature must be legal: %v", err)
	}
}

// TestProtectionReadsTheAbilitysSourceNotItsController is the case
// ADR 0038 §7 said could not be written: an ABILITY's quality is its
// SOURCE permanent's, and the source is neither the caster nor a
// spell. A red player activating a colourless artifact may point it
// at a pro-red creature; a red permanent's ability may not.
func TestProtectionReadsTheAbilitysSourceNotItsController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)

	victim := pushColouredCreature(g, opp, "Pro-Red Bear", []string{"W"}, "protection from red")

	pinger := func(name string, colors []string) *Card {
		c := NewCard(name, me.ID)
		c.TypeLine = "Artifact"
		c.OracleID = "test-protection-ability-" + name
		c.Colors = colors
		g.Battlefield.PushTop(c)
		idx := findCardOnBattlefield(g, c.InstanceID)
		return &g.Battlefield.Cards[idx]
	}
	colourless := pinger("Colourless Pinger", nil)
	red := pinger("Red Pinger", []string{"R"})

	var colourlessOK, redOK bool
	g.ReadSnapshot(func() {
		spec := anyCreatureSpec()
		ref := TargetRef{Kind: TargetCard, ID: victim}
		colourlessOK = g.targetLegalLocked(SourceObject(me.ID, colourless), spec, ref)
		redOK = g.targetLegalLocked(SourceObject(me.ID, red), spec, ref)
	})
	if !colourlessOK {
		t.Error("a COLOURLESS permanent's ability may target a pro-red creature, whoever controls it")
	}
	if redOK {
		t.Error("a RED permanent's ability may not (CR 702.16b tests the source object)")
	}
}

// TestProtectionGainedInResponseFizzlesTheSpell is CR 608.2b through
// the one function announce and resolution share: the spell was legal
// when it was cast and is not when it resolves.
func TestProtectionGainedInResponseFizzlesTheSpell(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	const oracle = "test-protection-recheck"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return anyCreatureSpec()
		}
		return nil
	})

	victim := pushColouredCreature(g, opp, "Plain Bear", []string{"G"})
	if err := castColouredSpellAt(t, g, me, oracle, []string{"R"},
		TargetRef{Kind: TargetCard, ID: victim}); err != nil {
		t.Fatalf("announce: %v", err)
	}

	var item *StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			if len(it.Targets) == 1 && it.Targets[0].ID == victim {
				item = it
			}
		}
	})
	if item == nil {
		t.Fatal("the spell is not on the stack")
	}

	var legalBefore, legalAfter bool
	g.WithWriteLock(func() {
		legalBefore = g.TargetStillLegalForEffect(item, item.Targets[0])
		idx := findCardOnBattlefield(g, victim)
		g.Battlefield.Cards[idx].Keywords = append(g.Battlefield.Cards[idx].Keywords, "protection from red")
		legalAfter = g.TargetStillLegalForEffect(item, item.Targets[0])
	})
	if !legalBefore {
		t.Error("the target was legal at announce and must still be before the grant")
	}
	if legalAfter {
		t.Error("CR 608.2b: a target that gained protection from the spell's colour is no longer legal")
	}
}

// TestPayingACostIsNotTargeting is ADR 0038 §4 holding for protection
// too. Convoking, tapping or sacrificing a pro-red creature is legal
// however red the thing being paid for is — the source-less
// TargetSource the cost paths pass is a DECLARATION, not a gap.
func TestPayingACostIsNotTargeting(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mine := pushColouredCreature(g, me, "My Pro-Red Bear", []string{"W"}, "protection from red")

	spec := anyCreatureSpec()
	ref := TargetRef{Kind: TargetCard, ID: mine}
	var asCost, asTarget bool
	g.ReadSnapshot(func() {
		red := &Characteristic{Colors: []string{"R"}}
		asCost = g.specMatchLocked(SourceChooser(me.ID), spec, ref, false)
		asTarget = g.specMatchLocked(SourceSnapshot(me.ID, red), spec, ref, true)
	})
	if !asCost {
		t.Error("paying a cost is not targeting (CR 601.2h): a pro-red creature is still convokable")
	}
	if asTarget {
		t.Error("the same walk WITH targeting on must refuse it")
	}
}

// --- E: CR 702.16c-d -----------------------------------------------

// TestARedAuraFallsOffAndARedEquipmentUnattaches is both halves of
// the attachment rule and both halves of its SBA: CR 704.5m puts the
// Aura in a graveyard, CR 704.5n merely unattaches the Equipment.
//
// The check has to run BEFORE attachmentLegalLocked's catalogued-Aura
// branch, which would answer the Aura's own "target creature" clause
// and return happy.
func TestARedAuraFallsOffAndARedEquipmentUnattaches(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	aura := pushAttachTestCard(g, me.ID, "Red Aura", "Enchantment — Aura")
	sword := pushAttachTestCard(g, me.ID, "Red Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{aura, sword} {
			idx := findCardOnBattlefield(g, id)
			g.Battlefield.Cards[idx].Colors = []string{"R"}
			if err := g.AttachForEffect(id, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
				t.Fatalf("attach %v: %v", id, err)
			}
		}
		g.runStateChecksLocked()
	})
	if c, ok := battlefieldCardByID(g, aura); !ok || !c.IsAttachedTo(bear) {
		t.Fatal("both attachments are legal before the grant")
	}

	g.WithWriteLock(func() {
		idx := findCardOnBattlefield(g, bear)
		g.Battlefield.Cards[idx].Keywords = append(g.Battlefield.Cards[idx].Keywords, "protection from red")
		// The layer cache was filled when the creature entered, and
		// protection is read off it like every other keyword.
		g.recomputeLayersLocked()
		g.runStateChecksLocked()
	})

	if _, ok := battlefieldCardByID(g, aura); ok {
		t.Error("CR 704.5m: a red Aura on a pro-red creature goes to its owner's graveyard")
	}
	if !graveyardHas(me, aura) {
		t.Error("the Aura is not in the graveyard either")
	}
	swordNow, ok := battlefieldCardByID(g, sword)
	if !ok {
		t.Fatal("CR 704.5n unattaches an Equipment; it must NOT go to a graveyard")
	}
	if swordNow.IsAttached() {
		t.Error("the red Equipment is still attached to a pro-red creature")
	}
}

// A WHITE Aura on a pro-red creature stays put. Without this the test
// above passes for a check that simply drops every attachment off a
// creature with any protection at all.
func TestAnAuraWithoutTheQualityStaysAttached(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	aura := pushAttachTestCard(g, me.ID, "White Aura", "Enchantment — Aura")
	g.WithWriteLock(func() {
		bIdx := findCardOnBattlefield(g, bear)
		g.Battlefield.Cards[bIdx].Keywords = []string{"protection from red"}
		g.recomputeLayersLocked()
		aIdx := findCardOnBattlefield(g, aura)
		g.Battlefield.Cards[aIdx].Colors = []string{"W"}
		if err := g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
			t.Fatalf("attach: %v", err)
		}
		g.runStateChecksLocked()
	})
	c, ok := battlefieldCardByID(g, aura)
	if !ok || !c.IsAttachedTo(bear) {
		t.Error("a white Aura is unaffected by protection from red")
	}
}

// --- B: CR 702.16f -------------------------------------------------

// TestARedCreatureCannotBlockAProRedAttacker fills the slot ADR 0045's
// addendum reserved, and pins the asymmetry with it: protection on the
// ATTACKER refuses the block, protection on the BLOCKER refuses
// nothing.
func TestARedCreatureCannotBlockAProRedAttacker(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	attacker := pushColouredCreature(g, me, "Pro-Red Attacker", []string{"W"}, "protection from red")
	redBlocker := pushColouredCreature(g, opp, "Red Blocker", []string{"R"})
	whiteBlocker := pushColouredCreature(g, opp, "White Blocker", []string{"W"})
	// The mirror: a pro-red BLOCKER against a red attacker.
	redAttacker := pushColouredCreature(g, me, "Red Attacker", []string{"R"})
	proRedBlocker := pushColouredCreature(g, opp, "Pro-Red Blocker", []string{"W"}, "protection from red")

	card := func(id uuid.UUID) *Card {
		idx := findCardOnBattlefield(g, id)
		return &g.Battlefield.Cards[idx]
	}
	var refusal, white, mirror BlockRefusal
	g.ReadSnapshot(func() {
		refusal = g.BlockPairRefusalLocked(card(attacker), card(redBlocker))
		white = g.BlockPairRefusalLocked(card(attacker), card(whiteBlocker))
		mirror = g.BlockPairRefusalLocked(card(redAttacker), card(proRedBlocker))
	})
	if refusal.Reason != BlockReasonProtection {
		t.Errorf("a red blocker against a pro-red attacker: got %q, want protection", refusal.Reason)
	}
	if refusal.Source != attacker {
		t.Error("the refusal names the creature carrying the ability")
	}
	if !white.Legal() {
		t.Errorf("a white blocker is legal: got %q", white.Reason)
	}
	if !mirror.Legal() {
		t.Errorf("protection does not stop you BLOCKING (CR 702.16e handles the damage): got %q", mirror.Reason)
	}
}

// The refusal reaches the player as a sentence that NAMES the
// quality. "That creature can't block" is not enough to work out
// which of your creatures could have.
func TestTheBlockRefusalNamesTheQuality(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	attacker := pushColouredCreature(g, me, "Baneslayer Angel", []string{"W"}, "protection from Demons")
	blocker := pushColouredCreature(g, opp, "Shivan Dragon", nil)
	g.WithWriteLock(func() {
		idx := findCardOnBattlefield(g, blocker)
		g.Battlefield.Cards[idx].TypeLine = "Creature — Demon"
	})

	var sentence string
	g.ReadSnapshot(func() {
		aIdx, bIdx := findCardOnBattlefield(g, attacker), findCardOnBattlefield(g, blocker)
		a, b := &g.Battlefield.Cards[aIdx], &g.Battlefield.Cards[bIdx]
		err := g.blockRefusedErrorLocked(a, b, g.BlockPairRefusalLocked(a, b))
		sentence = err.Sentence(opp.ID)
	})
	want := "Baneslayer Angel has protection from Demons, so Shivan Dragon can't block it."
	if sentence != want {
		t.Errorf("sentence = %q, want %q", sentence, want)
	}
}
