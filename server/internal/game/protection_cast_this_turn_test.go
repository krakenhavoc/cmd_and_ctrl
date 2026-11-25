package game

import (
	"testing"
)

// protection_cast_this_turn_test.go — #6098, Emrakul, the World Anew:
// "protection from spells and from permanents that were cast this turn"
// (CR 702.16a, two protection abilities by CR 702.16g). A spell is any spell; a
// permanent matches only if its spell was cast THIS turn
// (CastProvenance.CastThisTurn), which ends as the next turn begins.

func TestProtectionFromSpellsAndPermanentsCastThisTurnGrammar(t *testing.T) {
	for _, tc := range []struct {
		token string
		kind  ProtectionQualityKind
		wire  string
	}{
		{"protection from spells", ProtectionQualitySpells, "spells"},
		{"protection from permanents that were cast this turn", ProtectionQualityPermanentsCastThisTurn, "permanents_cast_this_turn"},
	} {
		q, ok := ParseProtectionQuality(tc.token)
		if !ok || q.Kind != tc.kind || q.Kind.String() != tc.wire {
			t.Errorf("%q = (%+v, %v)", tc.token, q, ok)
		}
		if _, ok := CanonicalKeywords(tc.token); !ok {
			t.Errorf("CanonicalKeywords refused %q", tc.token)
		}
	}
}

func TestProtectionFromSpellsAndPermanentsCastThisTurnMatch(t *testing.T) {
	spells := ProtectionQuality{Kind: ProtectionQualitySpells}
	cast := ProtectionQuality{Kind: ProtectionQualityPermanentsCastThisTurn}
	for _, tc := range []struct {
		name               string
		src                *Characteristic
		wantSpell, wantPer bool
	}{
		{"colorless spell", &Characteristic{SourceIsSpell: true}, true, false},
		{"red spell", &Characteristic{Colors: []string{"R"}, SourceIsSpell: true}, true, false},
		{"permanent cast this turn", &Characteristic{SourceIsPermanent: true, SourceCastThisTurn: true}, false, true},
		{"permanent cast earlier, or put onto the battlefield", &Characteristic{SourceIsPermanent: true}, false, false},
		{"an ability's source off the battlefield", &Characteristic{}, false, false},
		{"unknown source", nil, false, false},
	} {
		if got := spells.Matches(tc.src); got != tc.wantSpell {
			t.Errorf("%s: spells Matches = %v, want %v", tc.name, got, tc.wantSpell)
		}
		if got := cast.Matches(tc.src); got != tc.wantPer {
			t.Errorf("%s: permanents cast this turn Matches = %v, want %v", tc.name, got, tc.wantPer)
		}
	}
}

// A creature cast for real carries the fact as a permanent and as a
// departed one, a creature put onto the battlefield does not, and the
// next turn ends it on both.
func TestCastThisTurnIsStampedOnResolutionAndEndsWithTheTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceIntoStep(t, g, StepPrecombatMain)
	spell := NewCard("Cast Bear", me.ID)
	spell.TypeLine = "Creature — Bear"
	spell.Power, spell.Toughness = 2, 2
	me.Hand.PushTop(spell)
	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	put := pushCombatant(t, g, me, "Put Bear", 2, 2)
	dying := pushCombatant(t, g, me, "Dying Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == dying {
				g.Battlefield.Cards[i].Provenance.CastThisTurn = true
			}
		}
	})

	if !g.CastProvenanceForEffect(spell.InstanceID).CastThisTurn {
		t.Fatal("the cast creature's provenance does not say it was cast this turn")
	}
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.battlefieldCardLocked(spell.InstanceID)
		if ch := SourceCharacteristics(c); !ch.SourceCastThisTurn {
			t.Error("the cast creature's source snapshot does not say cast this turn")
		}
		p, _ := g.battlefieldCardLocked(put)
		if ch := SourceCharacteristics(p); ch.SourceCastThisTurn {
			t.Error("a creature put onto the battlefield reads as cast this turn")
		}
		if err := g.DestroyPermanentForEffect(dying); err != nil {
			t.Fatal(err)
		}
	})
	rec, ok := g.LastKnownPermanentForEffect(dying)
	if !ok || !lastKnownSourceCharacteristics(rec).SourceCastThisTurn {
		t.Error("a departed creature cast this turn does not read so as a last-known source")
	}

	advanceOneTurn(t, g)
	if g.CastProvenanceForEffect(spell.InstanceID).CastThisTurn {
		t.Error("the next turn began, but the creature still reads as cast this turn")
	}
	if !g.CastProvenanceForEffect(spell.InstanceID).Any() {
		t.Error("the rest of the provenance went with the turn")
	}
	if rec, _ := g.LastKnownPermanentForEffect(dying); rec.CastThisTurn {
		t.Error("the next turn began, but the last-known record still reads as cast this turn")
	}
}
