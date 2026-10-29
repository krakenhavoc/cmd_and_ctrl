package game

import (
	"testing"

	"github.com/google/uuid"
)

// protection_sticker_qualities_test.go — the four qualities the ability
// stickers of the Unfinity sticker sheets print (#4307): "protection from
// odd mana values", "from even mana values", "from creatures with two or
// more creature types" and "from noncreature permanents". Each is asserted
// through the checks that read a source (CR 702.16b-f), because the work is
// that every reader sees the new source facts.

func TestStickerProtectionGrammar(t *testing.T) {
	for _, tc := range []struct {
		token string
		kind  ProtectionQualityKind
		wire  string
	}{
		{"protection from odd mana values", ProtectionQualityManaValueOdd, "mana_value_odd"},
		{"protection from even mana values", ProtectionQualityManaValueEven, "mana_value_even"},
		{"protection from creatures with two or more creature types", ProtectionQualityMultitypedCreatures, "multityped_creatures"},
		{"protection from noncreature permanents", ProtectionQualityNoncreaturePermanents, "noncreature_permanents"},
		{"Protection from Odd Mana Values", ProtectionQualityManaValueOdd, "mana_value_odd"},
	} {
		q, ok := ParseProtectionQuality(tc.token)
		if !ok || q.Kind != tc.kind || q.Kind.String() != tc.wire {
			t.Errorf("%q = (%+v, %v), want kind %v wire %q", tc.token, q, ok, tc.kind, tc.wire)
		}
		if toks, ok := ProtectionTokens(tc.token); !ok || len(toks) != 1 {
			t.Errorf("ProtectionTokens(%q) = %v, %v", tc.token, toks, ok)
		}
		if _, ok := CanonicalKeywords(tc.token); !ok {
			t.Errorf("CanonicalKeywords(%q) refused it", tc.token)
		}
	}
	// The grammar stays closed.
	for _, bad := range []string{
		"protection from odd mana value",
		"protection from mana values",
		"protection from creatures with two creature types",
		"protection from creatures with three or more creature types",
		"protection from permanents",
		"protection from noncreature permanent",
	} {
		if _, ok := ParseProtectionQuality(bad); ok {
			t.Errorf("%q parsed; the grammar is closed", bad)
		}
	}
}

func TestStickerProtectionParityRefusesBlocksByManaValue(t *testing.T) {
	for _, tc := range []struct {
		quality string
		cost    string
		token   bool
		legal   bool
	}{
		{"protection from odd mana values", "{G}", false, false},
		{"protection from odd mana values", "{1}{G}", false, true},
		{"protection from odd mana values", "{2}{G}", false, false},
		{"protection from odd mana values", "", true, true}, // mana value 0 is even
		{"protection from even mana values", "{1}{G}", false, false},
		{"protection from even mana values", "{G}", false, true},
		{"protection from even mana values", "", true, false},
	} {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		attacker := pushMVCreature(g, me, "Shielded", "{7}", tc.quality)
		blocker := pushMVCreature(g, opp, "Blocker", tc.cost)
		card := func(id uuid.UUID) *Card { return &g.Battlefield.Cards[findCardOnBattlefield(g, id)] }
		if tc.token {
			g.WithWriteLock(func() { card(blocker).TypeLine = "Token Creature — Test" })
		}
		var r BlockRefusal
		g.ReadSnapshot(func() { r = g.BlockPairRefusalLocked(card(attacker), card(blocker)) })
		if tc.legal && !r.Legal() {
			t.Errorf("%s vs %q token=%v: refused (%q), want legal", tc.quality, tc.cost, tc.token, r.Reason)
		}
		if !tc.legal && r.Reason != BlockReasonProtection {
			t.Errorf("%s vs %q token=%v: reason %q, want protection", tc.quality, tc.cost, tc.token, r.Reason)
		}
	}
}

func TestStickerProtectionMultitypedCountsCreatureTypes(t *testing.T) {
	const quality = "protection from creatures with two or more creature types"
	for _, tc := range []struct {
		name     string
		typeLine string
		keywords []string
		legal    bool
	}{
		{"one creature type", "Creature — Elf", nil, true},
		{"two creature types", "Creature — Elf Warrior", nil, false},
		{"three creature types", "Creature — Human Cleric Soldier", nil, false},
		{"an artifact subtype is not a creature type", "Artifact Creature — Equipment Elf", nil, true},
		{"an artifact creature with two creature types", "Artifact Creature — Construct Elf", nil, false},
		{"a changeling is every creature type", "Creature — Shapeshifter", []string{"changeling"}, false},
	} {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		attacker := pushMVCreature(g, me, "Shielded", "{7}", quality)
		blocker := pushMVCreature(g, opp, tc.name, "{1}", tc.keywords...)
		card := func(id uuid.UUID) *Card { return &g.Battlefield.Cards[findCardOnBattlefield(g, id)] }
		g.WithWriteLock(func() {
			card(blocker).TypeLine = tc.typeLine
			g.recomputeLayersLocked()
		})
		var r BlockRefusal
		g.ReadSnapshot(func() { r = g.BlockPairRefusalLocked(card(attacker), card(blocker)) })
		if tc.legal && !r.Legal() {
			t.Errorf("%s: refused (%q), want legal", tc.name, r.Reason)
		}
		if !tc.legal && r.Reason != BlockReasonProtection {
			t.Errorf("%s: reason %q, want protection", tc.name, r.Reason)
		}
	}
}

func TestStickerProtectionNoncreaturePermanentsAreRealPermanents(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushMVCreature(g, opp, "Shielded", "{7}", "protection from noncreature permanents")

	rock := pushMVCreature(g, me, "Rock", "{2}")
	bear := pushMVCreature(g, me, "Bear", "{2}")
	card := func(id uuid.UUID) *Card { return &g.Battlefield.Cards[findCardOnBattlefield(g, id)] }
	g.WithWriteLock(func() {
		card(rock).TypeLine = "Artifact"
		g.recomputeLayersLocked()
	})
	deal := func(src uuid.UUID) int {
		before := damageOn(g, victim)
		g.WithWriteLock(func() {
			if err := g.DealDamageToCreatureForEffect(src, victim, 2); err != nil {
				t.Fatalf("damage: %v", err)
			}
		})
		return damageOn(g, victim) - before
	}
	if got := deal(rock); got != 0 {
		t.Errorf("a noncreature permanent dealt %d, want 0", got)
	}
	if got := deal(bear); got != 2 {
		t.Errorf("a creature dealt %d, want 2", got)
	}
	// An artifact SPELL is not a permanent.
	spell := NewCard("Artifact Spell", me.ID)
	spell.TypeLine = "Artifact"
	spell.ManaCost = "{2}"
	g.WithWriteLock(func() {
		g.Stack.PushTop(spell)
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*StackItem)
		}
		g.StackMeta[spell.InstanceID] = &StackItem{
			ID: spell.InstanceID, Kind: StackItemSpell, Controller: me.ID, Owner: me.ID,
			SourceCardID: spell.InstanceID, Seq: g.nextStackSeqLocked(),
		}
	})
	if got := deal(spell.InstanceID); got != 2 {
		t.Errorf("an artifact spell dealt %d, want 2: it is not a permanent", got)
	}
}

func TestStickerProtectionDepartedSourceStaysAPermanent(t *testing.T) {
	ch := lastKnownSourceCharacteristics(PermanentInfo{ManaValue: 2, Controller: uuid.New()})
	if !ch.SourceIsPermanent {
		t.Error("a last-known source was a permanent on the battlefield")
	}
}

func TestStickerProtectionMatchesByItsOwnFacts(t *testing.T) {
	odd, _ := ParseProtectionQuality("protection from odd mana values")
	unknown := &Characteristic{Types: []string{"Creature"}}
	if odd.Matches(unknown) {
		t.Error("a source whose mana value is unknown was matched: it errs weaker")
	}
	np, _ := ParseProtectionQuality("protection from noncreature permanents")
	if np.Matches(&Characteristic{Types: []string{"Artifact"}}) {
		t.Error("a source that is no permanent was matched")
	}
	if !np.Matches(&Characteristic{Types: []string{"Artifact"}, SourceIsPermanent: true}) {
		t.Error("a noncreature permanent was not matched")
	}
	if np.Matches(&Characteristic{Types: []string{"Artifact", "Creature"}, SourceIsPermanent: true}) {
		t.Error("an artifact creature is a creature")
	}
}
