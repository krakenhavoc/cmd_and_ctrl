package game

import (
	"testing"

	"github.com/google/uuid"
)

// protection_mana_value_test.go — #2181 ("protection from mana value N
// or less", Reaver Titan) and #2145 ("protection from Ring-bearers",
// Lord of the Nazgul): the two qualities that are not characteristics
// a layer computes. Each is asserted through all four DEBT checks
// (CR 702.16b-f), because the value of the work is that every reader
// sees the new source facts, not that the parser accepts a string.

const proMV3 = "protection from mana value 3 or less"

func TestProtectionGrammarParsesManaValueAndRingBearers(t *testing.T) {
	q, ok := ParseProtectionQuality(proMV3)
	if !ok || q.Kind != ProtectionQualityManaValueAtMost || q.Value != "3" || q.Printed != "mana value 3 or less" {
		t.Fatalf("parse = (%+v, %v)", q, ok)
	}
	if q.Kind.String() != "mana_value_at_most" {
		t.Errorf("wire token = %q", q.Kind.String())
	}
	if q, ok := ParseProtectionQuality("Protection from Mana Value 0 or less"); !ok || q.Value != "0" {
		t.Errorf("case-insensitive, zero bound: (%+v, %v)", q, ok)
	}
	rb, ok := ParseProtectionQuality("protection from Ring-bearers")
	if !ok || rb.Kind != ProtectionQualityRingBearer || rb.Kind.String() != "ring_bearer" {
		t.Fatalf("ring-bearers parse = (%+v, %v)", rb, ok)
	}
	toks, ok := ProtectionTokens("Protection from mana value 3 or less")
	if !ok || len(toks) != 1 || toks[0] != proMV3 {
		t.Errorf("ProtectionTokens = %v, %v", toks, ok)
	}
	// Shapes outside the closed grammar mint nothing.
	for _, bad := range []string{
		"protection from mana value 3 or greater",
		"protection from mana value 3 or more",
		"protection from mana value or less",
		"protection from mana value X or less",
		"protection from mana value 3",
		"protection from Ring-bearer's",
	} {
		if _, ok := ParseProtectionQuality(bad); ok {
			t.Errorf("%q parsed; the grammar is closed", bad)
		}
	}
}

// mvSource is a card on the battlefield with the given printed cost.
func pushMVCreature(g *Game, owner *Player, name, cost string, keywords ...string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Test"
	c.Power, c.Toughness = 2, 2
	c.ManaCost = cost
	c.Keywords = keywords
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func TestManaValueProtectionRefusesBlocksAtOrBelowTheBound(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	attacker := pushMVCreature(g, me, "Void Shields", "{7}", proMV3)
	cases := []struct {
		name  string
		cost  string
		token bool
		legal bool
	}{
		{"mana value 1", "{G}", false, false},
		{"mana value 3", "{1}{G}{G}", false, false},
		{"mana value 4", "{2}{G}{G}", false, true},
		{"an uncosted token is mana value 0", "", true, false},
		{"a token copy has the copiable cost's value", "{5}", true, true},
	}
	card := func(id uuid.UUID) *Card { return &g.Battlefield.Cards[findCardOnBattlefield(g, id)] }
	for _, tc := range cases {
		id := pushMVCreature(g, opp, tc.name, tc.cost)
		if tc.token {
			g.WithWriteLock(func() { card(id).TypeLine = "Token Creature — Test" })
		}
		var r BlockRefusal
		g.ReadSnapshot(func() { r = g.BlockPairRefusalLocked(card(attacker), card(id)) })
		if tc.legal && !r.Legal() {
			t.Errorf("%s: block refused (%q), want legal", tc.name, r.Reason)
		}
		if !tc.legal && r.Reason != BlockReasonProtection {
			t.Errorf("%s: reason = %q, want protection", tc.name, r.Reason)
		}
	}
}

func TestManaValueProtectionIsPreventingDamageAndCountsX(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushMVCreature(g, opp, "Void Shields", "{7}", proMV3)

	push := func(cost string, x int) uuid.UUID {
		c := NewCard("Spell "+cost, me.ID)
		c.TypeLine = "Instant"
		c.ManaCost = cost
		g.WithWriteLock(func() {
			g.Stack.PushTop(c)
			if g.StackMeta == nil {
				g.StackMeta = make(map[uuid.UUID]*StackItem)
			}
			g.StackMeta[c.InstanceID] = &StackItem{
				ID: c.InstanceID, Kind: StackItemSpell, Controller: me.ID, Owner: me.ID,
				SourceCardID: c.InstanceID, XValue: x, Seq: g.nextStackSeqLocked(),
			}
		})
		return c.InstanceID
	}
	deal := func(src uuid.UUID, n int) int {
		before := damageOn(g, victim)
		g.WithWriteLock(func() {
			if err := g.DealDamageToCreatureForEffect(src, victim, n); err != nil {
				t.Fatalf("damage: %v", err)
			}
		})
		return damageOn(g, victim) - before
	}
	if got := deal(push("{R}", 0), 2); got != 0 {
		t.Errorf("a mana value 1 spell dealt %d, want 0", got)
	}
	if got := deal(push("{3}", 0), 2); got != 0 {
		t.Errorf("a mana value 3 spell dealt %d, want 0", got)
	}
	if got := deal(push("{3}{R}", 0), 2); got != 2 {
		t.Errorf("a mana value 4 spell dealt %d, want 2", got)
	}
	// CR 202.3e: {X}{R} cast with X=2 is mana value 3 on the stack; X=3 is 4.
	if got := deal(push("{X}{R}", 2), 2); got != 0 {
		t.Errorf("{X}{R} with X=2 (mana value 3) dealt %d, want 0", got)
	}
	if got := deal(push("{X}{R}", 3), 2); got != 2 {
		t.Errorf("{X}{R} with X=3 (mana value 4) dealt %d, want 2", got)
	}
	// Off the stack, X is 0: a permanent {X}{R} is mana value 1.
	perm := pushMVCreature(g, me, "Permanent X", "{X}{R}")
	if got := deal(perm, 2); got != 0 {
		t.Errorf("a permanent with {X}{R} is mana value 1; dealt %d, want 0", got)
	}
}

func TestManaValueProtectionRefusesTargetingAndCountsXAtAnnounce(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	const oracle = "test-mv-protection-targeting"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return anyCreatureSpec()
		}
		return nil
	})
	victim := pushMVCreature(g, opp, "Void Shields", "{7}", proMV3)
	ref := TargetRef{Kind: TargetCard, ID: victim}

	cast := func(cost string, x int) error {
		spell := NewCard("Test Spell", me.ID)
		spell.TypeLine = "Instant"
		spell.OracleID = oracle
		spell.ManaCost = cost
		me.Hand.PushTop(spell)
		return g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{Targets: []TargetRef{ref}, XValue: x})
	}
	if err := cast("{1}{R}", 0); err != ErrIllegalTarget {
		t.Errorf("mana value 2 spell: got %v, want ErrIllegalTarget", err)
	}
	if err := cast("{X}{R}", 2); err != ErrIllegalTarget {
		t.Errorf("{X}{R} announced with X=2 is mana value 3: got %v, want ErrIllegalTarget", err)
	}
	if err := cast("{X}{R}", 3); err != nil {
		t.Errorf("{X}{R} announced with X=3 is mana value 4, legal: %v", err)
	}
	if err := cast("{3}{R}", 0); err != nil {
		t.Errorf("mana value 4 spell is legal: %v", err)
	}

	// An ability's source is the permanent: its mana value, X as 0.
	cheap := pushMVCreature(g, me, "Cheap Pinger", "{1}")
	dear := pushMVCreature(g, me, "Dear Pinger", "{5}")
	var cheapOK, dearOK bool
	g.ReadSnapshot(func() {
		spec := anyCreatureSpec()
		cheapOK = g.targetLegalLocked(SourceObject(me.ID, &g.Battlefield.Cards[findCardOnBattlefield(g, cheap)]), spec, ref)
		dearOK = g.targetLegalLocked(SourceObject(me.ID, &g.Battlefield.Cards[findCardOnBattlefield(g, dear)]), spec, ref)
	})
	if cheapOK || !dearOK {
		t.Errorf("ability sources: cheap legal = %v (want false), dear legal = %v (want true)", cheapOK, dearOK)
	}
}

func TestManaValueProtectionDropsACheapAttachment(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	host := pushMVCreature(g, me, "Void Shields", "{7}", proMV3)
	cheap := pushAttachTestCard(g, me.ID, "Cheap Sword", "Artifact — Equipment")
	dear := pushAttachTestCard(g, me.ID, "Dear Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		g.Battlefield.Cards[findCardOnBattlefield(g, cheap)].ManaCost = "{2}"
		g.Battlefield.Cards[findCardOnBattlefield(g, dear)].ManaCost = "{4}"
		g.recomputeLayersLocked()
		for _, id := range []uuid.UUID{cheap, dear} {
			if err := g.AttachForEffect(id, TargetRef{Kind: TargetCard, ID: host}); err != nil {
				t.Fatalf("attach: %v", err)
			}
		}
		g.runStateChecksLocked()
	})
	if c, ok := battlefieldCardByID(g, cheap); !ok || c.IsAttached() {
		t.Error("a mana value 2 Equipment must come off a protection from mana value 3 or less host")
	}
	if c, ok := battlefieldCardByID(g, dear); !ok || !c.IsAttachedTo(host) {
		t.Error("a mana value 4 Equipment stays attached")
	}
}

func TestDepartedSourceKeepsItsManaValueAndRingBearer(t *testing.T) {
	ch := lastKnownSourceCharacteristics(PermanentInfo{ManaValue: 2, RingBearer: true, Controller: uuid.New()})
	if !ch.SourceManaValueKnown || ch.SourceManaValue != 2 || !ch.SourceRingBearer {
		t.Errorf("departed source facts = %+v", ch)
	}
}

// --- Ring-bearers (#2145) -------------------------------------------

func TestRingBearerProtectionAcrossTheFourChecks(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	const oracle = "test-ring-bearer-targeting"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return anyCreatureSpec()
		}
		return nil
	})

	wraith := pushMVCreature(g, opp, "Wraith", "{1}", "protection from Ring-bearers")
	ref := TargetRef{Kind: TargetCard, ID: wraith}
	bearer := pushMVCreature(g, me, "Bearer", "{1}")
	plain := pushMVCreature(g, me, "Plain", "{1}")
	card := func(id uuid.UUID) *Card { return &g.Battlefield.Cards[findCardOnBattlefield(g, id)] }
	g.WithWriteLock(func() { card(bearer).RingBearer = true })

	// Block (CR 702.16f): the Wraith attacks, the bearer is refused.
	var rBearer, rPlain BlockRefusal
	g.ReadSnapshot(func() {
		rBearer = g.BlockPairRefusalLocked(card(wraith), card(bearer))
		rPlain = g.BlockPairRefusalLocked(card(wraith), card(plain))
	})
	if rBearer.Reason != BlockReasonProtection {
		t.Errorf("a Ring-bearer blocking the Wraith: reason = %q, want protection", rBearer.Reason)
	}
	if !rPlain.Legal() {
		t.Errorf("a non-bearer blocks fine: %q", rPlain.Reason)
	}

	// Targeting (CR 702.16b), by the bearer's ability.
	var bearerOK, plainOK bool
	g.ReadSnapshot(func() {
		spec := anyCreatureSpec()
		bearerOK = g.targetLegalLocked(SourceObject(me.ID, card(bearer)), spec, ref)
		plainOK = g.targetLegalLocked(SourceObject(me.ID, card(plain)), spec, ref)
	})
	if bearerOK || !plainOK {
		t.Errorf("targeting: bearer legal = %v (want false), plain legal = %v (want true)", bearerOK, plainOK)
	}

	// Damage (CR 702.16e).
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(bearer, wraith, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(plain, wraith, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, wraith); got != 1 {
		t.Errorf("Wraith has %d damage, want 1 (the Ring-bearer's prevented)", got)
	}

	// Attachment (CR 702.16c-d): an Equipment that is a Ring-bearer
	// (any permanent can be) comes off; one that is not stays.
	sword := pushAttachTestCard(g, me.ID, "Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		card(sword).RingBearer = true
		if err := g.AttachForEffect(sword, ref); err != nil {
			t.Fatal(err)
		}
		g.runStateChecksLocked()
	})
	if c, ok := battlefieldCardByID(g, sword); !ok || c.IsAttached() {
		t.Error("an Equipment that is a Ring-bearer must come off")
	}

	// The designation is the CONTROLLER's: a flag on a permanent its
	// controller does not control matches nothing (CR 701.54e).
	if IsRingBearerOf(*card(bearer), opp.ID) {
		t.Error("IsRingBearerOf must read the controller")
	}
}
